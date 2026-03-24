package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	tgmodels "github.com/go-telegram/bot/models"
	"github.com/go-telegram/ui/keyboard/reply"
	"go.uber.org/zap"
	"os"
	"bytes"
	"paSKUDa/internal/config"
	"paSKUDa/internal/models"
	"path/filepath"
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
		"/photo",
		bot.MatchTypeExact,
		tg.authMiddleware(config.UserRoleAdmin, tg.getPhotoHandler()),
	)
	tg.bot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/start",
		bot.MatchTypeExact,
		tg.authMiddleware(config.UserRoleUser, tg.getStartHandler()),
	)
	tg.bot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"/id",
		bot.MatchTypeExact,
		tg.idHandler,
	)
	tg.bot.RegisterHandler(
		bot.HandlerTypeMessageText,
		"Open",
		bot.MatchTypeExact,
		tg.authMiddleware(
			config.UserRoleUser,
			tg.getOpenHandler(),
		),
	)

}

func (tg *Telegram) idHandler(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
	userID, chatID, userName, ok := extractUserAndChat(update)
	if !ok {
		return
	}
	msg := fmt.Sprintf("chatID: `%d`, userID: `%d`, userName: `%s`", chatID, userID, userName)
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   msg,
	})
}

func (tg *Telegram) getMainReplyKbd() *reply.ReplyKeyboard {
	kbd := reply.New(
		reply.WithPrefix("main_keyboard"),
		reply.IsPersistent(),
	).
		Button("Open",
			tg.bot,
			bot.MatchTypeExact,
			tg.authMiddleware(
				config.UserRoleUser,
				tg.getOpenHandler(),
			),
		)
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
			zap.S().Errorf("can't get server version info: `%v`", err)
			return
		}
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID:    update.Message.Chat.ID,
			Text:      bot.EscapeMarkdown(ver),
			ParseMode: tgmodels.ParseModeMarkdown,
		})
	}
}

func (tg *Telegram) getPhotoHandler() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		if !tg.conf.Webcam.Enabled {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID:    update.Message.Chat.ID,
				Text:      "Camera not enabled!",
				ParseMode: tgmodels.ParseModeMarkdown,
			})
			return
		}
		photoFile, err := tg.webcam.GetPhoto("1920x1080", 5)
		if err != nil {
			zap.S().Errorf("can't get photo: `%v`", err)
			return
		}
		defer os.Remove(photoFile)

		fileData, err := os.ReadFile(photoFile)
		if err != nil {
			zap.S().Errorf("can't read photo: `%v`", err)
			return
		}

		params := &bot.SendPhotoParams{
			ChatID:  update.Message.Chat.ID,
			Photo:   &tgmodels.InputFileUpload{Filename: filepath.Base(photoFile), Data: bytes.NewReader(fileData)},
			Caption: "Photo",
		}

		b.SendPhoto(ctx, params)

		userID, _, _, _ := extractUserAndChat(update)
		user, _ := tg.conf.FindUserById(userID)
		select {
		case tg.adminMessageChan <- fmt.Sprintf("photo is taken by user: `%v`", user):
		default:
		}
	}
}

func (tg *Telegram) getOpenHandler() bot.HandlerFunc {
	return func(ctx context.Context, b *bot.Bot, update *tgmodels.Update) {
		userID, _, _, _ := extractUserAndChat(update)
		user, _ := tg.conf.FindUserById(userID)
		if tg.door.OpenInProgress {
			zap.S().Warnf("user: `%s` trying to open door while previous open in progress", user)
			return
		}

		select {
		case tg.openChan <- models.DoorSignalOpen:
		default:
		}

		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "done!",
		})

		select {
		case tg.adminMessageChan <- fmt.Sprintf("door opened by user: `%v`", user):
		default:
		}
	}
}
