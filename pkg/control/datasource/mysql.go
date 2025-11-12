package datasource

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	mysqlori "github.com/go-sql-driver/mysql"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	// Mysql MySQL数据库类型标识
	Mysql = "mysql"
)

// newMysqlConn 创建MySQL数据库连接
// @description 根据配置创建MySQL数据库连接
// @param dbControl *Control 数据源控制器实例
// @return *gorm.DB GORM数据库连接实例
// @return error 创建连接过程中可能产生的错误
func newMysqlConn(dbControl *Control) (*gorm.DB, error) {
	// 如果启用了TLS，则配置TLS
	if dbControl.config.TLSEnabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: dbControl.config.TLSSkipVerify,
		}

		// 如果配置了CA证书，则加载CA证书
		if !stringer.IsBlank(dbControl.config.TLSCAFile) {
			rootCertPool := x509.NewCertPool()
			pem, err := os.ReadFile(dbControl.config.TLSCAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA file: %v", err)
			}
			if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
				return nil, fmt.Errorf("failed to append CA cert from PEM")
			}
			tlsConfig.RootCAs = rootCertPool
		}

		// 如果配置了客户端证书和私钥，则加载它们
		if !stringer.IsBlank(dbControl.config.TLSCertFile) &&
			!stringer.IsBlank(dbControl.config.TLSKeyFile) {
			clientCert := make([]tls.Certificate, 0, 1)
			certs, err := tls.LoadX509KeyPair(dbControl.config.TLSCertFile, dbControl.config.TLSKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load client certificate: %v", err)
			}
			clientCert = append(clientCert, certs)
			tlsConfig.Certificates = clientCert
		}

		// 如果配置了服务器名称，则设置
		if !stringer.IsBlank(dbControl.config.TLSServerName) {
			tlsConfig.ServerName = dbControl.config.TLSServerName
		}

		// 注册TLS配置
		if err := mysqlori.RegisterTLSConfig("custom", tlsConfig); err != nil {
			return nil, err
		}
	}

	return gorm.Open(mysql.New(mysql.Config{
		DSN:                       dbControl.config.GenMysqlDSN(), // DSN data source name
		DefaultStringSize:         256,                            // string 类型字段的默认长度
		DisableDatetimePrecision:  true,                           // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,                           // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,                           // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false,                          // 根据当前 MySQL 版本自动配置
	}),
		&gorm.Config{
			PrepareStmt:          true,
			DisableAutomaticPing: false,
			CreateBatchSize:      1000,
			Logger:               dbControl.dbLogger,
		})
}
