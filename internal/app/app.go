package app

import (
	"fmt"

	"github.com/AliGhanizade/planix-telegram-bot/internal/config"
	"github.com/AliGhanizade/planix-telegram-bot/internal/httpapi"
	"github.com/AliGhanizade/planix-telegram-bot/internal/platform/database"
	"github.com/AliGhanizade/planix-telegram-bot/internal/platform/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type App struct {
	Config config.Config
	Logger *zap.Logger
	Router *gin.Engine
}

func New() (*App, error) {
	c, err := config.Load()
	if err != nil {
		return nil, err
	}
	log, err := logger.New(c.LogLevel)
	if err != nil {
		return nil, err
	}
	db, err := database.Open(c.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	return &App{Config: c, Logger: log, Router: httpapi.NewRouter(c, log)}, nil
}

func (a *App) Stop() {}
