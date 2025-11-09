package app

import (
	"context"
	"go.uber.org/zap"
	"paSKUDa/internal/config"
	"paSKUDa/internal/door"
	"paSKUDa/internal/telegram"
)

type App struct {
	ctx    context.Context
	config *config.Config
	tg     *telegram.Telegram
	door   *door.Door
}

func Init(config *config.Config, ctx context.Context, tg *telegram.Telegram, door *door.Door) (*App, error) {
	return &App{ctx: ctx, tg: tg, config: config, door: door}, nil
}

func (app *App) Run() {
	go app.tg.Start()
	zap.S().Infof("startup completed")
	<-app.ctx.Done()
	zap.S().Infof("shutdown completed")
}
