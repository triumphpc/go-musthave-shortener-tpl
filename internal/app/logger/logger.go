package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// instance singleton for logger
var instance *zap.Logger

// Instance new Config
func Instance() (*zap.Logger, error) {
	if instance == nil {
		instance = new(zap.Logger)
		logger, err := newLogger("info")
		if err != nil {
			return logger, err
		}
		instance = logger
	}
	return instance, nil
}

// new create new logger
func newLogger(level string) (*zap.Logger, error) {
	// Init config
	cfg := zap.NewProductionConfig()
	// Set level
	cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	// Log level
	atom := zap.NewAtomicLevel()
	if err := atom.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}
	cfg.Level = atom
	// Output set
	cfg.OutputPaths = []string{"stdout"}
	// Time format
	cfg.EncoderConfig.EncodeTime = customMillisTimeEncoder
	return cfg.Build()
}

// customMillisTimeEncoder set time format
func customMillisTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02 15:04:05"))
}

// Info write data to info level
func Info(msg string, fields ...zap.Field) {
	// Init logger instance
	l, err := Instance()
	if err == nil {
		l.Info(msg, fields...)
	}
}
