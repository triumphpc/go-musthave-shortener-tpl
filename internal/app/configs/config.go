package configs

import (
	"errors"
	"flag"
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

var instance *Config

// Instance new Config
func Instance() *Config {
	if instance == nil {
		instance = new(Config)
		instance.init()
	}
	return instance
}

// Param from configs
func (c *Config) Param(p string) (string, error) {
	switch p {
	case BaseURL:
		if c.baseURL != "" {
			return c.baseURL, nil
		}
		c.baseURL = fromInv(p)
		return c.baseURL, nil
	case ServerAddress:
		if c.serverAddress != "" {
			return c.serverAddress, nil
		}
		c.serverAddress = fromInv(p)
		return c.serverAddress, nil
	case FileStoragePath:
		if c.fileStoragePath != "" {
			return c.fileStoragePath, nil
		}
		c.fileStoragePath = fromInv(p)
		return c.fileStoragePath, nil
	}
	return "", ErrUnknownParam
}

// fromInv check from inv
func fromInv(p string) string {
	// Get from inv
	param := os.Getenv(p)
	if param != "" {
		return param
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
	return ""
}

// initParams from cli params
func (c *Config) init() {
	bu := flag.String(mapVarToInv[BaseURL], "", "")
	sa := flag.String(mapVarToInv[ServerAddress], "", "")
	fs := flag.String(mapVarToInv[FileStoragePath], "", "")
	flag.Parse()

	c.baseURL = *bu
	c.serverAddress = *sa
	c.fileStoragePath = *fs
}
