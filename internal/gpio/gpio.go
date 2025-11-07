package gpio

import "github.com/warthog618/go-gpiocdev"
import "fmt"
import "time"

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

func InitInput(chip string, device int) (*GPIO, error) {
	line, err := gpiocdev.RequestLine(chip, device, gpiocdev.AsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to request GPIO line: %w", err)
	}
	return &GPIO{line: line, chip: chip, device: device}, nil
}
func InitInputDebounce(chip string, device int) (*GPIO, error) {
	period := 10 * time.Millisecond
	line, err := gpiocdev.RequestLine(chip, device, gpiocdev.AsInput, gpiocdev.WithDebounce(period))
	if err != nil {
		return nil, fmt.Errorf("failed to request GPIO line: %w", err)
	}
	return &GPIO{line: line, chip: chip, device: device}, nil
}

func (gpio *GPIO) SetValue(value int) {
	gpio.line.SetValue(value)
}

func (gpio *GPIO) GetValue() (int, error) {
	return gpio.line.Value()
}

func (gpio *GPIO) Deinit() error {
	return gpio.line.Close()
}
