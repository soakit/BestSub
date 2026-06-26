package cmd

import (
	"bestsub/internal/conf"
	"bestsub/internal/store"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var (
	userConfig      string
	userCmdUsername string
	userCmdPassword string
)

var userCmd = &cobra.Command{
	Use:   "user --username <username> --password <password>",
	Short: "Set username and password",
	PreRun: func(cmd *cobra.Command, args []string) {
		log.SetReportCaller(true)
		if err := conf.Load(userConfig); err != nil {
			log.Fatalf("load config failed: %v", err)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if userCmdUsername == "" || userCmdPassword == "" {
			log.Fatal("username and password are required")
		}
		if err := store.InitDB(); err != nil {
			log.Fatalf("database init error: %v", err)
		}
		defer func() {
			if err := store.Close(); err != nil {
				log.Errorf("database close error: %v", err)
			}
		}()
		if err := store.UserSet(userCmdUsername, userCmdPassword); err != nil {
			log.Fatalf("user update error: %v", err)
		}
		log.Info("user updated successfully")
	},
}

func init() {
	userCmd.Flags().StringVar(&userConfig, "config", "", "config file (default is ./data/config.json)")
	userCmd.Flags().StringVar(&userCmdUsername, "username", "", "new username")
	userCmd.Flags().StringVar(&userCmdPassword, "password", "", "new password")
	rootCmd.AddCommand(userCmd)
}
