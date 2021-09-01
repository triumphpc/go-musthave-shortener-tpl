package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
)

// CustomPort for server
const CustomPort = "9080"
const DefaultHost = "http://localhost"

// Config project
type Config struct {
	BaseURL    string `env:"BASE_URL" envDefault:"http://localhost"`
	ServerPort string `env:"PORT" envDefault:"8080"`
}

// New Instance new Config
func New() Config {
	var cfg Config

	// Parse environment
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	// if is set base url on server
	if cfg.BaseURL != DefaultHost {
		cfg.ServerPort = CustomPort
	}

	return cfg
}
