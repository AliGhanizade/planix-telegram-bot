// Package database اتصال دیتابیس و مهاجرت اسکیما را مدیریت می‌کند.
package database

import (
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open اتصال PostgreSQL را باز می‌کند و همه‌ی مدل‌ها را مهاجرت می‌دهد.
func Open(url string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Planner{}, &domain.Task{}, &domain.TaskEvidence{}, &domain.ActivityLog{}, &domain.BotSession{}); err != nil {
		return nil, err
	}
	return db, nil
}
