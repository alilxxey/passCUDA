package telegram

import (
	 "github.com/go-telegram/bot"
	 "github.com/go-telegram/ui"
	 "go.uber.org/zap"
)

type Telegram struct {
    bot    *bot.Bot
}

func New(c *models.TelegramConfig) *Telegram {
    b, err := bot.New(c.BotToken)
    if err != nil {
        zap.S().Fatalf("failed to init telegram bot: %v", err)
    }

    user, _ := b.GetMe(context.Background())
    zap.S().Infof("Telegram bot info: %#v\n", user)

    return &Telegram{
        bot:    b,
        chatID: c.ChatID,
    }
}
