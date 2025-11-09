package telegram

import (
	"context"
	"github.com/go-telegram/bot"
	"paSKUDa/internal/config"
	//"github.com/go-telegram/ui/keyboard/reply"
)

type Telegram struct {
	ctx  context.Context
	bot  *bot.Bot
	conf *config.Config
}

func Init(c *config.Config, ctx context.Context, token string) (*Telegram, error) {
	b, err := bot.New(token)
	if err != nil {
		return nil, err
	}

	if _, err := b.GetMe(ctx); err != nil {
		return nil, err
	}

	tg := &Telegram{
		conf: c,
		ctx:  ctx,
		bot:  b,
	}

	tg.registerHandlers()

	return tg, nil
}

func (tg *Telegram) Start() {
	tg.bot.Start(tg.ctx)
}
