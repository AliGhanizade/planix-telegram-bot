// Package database manages the database connection and schema migration.
package database

import (
	"github.com/AliGhanizade/planix-telegram-bot/internal/domain"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open opens the postgres connection, wires the zap logger into gorm
// and migrates all models.
func Open(url string, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: NewGormLogger(log)})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&domain.User{}, &domain.Planner{}, &domain.Task{}, &domain.TaskEvidence{}, &domain.ActivityLog{}, &domain.BotSession{}, &domain.LoginCode{}, &domain.WebSession{}, &domain.Folder{}, &domain.TaskFolderLink{}); err != nil {
		return nil, err
	}
	return db, nil
}
