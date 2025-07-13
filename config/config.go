package config

import (
	"log"

	"github.com/spf13/viper"
)

func LoadConfig() Config {
	var cfg Config

	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("error read config file: %s", err)
	}

	if err := viper.Unmarshal(&cfg.App); err != nil {
		log.Fatalf("error unmarshal app config: %s", err)
	}

	if err := viper.Unmarshal(&cfg.Database); err != nil {
		log.Fatalf("error unmarshal database config: %s", err)
	}

	if err := viper.Unmarshal(&cfg.Redis); err != nil {
		log.Fatalf("error unmarshal redis config: %s", err)
	}

	if err := viper.Unmarshal(&cfg.Product); err != nil {
		log.Fatalf("error unmarshal product config: %s", err)
	}

	if err := viper.Unmarshal(&cfg.Kafka); err != nil {
		log.Fatalf("error unmarshal product config: %s", err)
	}

	return cfg
}
