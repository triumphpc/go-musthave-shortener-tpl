package db

import (
	"context"
	"database/sql"

	"go.uber.org/zap"
)

// New instance for db connection
func New(l *zap.Logger, dsn string) (*sql.DB, error) {
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
