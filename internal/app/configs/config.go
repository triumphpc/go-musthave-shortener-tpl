package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

// DefaultHost for server
const DefaultHost = "http://localhost"

// DefaultPort for server
const DefaultPort = "8080"

// Config project
type Config struct {
	ServerHost string `env:"BASE_URL" envDefault:"http://localhost"`
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

	if cfg.ServerHost == DefaultHost {
		cfg.ServerPort = DefaultPort
	}

	log.Println("ENVIRONMENTS:")
	log.Println(os.Environ())
	log.Println(cfg)

	return cfg
}
