// Package logger سازنده‌ی لاگر zap برای کل برنامه است؛ خروجی پروداکشن JSON
// با نمونه‌گیری است و محیط توسعه خروجی کنسولی رنگی می‌دهد. همه‌ی اجزا
// (HTTP، بات، دیتابیس و سرویس‌ها) لاگرشان از همین‌جا ساخته می‌شود.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New لاگر ریشه‌ی برنامه را با سطح و محیط مشخص می‌سازد.
func New(level, appEnv string) (*zap.Logger, error) {
	var config zap.Config
	if appEnv == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		// در پروداکشن حجم لاگ کنترل می‌شود: بعد از ۱۰۰ لاگ مشابه، هر صدتا یکی.
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

// NewNop لاگر بی‌اثر برای تست‌ها برمی‌گرداند.
func NewNop() *zap.Logger {
	return zap.NewNop()
}
