// Package logger builds the zap logger for the whole app; production uses json
// with sampling and development uses colored console output. All parts
// (HTTP, bot, database and services) get their logger from here.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New builds the root logger for the given level and environment.
func New(level, appEnv string) (*zap.Logger, error) {
	var config zap.Config
	if appEnv == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// production caps log volume: after 100 similar entries, one in a hundred.
		config.Sampling = &zap.SamplingConfig{Initial: 100, Thereafter: 100}
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	parsed, err := zapcore.ParseLevel(level)
	if err != nil {
		parsed = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(parsed)

	build, err := config.Build(zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return nil, err
	}
	return build.With(
		zap.String("service", "planix"),
		zap.String("env", appEnv),
	), nil
}

// NewNop returns a no-op logger for tests.
func NewNop() *zap.Logger {
	return zap.NewNop()
}
