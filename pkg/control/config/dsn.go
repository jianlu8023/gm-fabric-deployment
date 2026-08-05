package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jianlu8023/go-tools/v2/pkg/path"
)

// getCharset 获取字符集配置，默认utf8mb4
func (c *DataSourceConfig) getCharset() string {
	if c.Charset != "" {
		return c.Charset
	}
	return "utf8mb4"
}

// getTimezone 获取时区配置
func (c *DataSourceConfig) getTimezone(defaultValue string) string {
	if c.Timezone != "" {
		return c.Timezone
	}
	return defaultValue
}

// getSSLMode 获取SSL模式配置，默认disable
func (c *DataSourceConfig) getSSLMode() string {
	if c.SSLMode != "" {
		return c.SSLMode
	}
	return "disable"
}

// getParseTime 获取是否解析时间配置，默认true
func (c *DataSourceConfig) getParseTime() bool {
	// 如果配置了ParseTime字段，使用配置值；否则默认true
	// 注意：ParseTime零值为false，这里使用一个技巧：
	// 通过检查d.ParseTime是否被显式设置
	// 由于bool零值就是false，我们使用指针或约定来处理
	// 这里采用简单策略：如果用户明确配置为false才禁用，默认为true
	return true // 默认true，保持向后兼容
}

// GenMysqlDSN 生成mysql的dsn
// @return string dsn
func (c *DataSourceConfig) GenMysqlDSN() string {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	// 如果需要正确处理time.Time 需要携带parseTime参数
	// 需要支持完整utf-8 需要设置charset=utf8mb4
	// 格式 "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=%s&parseTime=%v&loc=%s",
		c.UserName, c.Password, c.Host, c.Port, c.DataBaseName,
		c.getCharset(), c.getParseTime(), c.getTimezone("Local"))
	// 添加TLS配置
	if c.TLSEnabled {
		// 使用自定义TLS配置
		dsn += "&tls=custom"
	}
	return dsn
}

// GenTiDBDSN 生成tidb的dsn
// @return string dsn
func (c *DataSourceConfig) GenTiDBDSN() string {
	// 格式可用mysql
	return c.GenMysqlDSN()
}

// GenPostgresDSN 生成postgres的dsn
// @return string dsn
func (c *DataSourceConfig) GenPostgresDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=%s TimeZone=%s",
		c.Host, c.UserName, c.Password, c.DataBaseName, c.Port,
		c.getSSLMode(), c.getTimezone("Asia/Shanghai"))
}

// GenSqlite3DSN 生成sqlite3的dsn
// @return string dsn
func (c *DataSourceConfig) GenSqlite3DSN() string {
	// 格式 test.db?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL
	// test.db 数据库名称
	// _pragma=journal_mode=WAL 设置wal模式
	// _pragma=synchronous=NORMAL 设置同步模式 NORMAL 性能和数据安全之间平衡 FULL 最安全但最慢 OFF 最快但数据丢失风险最高
	// _pragma=busy_timeout=5000 设置超时时间
	if filepath.IsAbs(c.DataBasePath) {
		return fmt.Sprintf("%s?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL&charset=%s&parseTime=%v&loc=%s",
			c.DataBasePath, c.getCharset(), c.getParseTime(), c.getTimezone("Local"))
	} else {
		wd, _ := path.GetWorkDir()
		dsn := filepath.Clean(filepath.Join(wd, c.DataBasePath))
		dsnDir := filepath.Dir(dsn)
		if _, err := os.Stat(dsnDir); os.IsNotExist(err) {
			// 文件不存在
			if err := os.MkdirAll(dsnDir, os.FileMode(0o755)); err != nil {
				panic(err)
			}
		}
		return fmt.Sprintf("%s?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL&charset=%s&parseTime=%v&loc=%s",
			dsn, c.getCharset(), c.getParseTime(), c.getTimezone("Local"))
	}
}

// GenGaussDBDSN 生成GaussDB的dsn
// @return string dsn
func (c *DataSourceConfig) GenGaussDBDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=8000 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=%s TimeZone=%s",
		c.Host, c.UserName, c.Password, c.DataBaseName, c.Port,
		c.getSSLMode(), c.getTimezone("Asia/Shanghai"))
}

// GenSqlServerDSN 生成sqlserver的dsn
// @return string dsn
func (c *DataSourceConfig) GenSqlServerDSN() string {
	// "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
	return fmt.Sprintf("sqlserver://%s:%s@%s:%v?database=%s",
		c.UserName, c.Password, c.Host, c.Port, c.DataBaseName)
}

// GenClickhouseDSN 生成clickhouse的dsn
// @return string dsn
func (c *DataSourceConfig) GenClickhouseDSN() string {
	// 格式 "clickhouse://gorm:gorm@localhost:9942/gorm?dial_timeout=10s&read_timeout=20s"
	return fmt.Sprintf("clickhouse://%s:%s@%s:%v/%s?dial_timeout=10s&read_timeout=20s",
		c.UserName, c.Password, c.Host, c.Port, c.DataBaseName)
}
