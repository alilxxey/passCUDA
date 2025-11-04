package gpio

import "github.com/warthog618/go-gpiocdev"
import "fmt"

type GPIO struct {
	chip   string
	device int
	line   *gpiocdev.Line
}

func InitOutput(chip string, device int, value int) (*GPIO, error) {
	line, err := gpiocdev.RequestLine(chip, device, gpiocdev.AsOutput(value))
	if err != nil {
		return nil, fmt.Errorf("failed to request GPIO line: %w", err)
	}
	return &GPIO{line: line, chip: chip, device: device}, nil
}

func (gpio *GPIO) SetValue(value int) {
	gpio.line.SetValue(value)
}

func (gpio *GPIO) Deinit() error {
	return gpio.line.Close()
}
