package config

type Pin struct {
	Chip        string `yaml:"chip"`
	Channel     int    `yaml:"channel"`
	ActiveLevel bool   `yaml:"active_level"`
}

type Gpio struct {
	DoorRelay Pin `yaml:"door_relay"`
	OpenBtn   Pin `yaml:"open_btn"`
}
