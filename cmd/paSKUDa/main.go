package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"paSKUDa/internal/config"
	//"paSKUDa/internal/gpio"
	"paSKUDa/internal/app"
	"paSKUDa/internal/telegram"

	"os/exec"

	"go.uber.org/zap"
)

func todoRemove() {
	cmd := exec.Command("/bin/bash", "-c", "kill $(gpioinfo | grep GPIO3 | awk -F 'gpiocdev-' '{print $2}' | tr -d '\"')")
	_, err := cmd.Output()
	if err != nil {
		zap.S().Infof("failed to exec: %w", err)
	}
}

func initLogging() *zap.Logger {
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
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

	//todoRemove()

	//relay, err := gpio.InitOutput("gpiochip0", 2, 1)
	//if err != nil {
	//	zap.S().Fatalf("failed to init relay: %w", err)
	//}

	//openBtn, err := gpio.InitTest("gpiochip0", 3)
	//if err != nil {
	//	zap.S().Fatalf("failed to init open btn: %w", err)
	//}

	c, err := config.Load("/tmp/config.yml")
	if err != nil {
		zap.S().Fatalf("failed to load config: %w", err)
	}

	tg, err := telegram.Init(c, ctx, c.Telegram.Token)
	if err != nil {
		zap.S().Fatalf("failed to init tg: %w", err)
	}

	app, err := app.Init(ctx, tg, c)
	if err != nil {
		zap.S().Fatalf("failed to init app: %w", err)
	}
	app.Run()

	//go tg.Start()

	<-ctx.Done()

	//for {
	//	val, _ := openBtn.GetValue()
	//	if val == 0 {
	//		relay.SetValue(0)
	//		time.Sleep(1 * time.Second)
	//		relay.SetValue(1)
	//		time.Sleep(1 * time.Second)
	//	}
	//}

	//if err := relay.Deinit(); err != nil {
	//	zap.S().Fatalf("failed to deinit relay: %w", err)
	//}
	//if err := openBtn.Deinit(); err != nil {
	//	zap.S().Fatalf("failed to deinit relay: %w", err)
	//}
}
