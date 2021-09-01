package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

// Config project
type Config struct {
	ServerAddress string `env:"BASE_URL" envDefault:""`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	ServerPort    string `env:"PORT" envDefault:"8080"`
}

// New Instance new Config
func New() Config {
	var cfg Config

	// Parse environment
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.BaseURL != "http://localhost" {
		cfg.ServerPort = "9080"
	}

	log.Println("ENVIRONMENTS:")
	log.Println(os.Environ())
	log.Println(cfg)

	return cfg
}
