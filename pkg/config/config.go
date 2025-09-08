package config

import (
	"encoding/json"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/wd"
	"path/filepath"
)

type DataSourceConfig struct {
	DataSourceType string `json:"data_source_type,omitempty" yaml:"data_source_type,omitempty" mapstructure:"data_source_type"` // 数据源类型 mysql postgres sqlite3
	UserName       string `json:"db_username,omitempty" yaml:"db_username,omitempty" mapstructure:"db_username"`                // 数据库用户名
	Password       string `json:"db_password,omitempty" yaml:"db_password,omitempty" mapstructure:"db_password"`                // 数据库密码
	Host           string `json:"db_host,omitempty" yaml:"db_host,omitempty" mapstructure:"db_host"`                            // 数据库地址
	Port           int    `json:"db_port,omitempty" yaml:"db_port,omitempty" mapstructure:"db_port"`                            // 数据库端口
	DataBaseName   string `json:"db_name,omitempty" yaml:"db_name,omitempty" mapstructure:"db_name"`                            // 数据库名称
	DataBasePath   string `json:"db_path,omitempty" yaml:"db_path,omitempty" mapstructure:"db_path"`                            // sqlite3 使用
	MaxIdleConn    int    `json:"max_idle_conn,omitempty" yaml:"max_idle_conn,omitempty" mapstructure:"max_idle_conn"`          // 最大空闲连接数
	MaxOpenConn    int    `json:"max_open_conn,omitempty" yaml:"max_open_conn,omitempty" mapstructure:"max_open_conn"`          // 最大连接数
}

// GenMysqlDSN 生成mysql的dsn
func (d DataSourceConfig) GenMysqlDSN() string {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	// 如果需要正确处理time.Time 需要携带parseTime参数
	// 需要支持完整utf-8 需要设置charset=utf8mb4
	// 格式 "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}

// GenTiDBDSN 生成tidb的dsn
func (d DataSourceConfig) GenTiDBDSN() string {
	// 格式可用mysql
	return d.GenMysqlDSN()
}

// GenPostgresDSN 生成postgres的dsn
func (d DataSourceConfig) GenPostgresDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.UserName, d.Password, d.DataBaseName, d.Port)
}

// GenSqlite3DSN 生成sqlite3的dsn
func (d DataSourceConfig) GenSqlite3DSN() string {
	// 格式 test.db
	if filepath.IsAbs(d.DataBasePath) {
		return d.DataBasePath
	} else {
		return filepath.Clean(filepath.Join(wd.GetWorkDir(), d.DataBasePath))
	}
}

// GenGaussDBDSN 生成GaussDB的dsn
func (d DataSourceConfig) GenGaussDBDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=8000 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.UserName, d.Password, d.DataBaseName, d.Port)
}

// GenSqlServerDSN 生成sqlserver的dsn
func (d DataSourceConfig) GenSqlServerDSN() string {
	// "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
	return fmt.Sprintf("sqlserver://%s:%s@%s:%v?database=%s",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}

// GenClickhouseDSN 生成clickhouse的dsn
func (d DataSourceConfig) GenClickhouseDSN() string {
	// 格式 "clickhouse://gorm:gorm@localhost:9942/gorm?dial_timeout=10s&read_timeout=20s"
	return fmt.Sprintf("clickhouse://%s:%s@%s:%v/%s?dial_timeout=10s&read_timeout=20s",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}

func (d DataSourceConfig) String() string {
	bytes, _ := json.MarshalIndent(d, "", " ")
	return string(bytes)
}

type LoggerConfig struct {
	DefaultLogLevel string            `json:"default_log_level,omitempty" yaml:"default_log_level,omitempty" mapstructure:"default_log_level"`
	StackLogLevel   string            `json:"stack_log_level,omitempty" yaml:"stack_log_level,omitempty" mapstructure:"stack_log_level"`
	PrintFormat     string            `json:"print_format,omitempty" yaml:"print_format,omitempty" mapstructure:"print_format"`
	FilePath        string            `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
	MaxAge          int               `json:"max_age,omitempty" yaml:"max_age,omitempty" mapstructure:"max_age"`
	RotationTime    int               `json:"rotation_time,omitempty" yaml:"rotation_time,omitempty" mapstructure:"rotation_time"`
	LoggerLevel     map[string]string `json:"logger_level,omitempty" yaml:"logger_level,omitempty" mapstructure:"logger_level"`
}

func (l LoggerConfig) String() string {
	bytes, _ := json.MarshalIndent(l, "", " ")
	return string(bytes)
}

type HttpServerConfig struct {
	Address     string `yaml:"address,omitempty" json:"address,omitempty" mapstructure:"address" `
	ContextPath string `yaml:"context_path,omitempty" json:"context_path,omitempty" mapstructure:"context_path"`
	RunMode     string `yaml:"run_mode,omitempty" json:"run_mode,omitempty" mapstructure:"run_mode" `
	TlsEnabled  bool   `yaml:"tls_enabled,omitempty" json:"tls_enabled,omitempty" mapstructure:"tls_enabled" `
	TlsCertFile string `yaml:"tls_cert_file,omitempty" json:"tls_cert_file,omitempty" mapstructure:"tls_cert_file" `
	TlsKeyFile  string `yaml:"tls_key_file,omitempty" json:"tls_key_file,omitempty" mapstructure:"tls_key_file"`
}

func (h HttpServerConfig) String() string {
	bytes, _ := json.MarshalIndent(h, "", " ")
	return string(bytes)
}

type GrpcServerConfig struct {
	Host           string `yaml:"host,omitempty" json:"host,omitempty" mapstructure:"host" `
	MaxRecvMsgSize int    `yaml:"max_recv_msg_size,omitempty" json:"max_recv_msg_size,omitempty" mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int    `yaml:"max_send_msg_size,omitempty" json:"max_send_msg_size,omitempty" mapstructure:"max_send_msg_size"`
	ChunkSize      int    `yaml:"chunk_size,omitempty" json:"chunk_size,omitempty" mapstructure:"chunk_size"`
	TlsEnabled     bool   `yaml:"tls_enabled,omitempty" json:"tls_enabled,omitempty" mapstructure:"tls_enabled"`
	TlsCertFile    string `yaml:"tls_cert_file,omitempty" json:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`
	TlsKeyFile     string `yaml:"tls_key_file,omitempty" json:"tls_key_file,omitempty" mapstructure:"tls_key_file"`
	TlsRCACertFile string `yaml:"tls_rca_cert_file,omitempty" json:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"`
}

func (g GrpcServerConfig) String() string {
	bytes, _ := json.MarshalIndent(g, "", " ")
	return string(bytes)
}

type GrpcClientConfig struct {
	Host               string `yaml:"host,omitempty" json:"host,omitempty" mapstructure:"host"`
	MaxCallRecvMsgSize int    `yaml:"max_call_recv_msg_size,omitempty" json:"max_call_recv_msg_size,omitempty" mapstructure:"max_call_recv_msg_size"`
	MaxCallSendMsgSize int    `yaml:"max_call_send_msg_size,omitempty" json:"max_call_send_msg_size,omitempty" mapstructure:"max_call_send_msg_size" `
	ChunkSize          int    `yaml:"chunk_size,omitempty" json:"chunk_size,omitempty" mapstructure:"chunk_size" `
	CallTimeout        int    `yaml:"call_timeout,omitempty" json:"call_timeout,omitempty" mapstructure:"call_timeout" `
	TlsEnabled         bool   `yaml:"tls_enabled,omitempty" json:"tls_enabled,omitempty" mapstructure:"tls_enabled"`
	TlsCertFile        string `yaml:"tls_cert_file,omitempty" json:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`
	TlsKeyFile         string `yaml:"tls_key_file,omitempty" json:"tls_key_file,omitempty" mapstructure:"tls_key_file"`
	TlsRCACertFile     string `yaml:"tls_rca_cert_file,omitempty" json:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"`
}

func (g GrpcClientConfig) String() string {
	bytes, _ := json.MarshalIndent(g, "", " ")
	return string(bytes)
}

type GrpcConfig struct {
	Server *GrpcServerConfig `yaml:"server,omitempty" json:"server,omitempty" mapstructure:"server"`
	Client *GrpcClientConfig `yaml:"client,omitempty" json:"client,omitempty" mapstructure:"client"`
}

func (g GrpcConfig) String() string {
	bytes, _ := json.MarshalIndent(g, "", " ")
	return string(bytes)
}

type Libp2pConfig struct {
	ListenAddr    []string `json:"listen_addr,omitempty" yaml:"listen_addr,omitempty" mapstructure:"listen_addr"`
	ProtocolID    string   `json:"protocol_id,omitempty" yaml:"protocol_id,omitempty" mapstructure:"protocol_id"`
	ServiceTag    string   `json:"service_tag,omitempty" yaml:"service_tag,omitempty" mapstructure:"service_tag"`
	BootstrapList []string `json:"bootstrap_list,omitempty" yaml:"bootstrap_list,omitempty" mapstructure:"bootstrap_list"`
}

func (l Libp2pConfig) String() string {
	bytes, _ := json.MarshalIndent(l, "", " ")
	return string(bytes)
}

type Config struct {
	GrpcConfig       *GrpcConfig       `json:"grpc_config,omitempty" yaml:"grpc_config,omitempty" mapstructure:"grpc"`
	LoggerConfig     *LoggerConfig     `json:"logger_config,omitempty" yaml:"logger_config,omitempty" mapstructure:"logger"`
	HttpConfig       *HttpServerConfig `json:"http_config,omitempty" yaml:"http_config,omitempty" mapstructure:"http"`
	Libp2pConfig     *Libp2pConfig     `json:"libp2p_config,omitempty" yaml:"libp2p_config,omitempty" mapstructure:"libp2p"`
	DataSourceConfig *DataSourceConfig `json:"datasource_config,omitempty" yaml:"datasource_config,omitempty" mapstructure:"datasource"`
}

func (c Config) String() string {
	bytes, _ := json.MarshalIndent(c, "", " ")
	return string(bytes)
}
