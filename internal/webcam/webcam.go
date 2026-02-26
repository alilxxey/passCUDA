package webcam

import (
	"os"
	"os/exec"
)

type Webcam struct {
	device string
}

func Init(device string) (*Webcam, error) {
	return &Webcam{device: device}, nil
}

func (w *Webcam) GetPhoto(resolution string, delay int) (string, error) {
	f, err := os.CreateTemp("", "passCUDAphoto")
	if err != nil {
		return "", err
	}

	cmd := exec.Command("fswebcam", "--device", w.device, "-r", resolution, "--jpeg", "100", "-D", string(delay), f.Name())
	if err := cmd.Run(); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}
