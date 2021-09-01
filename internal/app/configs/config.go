package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

// Config project
type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	BaseHost      string `env:"BASE_HOST" envDefault:":8080"`
	//Port          string `env:"APP_PORT" envDefault:"8080"`
}

// New Instance new Config
func New() Config {
	var cfg Config

	// Parse environment
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("ENVIRONMENTS:")
	log.Println(os.Environ())
	log.Println(cfg)

	return cfg
}
