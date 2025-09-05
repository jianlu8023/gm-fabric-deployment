package config

import (
	"github.com/spf13/viper"
)

func LoadConfig(path string) Config {

	var cfg Config

	viper.SetConfigName(path)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)

	}
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
