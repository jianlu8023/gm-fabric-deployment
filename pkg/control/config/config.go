package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// DockerConfig docker配置结构体
type DockerConfig struct {
	Enabled        bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                         // 是否启用
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
func (d *DockerConfig) String() string {
	pretty, _ := json.MarshalPretty(d)
	return string(pretty)
}

// DataSourceConfig 数据源配置结构体
type DataSourceConfig struct {
	Enabled        bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                            // 是否启用
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
func (d *DataSourceConfig) GenMysqlDSN() string {
	// 参考 https://github.com/go-sql-driver/mysql#dsn-data-source-name 获取详情
	// 如果需要正确处理time.Time 需要携带parseTime参数
	// 需要支持完整utf-8 需要设置charset=utf8mb4
	// 格式 "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	return fmt.Sprintf("%s:%s@tcp(%s:%v)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.UserName, d.Password, d.Host, d.Port, d.DataBaseName)
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

// String 返回DataSourceConfig的字符串表示
// @return string 字符串表示
func (d *DataSourceConfig) String() string {
	pretty, _ := json.MarshalPretty(d)
	return string(pretty)
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
func (l *LoggerConfig) String() string {
	pretty, _ := json.MarshalPretty(l)
	return string(pretty)
}

// HttpServerConfig http服务配置
type HttpServerConfig struct {
	Enabled        bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                               // 是否启用
	Address        string `json:"address,omitempty" yaml:"address,omitempty" mapstructure:"address"`                               // 服务地址
	ContextPath    string `json:"context_path,omitempty" yaml:"context_path,omitempty" mapstructure:"context_path"`                // 服务上下文路径
	RunMode        string `json:"run_mode,omitempty" yaml:"run_mode,omitempty" mapstructure:"run_mode" `                           // 服务运行模式
	TlsEnabled     bool   `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled" `                  // 是否启用TLS
	TlsGM          bool   `json:"tls_gm,omitempty" yaml:"tls_gm,omitempty" mapstructure:"tls_gm"`                                  // 是否启用国密TLS
	TlsCertFile    string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file" `            // TLS证书文件
	TlsKeyFile     string `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                // TLS私钥文件
	TlsRCACertFile string `json:"tls_rca_cert_file,omitempty" yaml:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"` // TLS根证书文件
	Pprof          bool   `json:"pprof,omitempty" yaml:"pprof,omitempty" mapstructure:"pprof"`                                     // 是否启用pprof
	// 黑白名单配置
	IPWhiteList struct {
		Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用IP白名单
		IPs     []string `json:"ips,omitempty" yaml:"ips,omitempty" mapstructure:"ips"`             // IP白名单，优先级高于黑名单
	} `json:"ip_white_list,omitempty" yaml:"ip_white_list,omitempty" mapstructure:"ip_white_list"`
	IPBlackList struct {
		Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用IP黑名单
		IPs     []string `json:"ips,omitempty" yaml:"ips,omitempty" mapstructure:"ips"`             // IP黑名单
	} `json:"ip_black_list,omitempty" yaml:"ip_black_list,omitempty" mapstructure:"ip_black_list"`
	// 限流配置
	RateLimit struct {
		Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用限流
		RPS     int64  `json:"rps,omitempty" yaml:"rps,omitempty" mapstructure:"rps"`             // 每秒请求数限制
		Burst   int    `json:"burst,omitempty" yaml:"burst,omitempty" mapstructure:"burst"`       // 令牌桶突发大小
		Type    string `json:"type,omitempty" yaml:"type,omitempty" mapstructure:"type"`          // 限流类型，可选值：time, ulule, ants, custom, juju
	} `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty" mapstructure:"rate_limit"`
}

// String 返回配置的字符串表示
// @return string 配置的字符串表示
func (h *HttpServerConfig) String() string {
	pretty, _ := json.MarshalPretty(h)
	return string(pretty)
}

// GrpcServerConfig grpc服务配置
type GrpcServerConfig struct {
	Host           string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"`                                        // 服务地址
	MaxRecvMsgSize int    `json:"max_recv_msg_size,omitempty" yaml:"max_recv_msg_size,omitempty" mapstructure:"max_recv_msg_size"` // 最大接收消息大小
	MaxSendMsgSize int    `json:"max_send_msg_size,omitempty" yaml:"max_send_msg_size,omitempty" mapstructure:"max_send_msg_size"` // 最大发送消息大小
	ChunkSize      int    `json:"chunk_size,omitempty" yaml:"chunk_size,omitempty" mapstructure:"chunk_size"`                      // 分块大小
	TlsEnabled     bool   `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled"`                   // 是否启用TLS
	TlsCertFile    string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`             // TLS证书文件
	TlsKeyFile     string `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                // TLS私钥文件
	TlsRCACertFile string `json:"tls_rca_cert_file,omitempty" yaml:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"` // TLS根证书文件
}

// String 返回GrpcServerConfig的JSON格式字符串
// @return string 返回GrpcServerConfig的JSON格式字符串
func (g *GrpcServerConfig) String() string {
	pretty, _ := json.MarshalPretty(g)
	return string(pretty)
}

// GrpcClientConfig 配置GrpcClient
type GrpcClientConfig struct {
	Host               string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"`                                                        // Grpc服务地址
	MaxCallRecvMsgSize int    `json:"max_call_recv_msg_size,omitempty" yaml:"max_call_recv_msg_size,omitempty" mapstructure:"max_call_recv_msg_size"`  // 最大接收消息大小
	MaxCallSendMsgSize int    `json:"max_call_send_msg_size,omitempty" yaml:"max_call_send_msg_size,omitempty" mapstructure:"max_call_send_msg_size" ` // 最大发送消息大小
	ChunkSize          int    `json:"chunk_size,omitempty" yaml:"chunk_size,omitempty" mapstructure:"chunk_size"`                                      // 分块大小
	CallTimeout        int    `json:"call_timeout,omitempty" yaml:"call_timeout,omitempty" mapstructure:"call_timeout" `                               // 调用超时时间
	TlsEnabled         bool   `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled"`                                   // 是否启用TLS
	TlsCertFile        string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`                             // TLS证书文件
	TlsKeyFile         string `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                                // TLS私钥文件
	TlsRCACertFile     string `json:"tls_rca_cert_file,omitempty" yaml:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"`                 // TLS根证书文件
}

// String GrpcClientConfig的字符串表示
// @return string GrpcClientConfig的字符串表示
func (g *GrpcClientConfig) String() string {
	pretty, _ := json.MarshalPretty(g)
	return string(pretty)
}

// GrpcConfig 配置Grpc
type GrpcConfig struct {
	Enabled bool              `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用
	Server  *GrpcServerConfig `json:"server,omitempty" yaml:"server,omitempty" mapstructure:"server"`    // 服务端配置
	Client  *GrpcClientConfig `json:"client,omitempty" yaml:"client,omitempty" mapstructure:"client"`    // 客户端配置
}

// String GrpcConfig的字符串表示
// @return string GrpcConfig的字符串表示
func (g *GrpcConfig) String() string {
	pretty, _ := json.MarshalPretty(g)
	return string(pretty)
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
	PrivKey string `json:"-" yaml:",omitempty" mapstructure:"privkey"`
}

// String 返回Identity的字符串表示
// @return string Identity的字符串表示
func (i *Identity) String() string {
	pretty, _ := json.MarshalPretty(i)
	return string(pretty)
}

// DecodePrivateKey 解码用户的私钥
// @param passphrase string 私钥密码（当前版本未使用）
// @return crypto.PrivKey 解码后的私钥对象
// @return error 解码过程中可能产生的错误
//
// nolint: unused
func (i *Identity) DecodePrivateKey(passphrase string) (crypto.PrivKey, error) {
	pkb, err := base64.StdEncoding.DecodeString(i.PrivKey)
	if err != nil {
		return nil, err
	}

	return crypto.UnmarshalPrivateKey(pkb)
}

// Libp2pConfig 配置Libp2p
type Libp2pConfig struct {
	Enabled       bool      `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                      // 是否启用
	ListenAddr    []string  `json:"listen_addr,omitempty" yaml:"listen_addr,omitempty" mapstructure:"listen_addr"`          // 监听地址
	Identity      *Identity `json:"identity,omitempty" yaml:"identity,omitempty" mapstructure:"identity"`                   // 身份信息
	ProtocolID    string    `json:"protocol_id,omitempty" yaml:"protocol_id,omitempty" mapstructure:"protocol_id"`          // 协议ID
	ServiceTag    string    `json:"service_tag,omitempty" yaml:"service_tag,omitempty" mapstructure:"service_tag"`          // mdns服务标签
	BootstrapList []string  `json:"bootstrap_list,omitempty" yaml:"bootstrap_list,omitempty" mapstructure:"bootstrap_list"` // bootstrap节点列表
}

// String Libp2pConfig的字符串表示
// @return string Libp2pConfig的字符串表示
func (l *Libp2pConfig) String() string {
	pretty, _ := json.MarshalPretty(l)
	return string(pretty)
}

// CaptchaConfig 验证码配置结构体
type CaptchaConfig struct {
	Enabled     bool    `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                // 是否启用
	MaxAge      int     `json:"max_age,omitempty" yaml:"max_age,omitempty" mapstructure:"max_age"`                // 验证码有效期（秒）
	Width       int     `json:"width,omitempty" yaml:"width,omitempty" mapstructure:"width"`                      // 默认宽度
	Height      int     `json:"height,omitempty" yaml:"height,omitempty" mapstructure:"height"`                   // 默认高度
	MaxSkew     float64 `json:"max_skew,omitempty" yaml:"max_skew,omitempty" mapstructure:"max_skew"`             // 最大倾斜度
	DotCount    int     `json:"dot_count,omitempty" yaml:"dot_count,omitempty" mapstructure:"dot_count"`          // 干扰点数量
	LineCount   int     `json:"line_count,omitempty" yaml:"line_count,omitempty" mapstructure:"line_count"`       // 干扰线数量
	DefaultType string  `json:"default_type,omitempty" yaml:"default_type,omitempty" mapstructure:"default_type"` // 默认验证码类型
}

// String CaptchConfig的字符串表示
// @return CaptchConfig的字符串表示
func (c *CaptchaConfig) String() string {
	pretty, _ := json.MarshalPretty(c)
	return string(pretty)
}

// IpfsConfig IPFS配置
type IpfsConfig struct {
	Enabled        bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                         // 是否启用
	GatewayAddress string `json:"gateway_address,omitempty" yaml:"gateway_address,omitempty" mapstructure:"gateway_address"` // Gateway地址
	ApiAddress     string `json:"api_address,omitempty" yaml:"api_address,omitempty" mapstructure:"api_address"`             // API地址
}

// String IpfsConfig的字符串表示
// @return string IpfsConfig的字符串表示
func (i *IpfsConfig) String() string {
	pretty, _ := json.MarshalPretty(i)
	return string(pretty)
}

// EmailConfig 邮件配置
type EmailConfig struct {
	Enabled       bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                      // 是否启用
	SmtpHost      string `json:"smtp_host,omitempty" yaml:"smtp_host,omitempty" mapstructure:"smtp_host"`                // SMTP服务器地址
	SmtpPort      int    `json:"smtp_port,omitempty" yaml:"smtp_port,omitempty" mapstructure:"smtp_port"`                // SMTP服务器端口
	Username      string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username"`                   // SMTP用户名
	Password      string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"`                   // SMTP密码
	SenderAddress string `json:"sender_address,omitempty" yaml:"sender_address,omitempty" mapstructure:"sender_address"` // 发件人地址
	SenderName    string `json:"sender_name,omitempty" yaml:"sender_name,omitempty" mapstructure:"sender_name"`          // 发件人名称
	TlsEnabled    bool   `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled"`          // 是否启用TLS
	Debug         bool   `json:"debug,omitempty" yaml:"debug,omitempty" mapstructure:"debug"`                            // 是否开启调试模式
}

// String EmailConfig的字符串表示
// @return string EmailConfig的字符串表示
func (e *EmailConfig) String() string {
	pretty, _ := json.MarshalPretty(e)
	return string(pretty)
}

// AntsPoolConfig Ants线程池配置
type AntsPoolConfig struct {
	Enabled        bool `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                         // 是否启用
	PoolSize       int  `json:"pool_size,omitempty" yaml:"pool_size,omitempty" mapstructure:"pool_size"`                   // 线程池大小
	MaxPoolSize    int  `json:"max_pool_size,omitempty" yaml:"max_pool_size,omitempty" mapstructure:"max_pool_size"`       // 最大线程池大小
	ExpiryDuration int  `json:"expiry_duration,omitempty" yaml:"expiry_duration,omitempty" mapstructure:"expiry_duration"` // 工作协程过期时间（秒）
	PreAlloc       bool `json:"pre_alloc,omitempty" yaml:"pre_alloc,omitempty" mapstructure:"pre_alloc"`                   // 是否预分配工作协程
	Nonblocking    bool `json:"nonblocking,omitempty" yaml:"nonblocking,omitempty" mapstructure:"nonblocking"`             // 是否非阻塞模式
}

// String 返回AntsPoolConfig的字符串表示
func (a *AntsPoolConfig) String() string {
	pretty, _ := json.MarshalPretty(a)
	return string(pretty)
}

// GetPoolSize 获取线程池大小
func (a *AntsPoolConfig) GetPoolSize() int {
	return a.PoolSize
}

// GetMaxPoolSize 获取最大线程池大小
func (a *AntsPoolConfig) GetMaxPoolSize() int {
	return a.MaxPoolSize
}

// GetExpiryDuration 获取工作协程过期时间
func (a *AntsPoolConfig) GetExpiryDuration() time.Duration {
	return time.Duration(a.ExpiryDuration) * time.Second
}

// IsPreAlloc 是否预分配工作协程
func (a *AntsPoolConfig) IsPreAlloc() bool {
	return a.PreAlloc
}

// IsNonblocking 是否非阻塞模式
func (a *AntsPoolConfig) IsNonblocking() bool {
	return a.Nonblocking
}

// TunnyPoolConfig Tunny线程池配置
// type TunnyPoolConfig struct {
// 	Enabled          bool `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                          // 是否启用
// 	WorkerCount      int  `json:"worker_count,omitempty" yaml:"worker_count,omitempty" mapstructure:"worker_count"`          // 工作协程数量
// 	QueueSize        int  `json:"queue_size,omitempty" yaml:"queue_size,omitempty" mapstructure:"queue_size"`                // 任务队列大小，0表示无限制
// 	CleanupInterval  int  `json:"cleanup_interval,omitempty" yaml:"cleanup_interval,omitempty" mapstructure:"cleanup_interval"` // 清理间隔（秒）
// 	PanicHandlerEnabled bool `json:"panic_handler_enabled,omitempty" yaml:"panic_handler_enabled,omitempty" mapstructure:"panic_handler_enabled"` // 是否启用panic处理
// }

// String 返回TunnyPoolConfig的字符串表示
// func (t *TunnyPoolConfig) String() string {
// 	bytes, _ := json.MarshalIndent(t, "", " ")
// 	return string(bytes)
// }

// GetWorkerCount 获取工作协程数量
// func (t *TunnyPoolConfig) GetWorkerCount() int {
// 	return t.WorkerCount
// }

// GetQueueSize 获取任务队列大小
// func (t *TunnyPoolConfig) GetQueueSize() int {
// 	return t.QueueSize
// }

// GetCleanupInterval 获取清理间隔
// func (t *TunnyPoolConfig) GetCleanupInterval() time.Duration {
// 	return time.Duration(t.CleanupInterval) * time.Second
// }

// IsPanicHandlerEnabled 是否启用panic处理
// func (t *TunnyPoolConfig) IsPanicHandlerEnabled() bool {
// 	return t.PanicHandlerEnabled
// }

type FabricCAConfig struct {
	ImageName           string `json:"image_name,omitempty" yaml:"image_name,omitempty" mapstructure:"image_name"`                                     // 镜像名称
	CAName              string `json:"ca_name,omitempty" yaml:"ca_name,omitempty" mapstructure:"ca_name"`                                              // CA名称
	EnabledTls          bool   `json:"enabled_tls,omitempty" yaml:"enabled_tls,omitempty" mapstructure:"enabled_tls"`                                  // 是否启用TLS
	CAAdminUser         string `json:"ca_admin_user,omitempty" yaml:"ca_admin_user,omitempty" mapstructure:"ca_admin_user"`                            // CA管理员用户
	CAAdminPassword     string `json:"ca_admin_password,omitempty" yaml:"ca_admin_password,omitempty" mapstructure:"ca_admin_password"`                // CA管理员密码
	CAServerPort        int    `json:"ca_server_port,omitempty" yaml:"ca_server_port,omitempty" mapstructure:"ca_server_port"`                         // CA服务器端口
	JoinDockerNetwork   string `json:"join_docker_network,omitempty" yaml:"join_docker_network,omitempty" mapstructure:"join_docker_network"`          // 加入docker网络
	DockerNetworkIpAddr string `json:"docker_network_ip_addr,omitempty" yaml:"docker_network_ip_addr,omitempty" mapstructure:"docker_network_ip_addr"` // docker网络ip地址
	LocalAbsPath        string `json:"local_abs_path,omitempty" yaml:"local_abs_path,omitempty" mapstructure:"local_abs_path"`                         // 本地绝对路径
	LogInConsole        bool   `json:"log_in_console,omitempty" yaml:"log_in_console,omitempty" mapstructure:"log_in_console"`                         // 是否在控制台打印日志
	TestMode            bool   `json:"test_mode,omitempty" yaml:"test_mode,omitempty" mapstructure:"test_mode"`                                        // 是否是开发模式
}

func (f *FabricCAConfig) String() string {
	pretty, _ := json.MarshalPretty(f)
	return string(pretty)
}

// WebRTCConfig WebRTC配置
// @return string WebRTCConfig的字符串表示
type WebRTCConfig struct {
	Enabled        bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                            // 是否启用
	ListenAddr     string   `json:"listen_addr,omitempty" yaml:"listen_addr,omitempty" mapstructure:"listen_addr"`                // 监听地址
	ICEServers     []string `json:"ice_servers,omitempty" yaml:"ice_servers,omitempty" mapstructure:"ice_servers"`                // ICE服务器列表（可同时包含STUN和TURN服务器）
	MaxMessageSize int      `json:"max_message_size,omitempty" yaml:"max_message_size,omitempty" mapstructure:"max_message_size"` // 最大消息大小
	MinPort        int      `json:"min_port,omitempty" yaml:"min_port,omitempty" mapstructure:"min_port"`                         // 最小端口范围
	MaxPort        int      `json:"max_port,omitempty" yaml:"max_port,omitempty" mapstructure:"max_port"`                         // 最大端口范围
	TurnUsername   string   `json:"turn_username,omitempty" yaml:"turn_username,omitempty" mapstructure:"turn_username"`          // TURN服务器用户名
	TurnPassword   string   `json:"turn_password,omitempty" yaml:"turn_password,omitempty" mapstructure:"turn_password"`          // TURN服务器密码
	LogInConsole   bool     `json:"log_in_console,omitempty" yaml:"log_in_console,omitempty" mapstructure:"log_in_console"`       // 是否在控制台打印日志
}

// String WebRTCConfig的字符串表示
func (w *WebRTCConfig) String() string {
	pretty, _ := json.MarshalPretty(w)
	return string(pretty)
}

// AuthzConfig 权限控制配置结构体
type AuthzConfig struct {
	Enabled        bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                            // 是否启用权限控制
	ModelFile      string `json:"model_file,omitempty" yaml:"model_file,omitempty" mapstructure:"model_file"`                   // 模型文件路径
	PolicyFile     string `json:"policy_file,omitempty" yaml:"policy_file,omitempty" mapstructure:"policy_file"`                // 策略文件路径
	AutoCreateFile bool   `json:"auto_create_file,omitempty" yaml:"auto_create_file,omitempty" mapstructure:"auto_create_file"` // 是否自动创建文件
	DefaultAllow   bool   `json:"default_allow,omitempty" yaml:"default_allow,omitempty" mapstructure:"default_allow"`          // 默认是否允许
	LogEnabled     bool   `json:"log_enabled,omitempty" yaml:"log_enabled,omitempty" mapstructure:"log_enabled"`                // 是否启用日志
}

// String 返回AuthzConfig的字符串表示
func (a *AuthzConfig) String() string {
	pretty, _ := json.MarshalPretty(a)
	return string(pretty)
}

// MFAConfig MFA配置
// @description 多因素认证配置
// @struct MFAConfig
type MFAConfig struct {
	Enabled         bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                            // 是否启用
	DefaultProvider string `json:"default_provider,omitempty" yaml:"default_provider,omitempty" mapstructure:"default_provider"` // 默认认证提供商 (google/microsoft)
	Google          struct {
		Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty" mapstructure:"issuer"` // Google认证器的颁发者名称
	} `json:"google,omitempty" yaml:"google,omitempty" mapstructure:"google"` // Google认证器配置
	Microsoft struct {
		TenantID     string `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty" mapstructure:"tenant_id"`             // Microsoft租户ID
		ClientID     string `json:"client_id,omitempty" yaml:"client_id,omitempty" mapstructure:"client_id"`             // Microsoft客户端ID
		ClientSecret string `json:"client_secret,omitempty" yaml:"client_secret,omitempty" mapstructure:"client_secret"` // Microsoft客户端密钥
	} `json:"microsoft,omitempty" yaml:"microsoft,omitempty" mapstructure:"microsoft"` // Microsoft认证器配置
}

// String MFAConfig的字符串表示
// @return string MFAConfig的字符串表示
func (m *MFAConfig) String() string {
	pretty, _ := json.MarshalPretty(m)
	return string(pretty)
}

type OTELConfig struct {
	Enabled                bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                                                    // 是否启用
	ExporterServiceName    string `json:"exporter_service_name,omitempty" yaml:"exporter_service_name,omitempty" mapstructure:"exporter_service_name"`          // OTEL服务名称
	ExporterFilePath       string `json:"exporter_file_path,omitempty" yaml:"exporter_file_path,omitempty" mapstructure:"exporter_file_path"`                   // OTEL文件路径
	ExporterOTELInsecure   bool   `json:"exporter_otel_insecure,omitempty" yaml:"exporter_otel_insecure,omitempty" mapstructure:"exporter_otel_insecure"`       // OTEL是否不安全
	TracesExporter         string `json:"traces_exporter,omitempty" yaml:"traces_exporter,omitempty" mapstructure:"traces_exporter"`                            // 类型 otel,file
	ExporterZipkinEndpoint string `json:"exporter_zipkin_endpoint,omitempty" yaml:"exporter_zipkin_endpoint,omitempty" mapstructure:"exporter_zipkin_endpoint"` // Zipkin端点
	ExporterOTELEndpoint   string `json:"exporter_otel_endpoint,omitempty" yaml:"exporter_otel_endpoint,omitempty" mapstructure:"exporter_otel_endpoint"`       // OTEL端点
	ExporterOTELProtocol   string `json:"exporter_otel_protocol,omitempty" yaml:"exporter_otel_protocol,omitempty" mapstructure:"exporter_otel_protocol"`       // OTEL协议
}

// Config 配置
type Config struct {
	GrpcConfig       *GrpcConfig       `json:"grpc_config,omitempty" yaml:"grpc_config,omitempty" mapstructure:"grpc"`                   // grpc配置
	LoggerConfig     *LoggerConfig     `json:"logger_config,omitempty" yaml:"logger_config,omitempty" mapstructure:"logger"`             // logger配置
	HttpConfig       *HttpServerConfig `json:"http_config,omitempty" yaml:"http_config,omitempty" mapstructure:"http"`                   // http配置
	Libp2pConfig     *Libp2pConfig     `json:"libp2p_config,omitempty" yaml:"libp2p_config,omitempty" mapstructure:"libp2p"`             // libp2p配置
	DataSourceConfig *DataSourceConfig `json:"datasource_config,omitempty" yaml:"datasource_config,omitempty" mapstructure:"datasource"` // 数据源配置
	DockerConfig     *DockerConfig     `json:"docker_config,omitempty" yaml:"docker_config,omitempty" mapstructure:"docker"`             // docker配置
	IpfsConfig       *IpfsConfig       `json:"ipfs_config,omitempty" yaml:"ipfs_config,omitempty" mapstructure:"ipfs"`                   // IPFS配置
	CaptchaConfig    *CaptchaConfig    `json:"captcha_config,omitempty" yaml:"captcha_config,omitempty" mapstructure:"captcha"`          // 验证码配置
	EmailConfig      *EmailConfig      `json:"email_config,omitempty" yaml:"email_config,omitempty" mapstructure:"email"`                // 邮件配置
	AntsPoolConfig   *AntsPoolConfig   `json:"ants_pool_config,omitempty" yaml:"ants_pool_config,omitempty" mapstructure:"ants_pool"`    // Ants线程池配置
	FabricCAConfig   *FabricCAConfig   `json:"fabric_ca_config,omitempty" yaml:"fabric_ca_config,omitempty" mapstructure:"fabric_ca"`    // Fabric CA配置
	WebRTCConfig     *WebRTCConfig     `json:"webrtc_config,omitempty" yaml:"webrtc_config,omitempty" mapstructure:"webrtc"`             // WebRTC配置
	AuthzConfig      *AuthzConfig      `json:"authz_config,omitempty" yaml:"authz_config,omitempty" mapstructure:"authz"`                // 权限控制配置
	MFAConfig        *MFAConfig        `json:"mfa_config,omitempty" yaml:"mfa_config,omitempty" mapstructure:"mfa"`                      // MFA配置
	// TunnyPoolConfig  *TunnyPoolConfig  `json:"tunny_pool_config,omitempty" yaml:"tunny_pool_config,omitempty" mapstructure:"tunny_pool"` // Tunny线程池配置
}

// String 返回配置的字符串表示
// @return string 配置的字符串表示
func (c *Config) String() string {
	pretty, _ := json.MarshalPretty(c)
	return string(pretty)
}
