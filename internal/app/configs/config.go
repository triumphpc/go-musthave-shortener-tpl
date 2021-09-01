package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
)

// DefaultPort for server
const DefaultPort = "8080"

// CustomPort for server
const CustomPort = "9080"
const DefaultHost = "http://localhost"

// Config project
type Config struct {
	BaseURL    string `env:"BASE_URL" envDefault:"http://localhost"`
	ServerPort string
}

// New Instance new Config
func New() Config {
	var cfg Config

	// Parse environment
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.BaseURL != DefaultHost {
		cfg.ServerPort = CustomPort

	} else {
		// if is set base url on server
		cfg.ServerPort = DefaultPort
	}

	return cfg
}
