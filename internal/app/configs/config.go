// Package configs implement functions for environment and project configs
package configs

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/caarlos0/env"
	_ "github.com/caarlos0/env/v6"
	"go.uber.org/zap"

	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	dbh "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/repository"
)

// ErrUnknownParam error for unknown param
var ErrUnknownParam = errors.New("unknown param")

// Config project
type Config struct {
	BaseURL         string `env:"BASE_URL" envDefault:""`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:""`
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:""`
	DatabaseDsn     string `env:"DATABASE_DSN" envDefault:""`
	EnableHTTPS     string `env:"ENABLE_HTTPS" envDefault:""`
	Storage         repository.Repository
	Logger          *zap.Logger
	Database        *sql.DB
}

const (
	BaseURL                = "BASE_URL"
	ServerAddress          = "SERVER_ADDRESS"
	FileStoragePath        = "FILE_STORAGE_PATH"
	DatabaseDsn            = "DATABASE_DSN"
	FileStoragePathDefault = "unknown"
	EnableHTTPS            = "ENABLE_HTTPS"
)

// Maps for take inv params
var mapVarToInv = map[string]string{
	BaseURL:         "b",
	ServerAddress:   "a",
	FileStoragePath: "f",
	DatabaseDsn:     "d",
	EnableHTTPS:     "s",
}

var instance *Config

// JSONConfig for json config
type JSONConfig struct {
	BaseURL         string `json:"base_url"`
	ServerAddress   string `json:"server_address"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDsn     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// Instance new Config
func Instance() *Config {
	if instance == nil {
		instance = new(Config)
		instance.initInv()
		instance.init()
		// Init logger
		l, err := logger.New()
		if err != nil {
			log.Fatal(err)
		}
		instance.Logger = l
		// Database
		dsn, _ := instance.Param(DatabaseDsn)
		var dbc *sql.DB
		err = nil
		if dsn != "" {
			dbc, err = db.New(l, dsn)
			if err != nil {
				l.Info("Db error", zap.Error(err))
			} else {
				instance.Database = dbc
			}
		}
		// Main handler
		if dsn != "" && err == nil {
			l.Info("Set db handler")
			instance.Storage, err = dbh.New(dbc, l)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			l.Info("Set file handler")
			// File and memory storage
			fs, err := instance.Param(FileStoragePath)
			if err != nil || fs == FileStoragePathDefault {
				fs = ""
			}
			instance.Storage, err = file.New(fs)
			if err != nil {
				log.Fatal(err)
			}
		}
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
	case EnableHTTPS:
		return c.EnableHTTPS, nil
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
	// Parse from flags
	bu := flag.String(mapVarToInv[BaseURL], "", "")
	sa := flag.String(mapVarToInv[ServerAddress], "", "")
	fs := flag.String(mapVarToInv[FileStoragePath], "", "")
	dbDSN := flag.String(mapVarToInv[DatabaseDsn], "", "")
	ssl := flag.String(mapVarToInv[EnableHTTPS], "", "")
	flag.Parse()

	// Parse from env
	if *bu != "" {
		c.BaseURL = *bu
	}
	if *sa != "" {
		c.ServerAddress = *sa
	}
	if *fs != "" {
		c.FileStoragePath = *fs
	}
	if *dbDSN != "" {
		c.DatabaseDsn = *dbDSN
	}
	if *ssl != "" {
		c.EnableHTTPS = *ssl
	}

	// Init from json evn config
	// Open our jsonFile
	pwd, _ := os.Getwd()
	byteValue, err := ioutil.ReadFile(pwd + "/../../configs/env.json")

	// if we os.Open returns an error then handle it
	if err != nil {
		// Nothing to do
		return
	}
	// we initialize our Users array
	var config JSONConfig

	// jsonFile's content into 'config' which we defined above
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return
	}

	if c.BaseURL == "" {
		c.BaseURL = config.BaseURL
	}
	if c.ServerAddress == "" {
		c.ServerAddress = config.ServerAddress
	}
	if c.FileStoragePath == "" {
		if _, err := os.Stat(config.FileStoragePath); !errors.Is(err, os.ErrNotExist) {
			c.FileStoragePath = config.FileStoragePath
		}
	}
	if c.DatabaseDsn == "" {
		c.DatabaseDsn = config.DatabaseDsn
	}
	if c.EnableHTTPS == "" {
		c.EnableHTTPS = strconv.FormatBool(config.EnableHTTPS)
	}
}
