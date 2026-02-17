package app

import (
	"context"
	"go.uber.org/zap"
	"paSKUDa/internal/card"
	"paSKUDa/internal/config"
	"paSKUDa/internal/door"
	"paSKUDa/internal/telegram"
)

type App struct {
	ctx    context.Context
	config *config.Config
	tg     *telegram.Telegram
	door   *door.Door
	card   *card.CardManager
}

func Init(
	config *config.Config,
	ctx context.Context,
	tg *telegram.Telegram,
	door *door.Door,
	card *card.CardManager,
) (*App, error) {
	zap.S().Infof("init done")
	return &App{ctx: ctx, tg: tg, config: config, door: door, card: card}, nil
}

func (app *App) Run() error {
	if err := app.tg.Start(); err != nil {
		return err
	}
	if err := app.card.Start(); err != nil {
		return err
	}
	if err := app.door.Start(); err != nil {
		return err
	}
	zap.S().Infof("startup completed")
	<-app.ctx.Done()
	zap.S().Infof("shutdown completed")
	return nil
}
