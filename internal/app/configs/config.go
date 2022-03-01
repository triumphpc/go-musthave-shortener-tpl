package configs

import (
	"database/sql"
	"errors"
	"flag"
	"github.com/caarlos0/env"
	_ "github.com/caarlos0/env/v6"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/helpers/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
	dbh "github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/db"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/file"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/storages/repository"
	"go.uber.org/zap"
	"log"
)

var ErrUnknownParam = errors.New("unknown param")

// Config project
type Config struct {
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"unknown"`
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:":8080"`
	DatabaseDsn     string `env:"DATABASE_DSN" envDefault:""`
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
	dbDSN := flag.String(mapVarToInv[DatabaseDsn], "", "")
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
	if *dbDSN != "" {
		c.DatabaseDsn = *dbDSN
	}
}
