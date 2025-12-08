package geoip

import (
	"fmt"
	"net"
	"sync"
	"sync/atomic"

	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/http"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/oschwald/geoip2-golang"
	"go.uber.org/zap"
)

// Control GeoIP控制器结构体
type Control struct {
	config  *config.GeoIPConfig
	logger  *zap.SugaredLogger
	db      *geoip2.Reader
	once    sync.Once
	started atomic.Bool
}

// NewGeoIPControl 创建一个新的GeoIP控制器
//
// @param geoIPConfig *config.GeoIPConfig GeoIP配置
// @param loggerControl *logger.Control 日志控制器
//
// @return *Control GeoIP控制器
func NewGeoIPControl(geoIPConfig *config.GeoIPConfig, loggerControl *logger.Control) *Control {
	geoIPConfig = check.IF[*config.GeoIPConfig](geoIPConfig == nil,
		getDefaultConfig(),
		geoIPConfig,
	)

	if !geoIPConfig.Enabled {
		return nil
	}

	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)

	geoIPLogger := loggerControl.GenLogger(logger.ModuleGeoIP)
	geoIPLogger.Infof("[control] starting new geoip control...")

	control := &Control{
		config: geoIPConfig,
		logger: geoIPLogger,
	}

	return control
}

// StartUp 启动GeoIP服务
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Debugf("[control] starting geoip server...")

			// 如果启用了自动下载并且配置了下载URL，则尝试下载数据库
			if c.config.AutoDownload && !stringer.IsBlank(c.config.DownloadURL) {
				if err := c.downloadDatabase(); err != nil {
					c.logger.Errorf("[control] failed to download geoip database: %v", err)
					if failedFunc != nil {
						failedFunc(err)
					}
					return
				}
			}

			// 打开GeoIP数据库
			if !stringer.IsBlank(c.config.DatabasePath) {
				db, err := geoip2.Open(c.config.DatabasePath)
				if err != nil {
					c.logger.Errorf("[control] failed to open geoip database: %v", err)
					if failedFunc != nil {
						failedFunc(err)
					}
					return
				}
				c.db = db
				c.started.Store(true)
				c.logger.Debugf("[control] geoip database loaded successfully")
			} else {
				err := fmt.Errorf("database path is empty")
				c.logger.Errorf("[control] %v", err)
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
		}
	})
}

// Shutdown 关闭GeoIP服务
func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutting down geoip server...")
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			c.logger.Errorf("[control] failed to close geoip database: %v", err)
			return err
		}
	}
	c.started.Store(false)
	return nil
}

// downloadDatabase 从URL下载GeoIP数据库文件
func (c *Control) downloadDatabase() error {
	c.logger.Debugf("[control] downloading geoip database from %s", c.config.DownloadURL)

	// 使用封装的HTTP客户端下载文件
	client := http.NewClient()
	_, err := client.DownloadFile(c.config.DownloadURL, c.config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to download geoip database: %w", err)
	}

	c.logger.Debugf("[control] geoip database downloaded successfully to %s", c.config.DatabasePath)
	return nil
}

// GetIPInfo 获取指定IP地址的详细地理位置信息
func (c *Control) GetIPInfo(ipStr string) (*IPInfo, error) {
	if !c.started.Load() || c.db == nil {
		return nil, fmt.Errorf("geoip service not started or database not loaded")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	record, err := c.db.City(ip)
	if err != nil {
		return nil, fmt.Errorf("failed to get city record for IP %s: %w", ipStr, err)
	}

	info := &IPInfo{
		IP: ipStr,
		Country: CountryInfo{
			Code: record.Country.IsoCode,
			Name: record.Country.Names["en"],
		},
		City: CityInfo{
			Name: record.City.Names["en"],
		},
		Location: LocationInfo{
			Latitude:  record.Location.Latitude,
			Longitude: record.Location.Longitude,
			TimeZone:  record.Location.TimeZone,
		},
		Postal:    record.Postal.Code,
		Continent: record.Continent.Names["en"],
	}

	return info, nil
}

// GetCountryCode 获取指定IP地址的国家代码
func (c *Control) GetCountryCode(ipStr string) (string, error) {
	if !c.started.Load() || c.db == nil {
		return "", fmt.Errorf("geoip service not started or database not loaded")
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", ipStr)
	}

	record, err := c.db.Country(ip)
	if err != nil {
		return "", fmt.Errorf("failed to get country record for IP %s: %w", ipStr, err)
	}

	return record.Country.IsoCode, nil
}

// IsStarted 检查GeoIP服务是否已启动
func (c *Control) IsStarted() bool {
	return c.started.Load()
}
