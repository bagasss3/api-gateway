package cmd

import (
	"api-gateway/internal/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "gateway",
	Short: "API Gateway",
	Long:  "API Gateway",
}

func init() {
	config.InitConfig()
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
