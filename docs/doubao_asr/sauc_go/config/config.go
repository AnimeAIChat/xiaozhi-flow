package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Auth struct {
		AppKey    string `toml:"app_key"`
		AccessKey string `toml:"access_key"`
	}
}

var config Config

func init() {
	if _, err := toml.DecodeFile("../config.toml", &config); err != nil {
		log.Fatalf("Error reading TOML file: %v", err)
	}
}

func AppKey() string {
	return config.Auth.AppKey
}

func AccessKey() string {
	return config.Auth.AccessKey
}
