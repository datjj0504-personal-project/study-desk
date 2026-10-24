package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"studydesk/internal/storage"
)

const appVersion = "0.1.0"

// AppInfo describes the running application.
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// App is the Wails-bound adapter. Methods are added here as features land.
type App struct {
	ctx        context.Context
	db         *sql.DB
	startupErr error
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.db, a.startupErr = storage.OpenAppDatabase(ctx)
	if a.startupErr != nil {
		log.Printf("initialize local database: %v", a.startupErr)
	}
}

func (a *App) shutdown(context.Context) {
	if a.db == nil {
		return
	}
	if err := a.db.Close(); err != nil {
		log.Printf("close local database: %v", err)
	}
}

// GetAppInfo returns basic application metadata to the frontend.
func (a *App) GetAppInfo() (AppInfo, error) {
	if a.startupErr != nil {
		return AppInfo{}, fmt.Errorf("initialize local database: %w", a.startupErr)
	}
	return AppInfo{Name: "StudyDesk", Version: appVersion}, nil
}
