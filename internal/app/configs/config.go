package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
)

// DefaultPort for server
const DefaultPort = "8080"

// Config project
type Config struct {
	BaseURL    string `env:"BASE_URL" envDefault:"http://localhost:8080"`
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
	// if is set base url on server
	cfg.ServerPort = DefaultPort

	return cfg
}
