package config

import "os"

type Config struct {
	WebAppAddr string
}

var Default Config

func init() {
	Default = Config{
		WebAppAddr: os.Getenv("WEBAPP_ADDR"),
	}
}
