package configs

import (
	"github.com/caarlos0/env/v6"
	"log"
	"os"
)

// DefaultPort for server
const DefaultPort = "8080"

// CustomPort for server
const CustomPort = "9080"
const DefaultHost = "http://localhost"

// Config project
type Config struct {
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	ServerPort      string
	FileStoragePath string `env:"PERSISTENT_FILE_PATH" envDefault:""`
}

// New Instance new Config
func New() Config {
	var c Config
	// Parse environment
	err := env.Parse(&c)
	if err != nil {
		panic(err)
	}

	// Set server port
	if c.BaseURL != DefaultHost+":"+DefaultPort {
		c.ServerPort = CustomPort
	} else {
		// if set base url on server
		c.ServerPort = DefaultPort
	}

	log.Println("ENVIRONMENTS:")
	log.Println(os.Environ())
	log.Println(c)
	return c
}
