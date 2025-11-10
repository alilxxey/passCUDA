package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"go.uber.org/zap"
	"paSKUDa/internal/config"
	"slices"
)

func (tg *Telegram) banUserAndReport(userID int64, msg string) {
	select {
	case tg.adminMessageChan <- msg:
	default:
	}
	zap.S().Warn(msg)
	tg.tempBanList = append(tg.tempBanList, userID)
	zap.S().Warnf("user with id: `%d` is banned", userID)
}

func (tg *Telegram) authMiddleware(minRole config.UserRole, next bot.HandlerFunc) bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		userID, chatID, ok := extractUserAndChat(update)
		if !ok {
			return
		}
		if slices.Contains(tg.tempBanList, userID) {
			return
		}
		user, err := tg.conf.FindUserById(userID)
		if err != nil {
			msg := fmt.Sprintf("unauthorized access from user: `%d`, chat: `%d`", userID, chatID)
			tg.banUserAndReport(userID, msg)
			return
		}
		if !user.IsChatAllowedForUser(chatID) {
			msg := fmt.Sprintf("user: `%d` not allowed to access from chat: `%d`", userID, chatID)
			tg.banUserAndReport(userID, msg)
			return
		}
		if user.Role < minRole {
			msg := fmt.Sprintf(
				"user: `%d` role: `%s` is too low for acess this resource, expected: `%s`",
				userID,
				user.Role,
				minRole,
			)
			tg.banUserAndReport(userID, msg)
			return
		}
		next(ctx, b, update)
	}
}
