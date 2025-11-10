package door

import (
	"context"
	"go.uber.org/zap"
	"paSKUDa/internal/config"
	"paSKUDa/internal/gpio"
	"paSKUDa/internal/models"
	"time"
)

type Door struct {
	ctx              context.Context
	config           *config.Config
	gpio             *gpio.Gpio
	openChan         chan models.DoorSignal
	adminMessageChan chan string
	OpenInProgress   bool
}

func Init(
	config *config.Config,
	ctx context.Context,
	gpio *gpio.Gpio,
	openChan chan models.DoorSignal,
	adminMessageChan chan string,
) (*Door, error) {
	zap.S().Infof("init done")
	return &Door{
		config:           config,
		ctx:              ctx,
		gpio:             gpio,
		openChan:         openChan,
		adminMessageChan: adminMessageChan,
		OpenInProgress:   false,
	}, nil
}

func (d *Door) Open() {
	d.OpenInProgress = true
	zap.S().Infof("opening door")
	d.gpio.DoorRelayPin.SetValue(!d.config.Hardware.DoorRelay.ActiveLevel)
	time.Sleep(time.Duration(d.config.Door.RelayOffHoldTimeS) * time.Second)
	d.gpio.DoorRelayPin.SetValue(d.config.Hardware.DoorRelay.ActiveLevel)
	d.OpenInProgress = false
	zap.S().Infof("door opened")
}

func (d *Door) Start() error {
	go func() {
		for {
			select {
			case <-d.ctx.Done():
				return
			case sig, ok := <-d.openChan:
				if !ok {
					return
				}
				if sig == models.DoorSignalOpen {
					d.Open()
				}
			}
		}
	}()
	return nil
}
