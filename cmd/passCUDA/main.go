package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"paSKUDa/internal/app"
	"paSKUDa/internal/card"
	"paSKUDa/internal/config"
	"paSKUDa/internal/door"
	"paSKUDa/internal/gpio"
	"paSKUDa/internal/models"
	"paSKUDa/internal/telegram"

	"go.uber.org/zap"
)

func initLogging() *zap.Logger {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	zap.ReplaceGlobals(logger)
	return logger
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logger := initLogging()
	defer logger.Sync()

	c, err := config.Load("/etc/passCUDA/config.yml")
	if err != nil {
		zap.S().Fatalf("failed to load config: %v", err)
	}

	openChan := make(chan models.DoorSignal, 10)
	adminMessageChan := make(chan string, 10)

	gpio, err := gpio.Init(c, ctx, openChan, adminMessageChan)
	if err != nil {
		zap.S().Fatalf("failed to init gpio: %v", err)
	}
	defer gpio.Deinit()

	door, err := door.Init(c, ctx, gpio, openChan, adminMessageChan)
	if err != nil {
		zap.S().Fatalf("failed to init door: %v", err)
	}

	tg, err := telegram.Init(c, ctx, c.Telegram.Token, openChan, adminMessageChan, door)
	if err != nil {
		zap.S().Fatalf("failed to init tg: %v", err)
	}

	card, err := card.Init(c, ctx, openChan, adminMessageChan, door)
	if err != nil {
		zap.S().Fatalf("failed to init card: %v", err)
	}

	app, err := app.Init(c, ctx, tg, door, card)
	if err != nil {
		zap.S().Fatalf("failed to init app: %v", err)
	}
	if err := app.Run(); err != nil {
		zap.S().Fatalf("failed to start app: %v", err)
	}

	<-ctx.Done()
}
