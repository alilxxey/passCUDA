package gpio

import "context"
import "time"
import "paSKUDa/internal/config"
import "paSKUDa/internal/models"
import "github.com/warthog618/go-gpiocdev"

type Gpio struct {
	config           *config.Config
	ctx              context.Context
	DoorRelayPin     *Pin
	OpenBtnPin       *Pin
	openChan         chan models.DoorSignal
	adminMessageChan chan string
}

func Init(
	config *config.Config,
	ctx context.Context,
	openChan chan models.DoorSignal,
	adminMessageChan chan string,
) (*Gpio, error) {
	gp := &Gpio{
		config:           config,
		ctx:              ctx,
		openChan:         openChan,
		adminMessageChan: adminMessageChan,
	}

	doorRelay := config.Hardware.DoorRelay
	doorRelayPin, err := InitOutput(doorRelay.Chip, doorRelay.Channel, !doorRelay.ActiveLevel)
	if err != nil {
		return nil, err
	}
	gp.DoorRelayPin = doorRelayPin

	openBtn := config.Hardware.OpenBtn
	deboucePeriod := 10 * time.Millisecond
	openBtnPin, err := InitInputDebounce(
		openBtn.Chip,
		openBtn.Channel,
		deboucePeriod,
		config.Hardware.OpenBtn.ActiveLevel,
		gp.getDoorOpenHandler(),
	)
	if err != nil {
		return nil, err
	}
	gp.OpenBtnPin = openBtnPin

	return gp, nil
}

func (g *Gpio) getDoorOpenHandler() gpiocdev.EventHandler {
	return func(evt gpiocdev.LineEvent) {
		select {
		case g.openChan <- models.DoorSignalOpen:
		default:
		}

		select {
		case g.adminMessageChan <- "door opened by button":
		default:
		}
	}
}

func (g *Gpio) inputPolling() error {
	OpenBtnPinValue, err := g.OpenBtnPin.GetValue()
	if err != nil {
		return err
	}
	if OpenBtnPinValue == g.config.Hardware.OpenBtn.ActiveLevel {
	}
	return nil
}

func (g *Gpio) StartPolling() error {
	for {
		select {
		case <-g.ctx.Done():
			return nil
		default:
		}
		err := g.inputPolling()
		if err != nil {
			return err
		}
	}
}

func (g *Gpio) Deinit() {
	g.DoorRelayPin.Deinit()
	g.OpenBtnPin.Deinit()
}
