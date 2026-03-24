package telegram

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	"go.uber.org/zap"
	"paSKUDa/internal/config"
	"paSKUDa/internal/door"
	"paSKUDa/internal/models"
	"paSKUDa/internal/webcam"
	"time"
)

type Telegram struct {
	ctx              context.Context
	bot              *bot.Bot
	conf             *config.Config
	openChan         chan models.DoorSignal
	adminMessageChan chan string
	door             *door.Door
	webcam           *webcam.Webcam
	tempBanList      []int64
}

func Init(
	c *config.Config,
	ctx context.Context,
	token string,
	openChan chan models.DoorSignal,
	adminMessageChan chan string,
	door *door.Door,
	webcam *webcam.Webcam,
) (*Telegram, error) {
	b, err := bot.New(token)
	if err != nil {
		return nil, err
	}

	if _, err := b.GetMe(ctx); err != nil {
		return nil, err
	}

	tg := &Telegram{
		conf:             c,
		ctx:              ctx,
		bot:              b,
		openChan:         openChan,
		adminMessageChan: adminMessageChan,
		door:             door,
		webcam:           webcam,
		tempBanList:      []int64{},
	}

	tg.registerHandlers()

	zap.S().Infof("init done")
	return tg, nil
}

func (tg *Telegram) Start() error {
	go tg.bot.Start(tg.ctx)
	go func() {
		for {
			select {
			case <-tg.ctx.Done():
				return
			case msg, ok := <-tg.adminMessageChan:
				if !ok {
					return
				}
				tg.NotifyAdmins(msg)
			}
		}
	}()
	return nil
}

func (tg *Telegram) NotifyAdmins(str string) error {
	msg := fmt.Sprintf("[%s]: %s",
		time.Now().UTC().Format("02-01-2006 15:04:05 UTC"),
		str,
	)
	_, err := tg.bot.SendMessage(tg.ctx, &bot.SendMessageParams{
		ChatID: tg.conf.Telegram.NotifyChatId,
		Text:   msg,
	})
	if err != nil {
		zap.S().Errorf("failed to send admin msg: `%s`, reason: `%v`", msg, err)
	}
	return err
}
