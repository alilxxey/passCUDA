package telegram

import (
	"context"
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
	"paSKUDa/internal/config"
	"paSKUDa/internal/door"
	"paSKUDa/internal/models"
	//"github.com/go-telegram/ui/keyboard/reply"
)

type Telegram struct {
	ctx      context.Context
	bot      *bot.Bot
	conf     *config.Config
	openChan chan models.DoorSignal
	door     *door.Door
}

func Init(
	c *config.Config,
	ctx context.Context,
	token string,
	openChan chan models.DoorSignal,
	door *door.Door,
) (*Telegram, error) {
	b, err := bot.New(token)
	if err != nil {
		return nil, err
	}

	if _, err := b.GetMe(ctx); err != nil {
		return nil, err
	}

	tg := &Telegram{
		conf:     c,
		ctx:      ctx,
		bot:      b,
		openChan: openChan,
		door:     door,
	}

	tg.registerHandlers()

	zap.S().Infof("init done")
	return tg, nil
}

func (tg *Telegram) Start() error {
	go tg.bot.Start(tg.ctx)
	return nil
}
