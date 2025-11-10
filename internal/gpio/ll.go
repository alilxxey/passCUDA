package gpio

// import "github.com/warthog618/go-gpiocdev"
import "fmt"
import "time"

type Direction int

const (
	DirectionInput Direction = iota
	DirectionOutput
)

type Pin struct {
	chip      string
	device    int
	direction Direction
	//line   *gpiocdev.Line
}

func InitOutput(chip string, device int, value bool) (*Pin, error) {
	//line, err := gpiocdev.RequestLine(chip, device, gpiocdev.AsOutput(value))
	//if err != nil {
	//	return nil, fmt.Errorf("failed to request Pin line: %w", err)
	//}
	return &Pin{
		//line: line,
		chip:      chip,
		device:    device,
		direction: DirectionOutput,
	}, nil
}

//	func InitInput(chip string, device int) (*Pin, error) {
//		line, err := gpiocdev.RequestLine(chip, device, gpiocdev.AsInput)
//		if err != nil {
//			return nil, fmt.Errorf("failed to request Pin line: %w", err)
//		}
//		return &Pin{line: line, chip: chip, device: device}, nil
//	}
func InitInputDebounce(chip string, device int, deboucePeriod time.Duration) (*Pin, error) {
	//line, err := gpiocdev.RequestLine(chip, device, gpiocdev.AsInput, gpiocdev.WithDebounce(period))
	//if err != nil {
	//	return nil, fmt.Errorf("failed to request Pin line: %w", err)
	//}
	return &Pin{
		//line: line,
		chip:      chip,
		device:    device,
		direction: DirectionInput,
	}, nil
}

func (gpio *Pin) SetValue(value bool) {
	// gpio.line.SetValue(value)
}

func (gpio *Pin) GetValue() (bool, error) {
	if gpio.direction != DirectionInput {
		return false, fmt.Errorf("failed to get value, pin: `%w` is not an input pin", gpio)
	}
	//return gpio.line.Value()
	return true, nil
}

func (gpio *Pin) Deinit() error {
	return nil
	//return gpio.line.Close()
}
