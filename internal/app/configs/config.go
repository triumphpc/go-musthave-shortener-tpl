package configs

import (
	"errors"
	"flag"
	"github.com/caarlos0/env"
	_ "github.com/caarlos0/env/v6"
)

var ErrUnknownParam = errors.New("unknown param")

// Config project
type Config struct {
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"unknown"`
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	DatabaseDsn     string `env:"DATABASE_DSN" envDefault:""`
}

const (
	BaseURL                = "BASE_URL"
	ServerAddress          = "SERVER_ADDRESS"
	FileStoragePath        = "FILE_STORAGE_PATH"
	DatabaseDsn            = "DATABASE_DSN"
	FileStoragePathDefault = "unknown"
)

// Maps for take inv params
var mapVarToInv = map[string]string{
	BaseURL:         "b",
	ServerAddress:   "a",
	FileStoragePath: "f",
	DatabaseDsn:     "d",
}

var instance *Config

// Instance new Config
func Instance() *Config {
	if instance == nil {
		instance = new(Config)
		instance.initInv()
		instance.init()
	}
	return instance
}

// Param from configs
func (c *Config) Param(p string) (string, error) {
	switch p {
	case BaseURL:
		return c.BaseURL, nil
	case ServerAddress:
		return c.ServerAddress, nil
	case FileStoragePath:
		return c.FileStoragePath, nil
	case DatabaseDsn:
		return c.DatabaseDsn, nil
	}
	return "", ErrUnknownParam
}

// initInv check from inv
func (c *Config) initInv() {
	// Get from inv
	if err := env.Parse(c); err != nil {
		return
	}
}

// initParams from cli params
func (c *Config) init() {
	bu := flag.String(mapVarToInv[BaseURL], "", "")
	sa := flag.String(mapVarToInv[ServerAddress], "", "")
	fs := flag.String(mapVarToInv[FileStoragePath], "", "")
	db := flag.String(mapVarToInv[DatabaseDsn], "", "")
	flag.Parse()

	if *bu != "" {
		c.BaseURL = *bu
	}
	if *sa != "" {
		c.ServerAddress = *sa
	}
	if *fs != "" {
		c.FileStoragePath = *fs
	}
	if *db != "" {
		c.DatabaseDsn = *db
	}
}
