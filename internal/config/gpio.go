package config

//hardware:
//  door_relay:
//    - chip: "gpiochip0"
//      channel: 2
//      active_level: false
//  open_btn:
//    - chip: "gpiochip0"
//      channel: 3
//      active_level: false

type Pin struct {
	Chip        string `yaml:"chip"`
	Channel     int    `yaml:"channel"`
	ActiveLevel bool   `yaml:"active_level"`
}

type Gpio struct {
	DoorRelay Pin `yaml:"door_relay"`
	OpenBtn   Pin `yaml:"open_btn"`
}
