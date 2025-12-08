package geoip

import "github.com/jianlu8023/golang-example/pkg/control/config"

// getDefaultConfig 获取默认的GeoIP配置
func getDefaultConfig() *config.GeoIPConfig {
	return &config.GeoIPConfig{
		Enabled: false,
	}
}
