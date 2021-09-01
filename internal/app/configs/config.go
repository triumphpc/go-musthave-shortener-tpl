package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

// CustomPort for server
const CustomPort = "9080"

// Config project
type Config struct {
	ServerHost string `env:"BASE_URL" envDefault:""`
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
	if cfg.ServerHost != "" {
		cfg.ServerPort = CustomPort
	}

	log.Println("ENVIRONMENTS:")
	log.Println(os.Environ())
	log.Println(cfg)

	return cfg
}
