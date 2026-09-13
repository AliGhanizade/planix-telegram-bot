// Package database اتصال دیتابیس و مهاجرت اسکیما را مدیریت می‌کند.
package database

import (
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open اتصال PostgreSQL را باز می‌کند، لاگر zap را به GORM وصل می‌کند
// و همه‌ی مدل‌ها را مهاجرت می‌دهد.
func Open(url string, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: NewGormLogger(log)})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Planner{}, &domain.Task{}, &domain.TaskEvidence{}, &domain.ActivityLog{}, &domain.BotSession{}); err != nil {
		return nil, err
	}
	return db, nil
}
