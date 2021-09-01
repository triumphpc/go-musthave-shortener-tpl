package configs

import (
	"github.com/caarlos0/env"
	"log"
)

// Config project
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS,required"`
	BaseURL       string `env:"BASE_URL,required"`
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
