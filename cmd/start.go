package cmd

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bestruirui/bestsub/internal/conf"
	"github.com/bestruirui/bestsub/internal/model"
	"github.com/bestruirui/bestsub/internal/node"
	_ "github.com/bestruirui/bestsub/internal/server/handlers"
	"github.com/bestruirui/bestsub/internal/server/middleware"
	"github.com/bestruirui/bestsub/internal/server/router"
	"github.com/bestruirui/bestsub/internal/service"
	"github.com/bestruirui/bestsub/internal/store"
	"github.com/bestruirui/bestsub/internal/task"
	"github.com/bestruirui/bestsub/internal/utils"
	"github.com/bestruirui/bestsub/pkg/mihomo"
	"github.com/bestruirui/bestsub/static"

	"github.com/charmbracelet/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var startConfig string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start " + conf.APP_NAME,
	PreRun: func(cmd *cobra.Command, args []string) {
		conf.PrintBanner()
		log.SetReportCaller(true)
		if err := conf.Load(startConfig); err != nil {
			log.Fatalf("load config failed: %v", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if conf.IsDebug() {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}

		r := gin.New()
		if conf.IsDebug() {
			r.Use(middleware.Logger())
		}
		r.Use(middleware.Cors())
		r.Use(middleware.StaticEmbed("/", static.StaticFS))

		router.RegisterAll(r)

		if err := store.InitDB(); err != nil {
			log.Errorf("database init error: %v", err)
			return
		}
		defer func() {
			if err := store.Close(); err != nil {
				log.Errorf("database close error: %v", err)
			}
		}()

		if err := store.InitStore(); err != nil {
			log.Errorf("user init error: %v", err)
			return
		}
		if err := node.PoolLoad(); err != nil {
			log.Errorf("node pool snapshot load error: %v", err)
		}

		// 从数据库加载 DNS 配置并生效
		defStr := store.SettingGet(model.SettingDNSDefault)
		mainStr := store.SettingGet(model.SettingDNSMain)
		if defStr != "" || mainStr != "" {
			mihomo.UpdateDNSConfig(utils.SplitComma(defStr), utils.SplitComma(mainStr))
		}

		if err := service.StartSubscriptionScheduler(); err != nil {
			log.Errorf("subscription scheduler init error: %v", err)
			return
		}

		if err := task.Start(cmd.Context()); err != nil {
			log.Errorf("task service init error: %v", err)
			service.StopSubscriptionScheduler()
			return
		}

		addr := fmt.Sprintf("%s:%d", conf.AppConfig.Server.Host, conf.AppConfig.Server.Port)
		log.Infof("http server listening on http://%s", addr)
		// SSE 等长连接永远不会变为 idle，Shutdown 只能等超时；
		// 用可取消的 BaseContext 让处理器在关停时主动返回。
		srvCtx, srvCancel := context.WithCancel(context.Background())
		defer srvCancel()
		httpSrv := &http.Server{
			Addr:        addr,
			Handler:     r,
			BaseContext: func(net.Listener) context.Context { return srvCtx },
		}

		go func() {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Errorf("http server listen and serve error: %v", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		// 等 Ctrl+C，让命令正常返回，避免 Windows 默认中断退出码 0xc000013a。
		<-quit
		signal.Stop(quit)

		// 先断开长连接处理器，再关闭 http 服务，避免白等 Shutdown 超时。
		srvCancel()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Errorf("http server shutdown error: %v", err)
		}
		cancel()
		task.Stop()
		service.StopSubscriptionScheduler()
		if err := node.PoolSave(); err != nil {
			log.Errorf("node pool snapshot save error: %v", err)
		}
	},
}

func init() {
	startCmd.Flags().StringVar(&startConfig, "config", "", "config file (default is ./data/config.json)")
	rootCmd.AddCommand(startCmd)
}
