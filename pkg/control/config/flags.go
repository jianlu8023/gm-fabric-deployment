package config

import (
	"flag"
)

var _defaultConfig = "configs/default.yaml"

var (
	configPath string
	configType string
)

func init() {
	flag.StringVar(&configPath, "config", _defaultConfig, "config file path: configs/default.yaml")
	flag.StringVar(&configType, "type", "", "config file type: dev or prod")
	flag.Parse()
}
