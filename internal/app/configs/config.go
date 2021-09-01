package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
)

// Config project
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:""`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost"`
	Port          string `env:"PORT" envDefault:"8080"`
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
