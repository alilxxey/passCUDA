package telegram

import (
	tgmodels "github.com/go-telegram/bot/models"
)

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
