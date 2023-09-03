package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func InitConfig() {
	viper.AddConfigPath(".")
	viper.AddConfigPath("./../..")
	viper.SetConfigName("config")

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Warningf("%v", err)
	}
	log.Info("Using config file: ", viper.ConfigFileUsed())
}

func Port() string {
	if !viper.IsSet("port") {
		return "8080"
	}
	return viper.GetString("port")
}

func CakeService() string {
	return viper.GetString("service.cake-service")
}

func Servis2() string {
	return viper.GetString("service.servis2")
}
