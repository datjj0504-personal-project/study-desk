package main

import "context"

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
