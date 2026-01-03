package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jianlu8023/go-tools/v2/pkg/path"
)

// GenMysqlDSN 生成mysql的dsn
// @return string dsn
func (d *DataSourceConfig) GenMysqlDSN() string {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	// 如果需要正确处理time.Time 需要携带parseTime参数
	// 需要支持完整utf-8 需要设置charset=utf8mb4
	// 格式 "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
	// 添加TLS配置
	if d.TLSEnabled {
		// 使用自定义TLS配置
		dsn += "&tls=custom"
	}
	return dsn
}

// GenTiDBDSN 生成tidb的dsn
// @return string dsn
func (d *DataSourceConfig) GenTiDBDSN() string {
	// 格式可用mysql
	return d.GenMysqlDSN()
}

// GenPostgresDSN 生成postgres的dsn
// @return string dsn
func (d *DataSourceConfig) GenPostgresDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.UserName, d.Password, d.DataBaseName, d.Port)
}

// GenSqlite3DSN 生成sqlite3的dsn
// @return string dsn
func (d *DataSourceConfig) GenSqlite3DSN() string {
	// 格式 test.db?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL
	// test.db 数据库名称
	// _pragma=journal_mode=WAL 设置wal模式
	// _pragma=synchronous=NORMAL 设置同步模式 NORMAL 性能和数据安全之间平衡 FULL 最安全但最慢 OFF 最快但数据丢失风险最高
	// _pragma=busy_timeout=5000 设置超时时间
	if filepath.IsAbs(d.DataBasePath) {
		return fmt.Sprintf("%s?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL&charset=utf8mb4&parseTime=True&loc=Local",
			d.DataBasePath)
	} else {
		wd, _ := path.GetWorkDir()
		dsn := filepath.Clean(filepath.Join(wd, d.DataBasePath))
		dsnDir := filepath.Dir(dsn)
		if _, err := os.Stat(dsnDir); os.IsNotExist(err) {
			// 文件不存在
			if err := os.MkdirAll(dsnDir, os.FileMode(0o755)); err != nil {
				panic(err)
			}
		}
		return fmt.Sprintf("%s?_pragma=busy_timeout=5000&_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL&charset=utf8mb4&parseTime=True&loc=Local",
			dsn)
	}
}

// GenGaussDBDSN 生成GaussDB的dsn
// @return string dsn
func (d *DataSourceConfig) GenGaussDBDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=8000 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.UserName, d.Password, d.DataBaseName, d.Port)
}

// GenSqlServerDSN 生成sqlserver的dsn
// @return string dsn
func (d *DataSourceConfig) GenSqlServerDSN() string {
	// "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
	return fmt.Sprintf("sqlserver://%s:%s@%s:%v?database=%s",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}

// GenClickhouseDSN 生成clickhouse的dsn
// @return string dsn
func (d *DataSourceConfig) GenClickhouseDSN() string {
	// 格式 "clickhouse://gorm:gorm@localhost:9942/gorm?dial_timeout=10s&read_timeout=20s"
	return fmt.Sprintf("clickhouse://%s:%s@%s:%v/%s?dial_timeout=10s&read_timeout=20s",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}
