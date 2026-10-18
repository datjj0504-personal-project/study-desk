package main

import "context"

const appVersion = "0.1.0"

// AppInfo describes the running application.
type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// App is the Wails-bound adapter. Methods are added here as features land.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetAppInfo returns basic application metadata to the frontend.
func (a *App) GetAppInfo() (AppInfo, error) {
	return AppInfo{Name: "StudyDesk", Version: appVersion}, nil
}
