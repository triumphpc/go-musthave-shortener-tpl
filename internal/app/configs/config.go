package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
)

// Config project
type Config struct {
	ServerAddress string `env:"APP_BASE_HOST" envDefault:""`
	BaseURL       string `env:"APP_BASE_URL" envDefault:"http://localhost"`
	Port          string `env:"APP_PORT" envDefault:"8080"`
}

// New Instance new Config
func New() Config {
	var cfg Config

	// Parse environment
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}
