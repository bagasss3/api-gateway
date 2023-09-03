package console

import (
	"api-gateway/src/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "test",
	Short: "Example Cobra",
	Long:  "Example of using CLI created by Cobra",
}

func init() {
	config.InitConfig()
}

// Execute :nodoc:
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
