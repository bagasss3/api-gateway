package config

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type RouteConfig struct {
	Name           string   `mapstructure:"name"`
	Prefix         string   `mapstructure:"prefix"`
	StripPrefix    string   `mapstructure:"strip_prefix"`
	Auth           string   `mapstructure:"auth"` // "none" | "api_key" | "jwt"
	RequireAPIKey  bool     `mapstructure:"require_api_key"`
	PublicPrefixes []string `mapstructure:"public_prefixes"`
	Targets        []string `mapstructure:"targets"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
	Credentials    bool     `mapstructure:"credentials"`
}

type LimitsConfig struct {
	GlobalRPS   int `mapstructure:"global_rps"`
	PerIPRPS    int `mapstructure:"per_ip_rps"`
	PerIPBurst  int `mapstructure:"per_ip_burst"`
	Concurrency int `mapstructure:"concurrency"`
	MaxBodyMB   int `mapstructure:"max_body_mb"`
}

type AuthConfig struct {
	IntrospectURL string `mapstructure:"introspect_url"`
}

type Config struct {
	Port    string        `mapstructure:"port"`
	Auth    AuthConfig    `mapstructure:"auth"`
	CORS    CORSConfig    `mapstructure:"cors"`
	Limits  LimitsConfig  `mapstructure:"limits"`
	APIKeys []string      `mapstructure:"api_keys"`
	Routes  []RouteConfig `mapstructure:"routes"`
}

var cfg Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./../..") // keep your original search path
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Warningf("config: %v", err)
	} else {
		log.Info("Using config file: ", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("config unmarshal: %v", err)
	}
	// sensible defaults
	if cfg.Port == "" {
		cfg.Port = ":8080"
	}
	if len(cfg.CORS.AllowedMethods) == 0 {
		cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}
	if len(cfg.CORS.AllowedHeaders) == 0 {
		cfg.CORS.AllowedHeaders = []string{"Authorization", "Content-Type", "X-API-Key"}
	}
	if cfg.Limits.GlobalRPS == 0 {
		cfg.Limits.GlobalRPS = 200
	}
	if cfg.Limits.PerIPRPS == 0 {
		cfg.Limits.PerIPRPS = 20
	}
	if cfg.Limits.PerIPBurst == 0 {
		cfg.Limits.PerIPBurst = 40
	}
	if cfg.Limits.Concurrency == 0 {
		cfg.Limits.Concurrency = 500
	}
	if cfg.Limits.MaxBodyMB == 0 {
		cfg.Limits.MaxBodyMB = 10
	}
}

func Get() Config { return cfg }
