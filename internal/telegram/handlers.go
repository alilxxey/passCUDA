package telegram

import (
	"context"
	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
	"go.uber.org/zap"
	"paSKUDa/internal/config"
	"paSKUDa/internal/models"
)

func (tg *Telegram) registerHandlers() {
	tg.bot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/version",
		bot.MatchTypeExact,
		tg.authMiddleware(config.UserRoleAdmin, tg.getVersionHandler()),
	)
	tg.bot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/start",
		bot.MatchTypeExact,
		tg.authMiddleware(config.UserRoleUser, tg.getStartHandler()),
	)

}

func (tg *Telegram) getMainReplyKbd() *reply.ReplyKeyboard {
	kbd := reply.New(
		reply.WithPrefix("main_keyboard"),
		reply.IsPersistent(),
	).
		Button("Open", tg.bot, bot.MatchTypeExact, tg.getOpenHandler())
	return kbd
}

func (tg *Telegram) getStartHandler() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:      update.Message.Chat.ID,
			Text:        "Now u'r admitted to open the door!",
			ReplyMarkup: tg.getMainReplyKbd(),
		})
	}
}

func (tg *Telegram) getVersionHandler() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		ver, err := models.GetPrintableVersionInfo()
		if err != nil {
			zap.S().Errorf("can't get server version info: `%w`", err)
			return
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      bot.EscapeMarkdown(ver),
			ParseMode: tgmodels.ParseModeMarkdown,
		})
	}
}

func (tg *Telegram) getOpenHandler() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		// tg.dm.Open()
		msg := "done"
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   msg,
		})
	}
}
