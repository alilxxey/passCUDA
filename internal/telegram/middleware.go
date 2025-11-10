package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"paSKUDa/internal/config"
)

func (tg *Telegram) authMiddleware(minRole config.UserRole, next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		userID, chatID, ok := extractUserAndChat(update)
		if !ok {
			return
		}
		user, err := tg.conf.FindUserById(userID)
		if err != nil {
			msg := fmt.Sprintf("unauthorized access from user: `%d`, chat: `%d`", userID, chatID)
			tg.NotifyAdmins(msg)
			zap.S().Warn(msg)
			return
		}
		if !user.IsChatAllowedForUser(chatID) {
			msg := fmt.Sprintf("user: `%d` not allowed to access from chat: `%d`", userID, chatID)
			tg.NotifyAdmins(msg)
			zap.S().Warn(msg)
			return
		}
		if user.Role < minRole {
			msg := fmt.Sprintf(
				"user: `%d` role: `%s` is too low for acess this resource, expected: `%s`",
				userID,
				user.Role,
				minRole,
			)
			tg.NotifyAdmins(msg)
			zap.S().Warn(msg)
			return
		}
		next(ctx, b, update)
	}
}
