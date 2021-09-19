package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/logger"
)

// ErrDatabaseNotAvailable if db not determine
var ErrDatabaseNotAvailable = errors.New("db not available")

// instance singleton for DB
var instance *sql.DB

// Instance new Config
func Instance() (*sql.DB, error) {
	if instance == nil {
		dsn, _ := configs.Instance().Param(configs.DatabaseDsn)
		if dsn == "" {
			return instance, ErrDatabaseNotAvailable
		}
		instance = new(sql.DB)
		// Database init
		inst, err := sql.Open("postgres", dsn)
		if err != nil {
			return instance, err
		}
		instance = inst
		logger.Info("Connect to database")
	}
	return instance, nil
}

// Insert execute query to active connect
func Insert(ctx context.Context, query string, args ...interface{}) error {
	c, err := Instance()
	if err == nil {
		if _, err := c.ExecContext(ctx, query, args...); err == nil {
			return nil
		} else {
			return err
		}
	}
	return err
}
