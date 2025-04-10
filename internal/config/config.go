package config

import (
	"github.com/spf13/viper"
	"log"
	"paymentSystem/internal/models"
)

type Config struct {
	Server   models.Server
	Database models.Database
	LevelLog string `yaml:"level_log"`
}

func GetConfig() *Config {
	configPath := "../."
	viper.AddConfigPath(configPath)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("error reading config file", "error", err)
	}
	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal("error reading config file", "error", err)
	}
	return &cfg
}
