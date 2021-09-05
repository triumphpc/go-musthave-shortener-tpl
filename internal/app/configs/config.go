package configs

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

var (
	baseURL         *string
	serverAddress   *string
	fileStoragePath *string
)

// Config project
type Config struct {
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
}

func init() {
	serverAddress = flag.String("a", "", "server address")
	baseURL = flag.String("b", "", "base url")
	fileStoragePath = flag.String("f", "", "a string")
}

// New Instance new Config
func New() Config {
	var c Config
	// Parse environment
	err := env.Parse(&c)
	if err != nil {
		panic(err)
	}

	// parse args
	flag.Parse()
	if *serverAddress != "" {
		c.ServerAddress = *serverAddress
	}
	if *baseURL != "" {
		c.BaseURL = *baseURL
	}

	if *fileStoragePath != "" {
		c.FileStoragePath = *fileStoragePath
	}
	return c
}
