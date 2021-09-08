package configs

import (
	"errors"
	"flag"
	"log"
	"os"
)

var ErrUnknownParam = errors.New("unknown param")

// Config project
type Config struct {
	baseURL         string
	fileStoragePath string
	serverAddress   string
}

const (
	BaseURL                = "BASE_URL"
	BaseURLDefault         = "http://localhost:8080"
	ServerAddress          = "SERVER_ADDRESS"
	ServerAddressDefault   = ":8080"
	FileStoragePath        = "FILE_STORAGE_PATH"
	FileStoragePathDefault = "unknown"
)

// Maps for take inv params
var mapVarToInv = map[string]string{
	BaseURL:         "b",
	ServerAddress:   "a",
	FileStoragePath: "f",
}

var instance = Config{}

// Instance new Config
func Instance() *Config {
	return &instance
}

// Param from configs
func (c *Config) Param(p string) (string, error) {
	switch p {
	case BaseURL:
		if c.baseURL != "" {
			return c.baseURL, nil
		}
		c.baseURL = initParam(p)
		return c.baseURL, nil
	case ServerAddress:
		if c.serverAddress != "" {
			log.Println("SERVER ADD: " + c.serverAddress)
			return c.serverAddress, nil
		}
		c.serverAddress = initParam(p)
		log.Println("SERVER ADD 2: " + c.serverAddress)
		return c.serverAddress, nil
	case FileStoragePath:
		if c.fileStoragePath != "" {
			return c.fileStoragePath, nil
		}
		c.fileStoragePath = initParam(p)
		return c.fileStoragePath, nil
	}
	return "", ErrUnknownParam
}

func initParam(p string) string {
	param := flag.String(mapVarToInv[p], "", "")
	flag.Parse()

	if *param != "" {
		return *param
	} else {
		// Get from inv
		invPar := os.Getenv(p)
		if invPar != "" {
			return invPar
		} else {
			// return default value
			switch p {
			case BaseURL:
				return BaseURLDefault
			case ServerAddress:
				return ServerAddressDefault
			case FileStoragePath:
				return FileStoragePathDefault
			}
		}
	}
	return ""
}
