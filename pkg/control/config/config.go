package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/system/wd"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"os"
	"path/filepath"
)

// DockerConfig docker配置结构体
type DockerConfig struct {
	Host           string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"`                                  // host地址
	APIVersion     string `json:"api_version,omitempty" yaml:"api_version,omitempty" mapstructure:"api_version"`             // docker api版本
	TlsEnabled     bool   `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled"`             // 是否启用tls
	TlsCertFile    string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`       // tls证书文件
	TlsKeyFile     string `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`          // tls私钥文件
	TlsCAFile      string `json:"tls_ca_file,omitempty" yaml:"tls_ca_file,omitempty" mapstructure:"tls_ca_file"`             // tls ca文件
	DefaultTimeout int    `json:"default_timeout,omitempty" yaml:"default_timeout,omitempty" mapstructure:"default_timeout"` // 默认超时时间
}

// String 返回DockerConfig的字符串表示
// @return string DockerConfig的字符串表示
func (d DockerConfig) String() string {
	bytes, _ := json.MarshalIndent(d, "", " ")
	return string(bytes)
}

// DataSourceConfig 数据源配置结构体
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
	LogInConsole   bool   `json:"log_in_console,omitempty" yaml:"log_in_console,omitempty" mapstructure:"log_in_console"`       // 是否在控制台打印日志
}

// GenMysqlDSN 生成mysql的dsn
// @return string dsn
func (d DataSourceConfig) GenMysqlDSN() string {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	// 如果需要正确处理time.Time 需要携带parseTime参数
	// 需要支持完整utf-8 需要设置charset=utf8mb4
	// 格式 "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}

// GenTiDBDSN 生成tidb的dsn
// @return string dsn
func (d DataSourceConfig) GenTiDBDSN() string {
	// 格式可用mysql
	return d.GenMysqlDSN()
}

// GenPostgresDSN 生成postgres的dsn
// @return string dsn
func (d DataSourceConfig) GenPostgresDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.UserName, d.Password, d.DataBaseName, d.Port)
}

// GenSqlite3DSN 生成sqlite3的dsn
// @return string dsn
func (d DataSourceConfig) GenSqlite3DSN() string {
	// 格式 test.db
	if filepath.IsAbs(d.DataBasePath) {
		return d.DataBasePath
	} else {
		dsn := filepath.Clean(filepath.Join(wd.GetWorkDir(), d.DataBasePath))
		dsnDir := filepath.Dir(dsn)
		if _, err := os.Stat(dsnDir); os.IsNotExist(err) {
			// 文件不存在
			if err := os.MkdirAll(dsnDir, os.FileMode(0755)); err != nil {
				panic(err)
			}
		}
		return dsn
	}
}

// GenGaussDBDSN 生成GaussDB的dsn
// @return string dsn
func (d DataSourceConfig) GenGaussDBDSN() string {
	// 格式 "host=localhost user=gorm password=gorm dbname=gorm port=8000 sslmode=disable TimeZone=Asia/Shanghai"
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%v sslmode=disable TimeZone=Asia/Shanghai",
		d.Host, d.UserName, d.Password, d.DataBaseName, d.Port)
}

// GenSqlServerDSN 生成sqlserver的dsn
// @return string dsn
func (d DataSourceConfig) GenSqlServerDSN() string {
	// "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
	return fmt.Sprintf("sqlserver://%s:%s@%s:%v?database=%s",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}

// GenClickhouseDSN 生成clickhouse的dsn
// @return string dsn
func (d DataSourceConfig) GenClickhouseDSN() string {
	// 格式 "clickhouse://gorm:gorm@localhost:9942/gorm?dial_timeout=10s&read_timeout=20s"
	return fmt.Sprintf("clickhouse://%s:%s@%s:%v/%s?dial_timeout=10s&read_timeout=20s",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
}

// String 返回DataSourceConfig的字符串表示
// @return string 字符串表示
func (d DataSourceConfig) String() string {
	bytes, _ := json.MarshalIndent(d, "", " ")
	return string(bytes)
}

// LoggerConfig 日志配置
type LoggerConfig struct {
	DefaultLogLevel string            `json:"default_log_level,omitempty" yaml:"default_log_level,omitempty" mapstructure:"default_log_level"` // 默认日志级别
	StackLogLevel   string            `json:"stack_log_level,omitempty" yaml:"stack_log_level,omitempty" mapstructure:"stack_log_level"`       // 堆栈日志级别
	PrintFormat     string            `json:"print_format,omitempty" yaml:"print_format,omitempty" mapstructure:"print_format"`                // 日志格式
	FilePath        string            `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`                         // 日志文件路径
	MaxAge          int               `json:"max_age,omitempty" yaml:"max_age,omitempty" mapstructure:"max_age"`                               // 日志文件最大保留时间 (天)
	RotationTime    int               `json:"rotation_time,omitempty" yaml:"rotation_time,omitempty" mapstructure:"rotation_time"`             // 日志文件轮转时间 (小时)
	LoggerLevel     map[string]string `json:"logger_level,omitempty" yaml:"logger_level,omitempty" mapstructure:"logger_level"`                // 日志级别
}

// String 返回配置的字符串表示
// @return string 配置的字符串表示
func (l LoggerConfig) String() string {
	bytes, _ := json.MarshalIndent(l, "", " ")
	return string(bytes)
}

// HttpServerConfig http服务配置
type HttpServerConfig struct {
	Address     string `yaml:"address,omitempty" json:"address,omitempty" mapstructure:"address"`                    // 服务地址
	ContextPath string `yaml:"context_path,omitempty" json:"context_path,omitempty" mapstructure:"context_path"`     // 服务上下文路径
	RunMode     string `yaml:"run_mode,omitempty" json:"run_mode,omitempty" mapstructure:"run_mode" `                // 服务运行模式
	TlsEnabled  bool   `yaml:"tls_enabled,omitempty" json:"tls_enabled,omitempty" mapstructure:"tls_enabled" `       // 是否启用TLS
	TlsCertFile string `yaml:"tls_cert_file,omitempty" json:"tls_cert_file,omitempty" mapstructure:"tls_cert_file" ` // TLS证书文件
	TlsKeyFile  string `yaml:"tls_key_file,omitempty" json:"tls_key_file,omitempty" mapstructure:"tls_key_file"`     // TLS私钥文件
}

// String 返回配置的字符串表示
// @return string 配置的字符串表示
func (h HttpServerConfig) String() string {
	bytes, _ := json.MarshalIndent(h, "", " ")
	return string(bytes)
}

// GrpcServerConfig grpc服务配置
type GrpcServerConfig struct {
	Host           string `yaml:"host,omitempty" json:"host,omitempty" mapstructure:"host"`                                        // 服务地址
	MaxRecvMsgSize int    `yaml:"max_recv_msg_size,omitempty" json:"max_recv_msg_size,omitempty" mapstructure:"max_recv_msg_size"` // 最大接收消息大小
	MaxSendMsgSize int    `yaml:"max_send_msg_size,omitempty" json:"max_send_msg_size,omitempty" mapstructure:"max_send_msg_size"` // 最大发送消息大小
	ChunkSize      int    `yaml:"chunk_size,omitempty" json:"chunk_size,omitempty" mapstructure:"chunk_size"`                      // 分块大小
	TlsEnabled     bool   `yaml:"tls_enabled,omitempty" json:"tls_enabled,omitempty" mapstructure:"tls_enabled"`                   // 是否启用TLS
	TlsCertFile    string `yaml:"tls_cert_file,omitempty" json:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`             // TLS证书文件
	TlsKeyFile     string `yaml:"tls_key_file,omitempty" json:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                // TLS私钥文件
	TlsRCACertFile string `yaml:"tls_rca_cert_file,omitempty" json:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"` // TLS根证书文件
}

// String 返回GrpcServerConfig的JSON格式字符串
// @return string 返回GrpcServerConfig的JSON格式字符串
func (g GrpcServerConfig) String() string {
	bytes, _ := json.MarshalIndent(g, "", " ")
	return string(bytes)
}

// GrpcClientConfig 配置GrpcClient
type GrpcClientConfig struct {
	Host               string `yaml:"host,omitempty" json:"host,omitempty" mapstructure:"host"`                                                        // Grpc服务地址
	MaxCallRecvMsgSize int    `yaml:"max_call_recv_msg_size,omitempty" json:"max_call_recv_msg_size,omitempty" mapstructure:"max_call_recv_msg_size"`  // 最大接收消息大小
	MaxCallSendMsgSize int    `yaml:"max_call_send_msg_size,omitempty" json:"max_call_send_msg_size,omitempty" mapstructure:"max_call_send_msg_size" ` // 最大发送消息大小
	ChunkSize          int    `yaml:"chunk_size,omitempty" json:"chunk_size,omitempty" mapstructure:"chunk_size"`                                      // 分块大小
	CallTimeout        int    `yaml:"call_timeout,omitempty" json:"call_timeout,omitempty" mapstructure:"call_timeout" `                               // 调用超时时间
	TlsEnabled         bool   `yaml:"tls_enabled,omitempty" json:"tls_enabled,omitempty" mapstructure:"tls_enabled"`                                   // 是否启用TLS
	TlsCertFile        string `yaml:"tls_cert_file,omitempty" json:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`                             // TLS证书文件
	TlsKeyFile         string `yaml:"tls_key_file,omitempty" json:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                                // TLS私钥文件
	TlsRCACertFile     string `yaml:"tls_rca_cert_file,omitempty" json:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"`                 // TLS根证书文件
}

// String GrpcClientConfig的字符串表示
// @return string GrpcClientConfig的字符串表示
func (g GrpcClientConfig) String() string {
	bytes, _ := json.MarshalIndent(g, "", " ")
	return string(bytes)
}

// GrpcConfig 配置Grpc
type GrpcConfig struct {
	Server *GrpcServerConfig `yaml:"server,omitempty" json:"server,omitempty" mapstructure:"server"` // 服务端配置
	Client *GrpcClientConfig `yaml:"client,omitempty" json:"client,omitempty" mapstructure:"client"` // 客户端配置
}

// String GrpcConfig的字符串表示
// @return string GrpcConfig的字符串表示
func (g GrpcConfig) String() string {
	bytes, _ := json.MarshalIndent(g, "", " ")
	return string(bytes)
}

const (
	Ed25519 = "ed25519"
	Rsa     = "rsa"
)

// CreateIdentity 生成libp2p身份
// @param algorithm 身份算法
// @param rsaKeyLen rsa密钥长度
// @return Identity libp2p身份
// @return error 错误信息
func CreateIdentity(algorithm string, rsaKeyLen int) (Identity, error) {
	ident := Identity{}

	var sk crypto.PrivKey
	var pk crypto.PubKey

	switch algorithm {
	case Rsa:
		fmt.Printf("generate rsa key pair with key length: %d\n", rsaKeyLen)
		privK, pubK, err := crypto.GenerateKeyPair(crypto.RSA, rsaKeyLen)
		if err != nil {
			fmt.Printf("generate rsa key pair failed: %v\n", err)
			return ident, err
		}
		sk = privK
		pk = pubK
	case Ed25519:
		fmt.Println("generate ed25519 key pair")
		privK, pubK, err := crypto.GenerateEd25519Key(rand.Reader)
		if err != nil {
			fmt.Printf("generate ed25519 key pair failed: %v\n", err)
			return ident, err
		}
		sk = privK
		pk = pubK
	default:
		fmt.Println("algorithm no support...")
		return ident, errors.New("algorithm no support")
	}

	skBytes, err := crypto.MarshalPrivateKey(sk)
	if err != nil {
		fmt.Printf("marshal private key failed: %v\n", err)
		return ident, err
	}

	ident.PrivKey = base64.StdEncoding.EncodeToString(skBytes)
	peerId, err := peer.IDFromPublicKey(pk)
	if err != nil {
		fmt.Printf("generate peer id failed: %v\n", err)
		return ident, err
	}
	ident.PeerID = peerId.String()
	return ident, nil
}

// Identity 配置身份信息
type Identity struct {
	PeerID  string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" mapstructure:"peer_id"`
	PrivKey string `json:",omitempty" yaml:",omitempty" mapstructure:"privkey"`
}

// DecodePrivateKey is a helper to decode the users PrivateKey.
func (i *Identity) DecodePrivateKey(passphrase string) (crypto.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(i.PrivKey)
	if err != nil {
		return nil, err
	}

	// currently storing key unencrypted. in the future we need to encrypt it.
	// TODO(security)
	return crypto.UnmarshalPrivateKey(pkb)
}

// Libp2pConfig 配置Libp2p
type Libp2pConfig struct {
	ListenAddr    []string  `json:"listen_addr,omitempty" yaml:"listen_addr,omitempty" mapstructure:"listen_addr"`          // 监听地址
	Identity      *Identity `json:"identity,omitempty" yaml:"identity,omitempty" mapstructure:"identity"`                   // 身份信息
	ProtocolID    string    `json:"protocol_id,omitempty" yaml:"protocol_id,omitempty" mapstructure:"protocol_id"`          // 协议ID
	ServiceTag    string    `json:"service_tag,omitempty" yaml:"service_tag,omitempty" mapstructure:"service_tag"`          // mdns服务标签
	BootstrapList []string  `json:"bootstrap_list,omitempty" yaml:"bootstrap_list,omitempty" mapstructure:"bootstrap_list"` // bootstrap节点列表
}

// String Libp2pConfig的字符串表示
// @return string Libp2pConfig的字符串表示
func (l Libp2pConfig) String() string {
	bytes, _ := json.MarshalIndent(l, "", " ")
	return string(bytes)
}

// Config 配置
type Config struct {
	GrpcConfig       *GrpcConfig       `json:"grpc_config,omitempty" yaml:"grpc_config,omitempty" mapstructure:"grpc"`                   // grpc配置
	LoggerConfig     *LoggerConfig     `json:"logger_config,omitempty" yaml:"logger_config,omitempty" mapstructure:"logger"`             // logger配置
	HttpConfig       *HttpServerConfig `json:"http_config,omitempty" yaml:"http_config,omitempty" mapstructure:"http"`                   // http配置
	Libp2pConfig     *Libp2pConfig     `json:"libp2p_config,omitempty" yaml:"libp2p_config,omitempty" mapstructure:"libp2p"`             // libp2p配置
	DataSourceConfig *DataSourceConfig `json:"datasource_config,omitempty" yaml:"datasource_config,omitempty" mapstructure:"datasource"` // 数据源配置
	DockerConfig     *DockerConfig     `json:"docker_config,omitempty" yaml:"docker_config,omitempty" mapstructure:"docker"`             // docker配置
}

// String 返回配置的字符串表示
// @return string 配置的字符串表示
func (c Config) String() string {
	bytes, _ := json.MarshalIndent(c, "", " ")
	return string(bytes)
}
