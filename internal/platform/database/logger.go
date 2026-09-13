package database

import (
	"context"
	"errors"
	"time"

	"strconv"

	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// zapGormLogger routes gorm logging into the app zap logger.
type zapGormLogger struct {
	log   *zap.Logger
	cfg   gormlogger.Config
	level gormlogger.LogLevel
}

// NewGormLogger adapts gorm logging into zap and flags slow queries.
func NewGormLogger(log *zap.Logger) gormlogger.Interface {
	return &zapGormLogger{
		log: log,
		cfg: gormlogger.Config{
			SlowThreshold:             500 * time.Millisecond,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			LogLevel:                  gormlogger.Warn,
		},
		level: gormlogger.Warn,
	}
}

func (l *zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	next := *l
	next.level = level
	return &next
}

func (l *zapGormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Info {
		l.log.Info(msg, argsToFields(args)...)
	}
}

func (l *zapGormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Warn {
		l.log.Warn(msg, argsToFields(args)...)
	}
}

func (l *zapGormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormlogger.Error {
		l.log.Error(msg, argsToFields(args)...)
	}
}

// Trace inspects each query: errors, slow ones and debug level traces.
func (l *zapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.cfg.IgnoreRecordNotFoundError && errors.Is(err, gormlogger.ErrRecordNotFound):
		return
	case err != nil:
		l.log.Error("database query failed",
			zap.Error(err),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	case l.cfg.SlowThreshold != 0 && elapsed > l.cfg.SlowThreshold:
		l.log.Warn("slow database query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
			zap.Duration("threshold", l.cfg.SlowThreshold),
		)
	case l.level >= gormlogger.Info:
		l.log.Debug("database query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	}
}

// argsToFields converts gorm log args into zap fields.
func argsToFields(args []interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(args))
	for i, a := range args {
		fields = append(fields, zap.Any("arg_"+strconv.Itoa(i), a))
	}
	return fields
}
