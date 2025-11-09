package gpio

import "context"
import "time"
import "paSKUDa/internal/config"

type Gpio struct {
	config       *config.Config
	ctx          context.Context
	doorRelayPin *Pin
	openBtnPin   *Pin
}

func Init(config *config.Config, ctx context.Context) (*Gpio, error) {
	doorRelay := config.Hardware.DoorRelay
	doorRelayPin, err := InitOutput(doorRelay.Chip, doorRelay.Channel, !doorRelay.ActiveLevel)
	if err != nil {
		return nil, err
	}

	openBtn := config.Hardware.OpenBtn
	deboucePeriod := 10 * time.Millisecond
	openBtnPin, err := InitInputDebounce(openBtn.Chip, openBtn.Channel, deboucePeriod)
	if err != nil {
		return nil, err
	}

	return &Gpio{
		config:       config,
		ctx:          ctx,
		doorRelayPin: doorRelayPin,
		openBtnPin:   openBtnPin,
	}, nil
}

func (g *Gpio) inputPolling() error {
	openBtnPinValue, err := g.openBtnPin.GetValue()
	if err != nil {
		return err
	}
	if openBtnPinValue == g.config.Hardware.OpenBtn.ActiveLevel {
	}
	//for {
	//	val, _ := openBtn.GetValue()
	//	if val == 0 {
	//		relay.SetValue(0)
	//		time.Sleep(1 * time.Second)
	//		relay.SetValue(1)
	//		time.Sleep(1 * time.Second)
	//	}
	//}
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
	g.doorRelayPin.Deinit()
}
