package app

import (
	"paSKUDa/internal/config"
	"paSKUDa/internal/telegram"
	"go.uber.org/zap"
	"context"
)

type App struct {
	ctx context.Context
	config *config.Config
	tg     *telegram.Telegram
}

func Init(ctx context.Context, tg *telegram.Telegram, config *config.Config) (*App, error) {
	return &App{ctx: ctx, tg: tg, config: config}, nil
}

func (app *App) Run() {
	go app.tg.Start()
	zap.S().Infof("startup completed")
	<-app.ctx.Done()
	zap.S().Infof("shutdown completed")
}
