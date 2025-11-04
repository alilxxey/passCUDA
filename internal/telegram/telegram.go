package telegram

import (
	"context"
	"time"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	//"github.com/go-telegram/ui/keyboard/reply"
	"go.uber.org/zap"
	"paSKUDa/internal/gpio"
)

type Telegram struct {
	bot   *bot.Bot
	relay *gpio.GPIO
}

func New(token string, relay *gpio.GPIO) *Telegram {
	b, err := bot.New(token)
	if err != nil {
		zap.S().Fatalf("failed to init telegram bot: %v", err)
	}

	user, _ := b.GetMe(context.Background())
	zap.S().Infof("Telegram bot info: %#v\n", user)

	tg := &Telegram{
		bot:   b,
		relay: relay,
	}

	b.RegisterHandler(bot.HandlerTypeMessageText, "/fuckuTvarBlyad", bot.MatchTypeExact, makeFooHandler(tg))

	return tg
}

func (tg *Telegram) Start() {
	tg.bot.Start(context.Background())
}

func makeFooHandler(tg *Telegram) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *models.Update) {
		tg.relay.SetValue(0)
		time.Sleep(1 * time.Second)
		tg.relay.SetValue(1)
		time.Sleep(1 * time.Second)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      "Hello, *" + bot.EscapeMarkdown(update.Message.From.FirstName) + "*",
			ParseMode: models.ParseModeMarkdown,
		})
	}
}
