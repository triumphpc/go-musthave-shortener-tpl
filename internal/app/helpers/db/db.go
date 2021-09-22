package db

import (
	"database/sql"
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
)

// ErrDatabaseNotAvailable if db not determine
var ErrDatabaseNotAvailable = errors.New("db not available")

// New instance for db connection
func New() (*sql.DB, error) {
	dsn, _ := configs.Instance().Param(configs.DatabaseDsn)
	if dsn == "" {
		return nil, ErrDatabaseNotAvailable
	}
	// Database init
	inst, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	logger.Info("Connect to database")
	return inst, nil
}
