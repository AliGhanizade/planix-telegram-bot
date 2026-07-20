// Package logger سازنده‌ی لاگر zap بر اساس سطح لاگ است.
package logger

import "go.uber.org/zap"

// New لاگر zap را با سطح مشخص‌شده برمی‌گرداند؛ سطح debug خروجی توسعه‌ی خواناتری دارد.
func New(level string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	if level == "debug" {
		config = zap.NewDevelopmentConfig()
	}
	return config.Build()
}
