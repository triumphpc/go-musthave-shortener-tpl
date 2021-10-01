package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/triumphpc/go-musthave-shortener-tpl/internal/app/configs"
	"go.uber.org/zap"
)

// ErrDatabaseNotAvailable if db not determine
var ErrDatabaseNotAvailable = errors.New("db not available")

// New instance for db connection
func New(l *zap.Logger) (*sql.DB, error) {
	dsn, _ := configs.Instance().Param(configs.DatabaseDsn)
	if dsn == "" {
		return nil, ErrDatabaseNotAvailable
	}
	// Database init
	inst, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Ping
	if err := inst.PingContext(context.Background()); err != nil {
		return nil, err
	}

	l.Info("Connect to database")
	return inst, nil
}
