package main

import "log"
import "time"

//import "paSKUDa/internal/config"
import "paSKUDa/internal/gpio"

import "go.uber.org/zap"

func main() {
	zapConfig := zap.NewDevelopmentConfig()
    zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
    logger, err := zapConfig.Build()
    if err != nil {
        log.Fatalf("can't initialize zap logger: %v", err)
    }
    defer logger.Sync()
    zap.ReplaceGlobals(logger)

	relay, err := gpio.InitOutput("gpiochip0", 2, 1)
	if err != nil {
		zap.S().Fatalf("failed to init gpio: %w", err)
	}

	relay.SetValue(0)
	time.Sleep(1 * time.Second)
	relay.SetValue(1)
	time.Sleep(1 * time.Second)

	if err := relay.Deinit(); err != nil {
		zap.S().Fatalf("failed to deinit relay: %w", err)
	}
	c, err := config.Load("/etc/passCUDA/config.yml")
	if err != nil {
		zap.S().Fatalf("failed to load config: %w", err)
	}
	fmt.Printf("%s", c.Telegram.Token)
}
