package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// New create new logger
func New(level string) (*zap.Logger, error) {
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

func customMillisTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.UTC().Format("2006-01-02T15:04:05.000Z07"))
}
