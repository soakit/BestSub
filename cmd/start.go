package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"bestsub/internal/conf"
	"bestsub/internal/mihomo"
	"bestsub/internal/model"
	_ "bestsub/internal/server/handlers"
	"bestsub/internal/server/middleware"
	"bestsub/internal/server/router"
	"bestsub/internal/store"
	"bestsub/internal/utils"

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

		r.Use(middleware.Cors())
		r.Use(middleware.Logger())
		r.Use(middleware.StaticLocal("/", "static"))

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

		// 从数据库加载 DNS 配置并生效
		defStr := store.SettingGet(model.SettingDNSDefault)
		mainStr := store.SettingGet(model.SettingDNSMain)
		if defStr != "" || mainStr != "" {
			mihomo.UpdateDNSConfig(utils.SplitComma(defStr), utils.SplitComma(mainStr))
		}

		addr := fmt.Sprintf("%s:%d", conf.AppConfig.Server.Host, conf.AppConfig.Server.Port)
		log.Infof("http server listening on http://%s", addr)
		httpSrv := &http.Server{Addr: addr, Handler: r}

		go func() {
			if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Errorf("http server listen and serve error: %v", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		// 等 Ctrl+C，让命令正常返回，避免 Windows 默认中断退出码 0xc000013a。
		<-quit
	},
}

func init() {
	startCmd.Flags().StringVar(&startConfig, "config", "", "config file (default is ./data/config.json)")
	rootCmd.AddCommand(startCmd)
}
