package telegram

import (
	"context"
	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"paSKUDa/internal/models"
	"go.uber.org/zap"
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

	b.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/version",
		bot.MatchTypeExact,
		tg.authMiddleware(config.UserRoleAdmin, tg.getVersionHandler()),
	)

	return tg, nil

}

func extractUserAndChat(update *tgmodels.Update) (userID int64, chatID int64, ok bool) {
	switch {
	case update.Message != nil:
		return update.Message.From.ID, update.Message.Chat.ID, true

	//case update.CallbackQuery != nil:
	//	msg := update.CallbackQuery.Message.GetMessage()
	//	return msg.Chat.ID, update.CallbackQuery.From.ID, true

	case update.MyChatMember != nil:
		return update.MyChatMember.From.ID, update.MyChatMember.Chat.ID, true
	}

	return 0, 0, false
}

func (tg *Telegram) authMiddleware(minRole config.UserRole, next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		userID, chatID, ok := extractUserAndChat(update)
		if !ok {
			return
		}
		user, err := tg.conf.FindUserById(userID)
		if err != nil {
			zap.S().Warnf("unauthorized access from user: `%d`, chat: `%d`", userID, chatID)
			return
		}
		if !user.IsChatAllowedForUser(chatID) {
			zap.S().Warnf("user: `%d` not allowed to access from chat: `%d`", userID, chatID)
			return
		}
		if user.Role < minRole {
			zap.S().Warnf("user: `%d` role: `%s` is too low for acess this resource, expected: `%s`", userID, user.Role, minRole)
			return
		}
		next(ctx, b, update)
	}
}

func (tg *Telegram) getVersionHandler() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		ver, err := models.GetPrintableVersionInfo()
		if err != nil {
			zap.S().Errorf("can't get server version info: `%w`", err)
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      bot.EscapeMarkdown(ver),
			ParseMode: tgmodels.ParseModeMarkdown,
		})
	}
}

func (tg *Telegram) Start() {
	tg.bot.Start(tg.ctx)
}
