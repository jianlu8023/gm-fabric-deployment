package config

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
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
	TLSEnabled     bool   `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled"`                // 是否启用TLS
	TLSCertFile    string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`          // TLS证书文件
	TLSKeyFile     string `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`             // TLS私钥文件
	TLSCAFile      string `json:"tls_ca_file,omitempty" yaml:"tls_ca_file,omitempty" mapstructure:"tls_ca_file"`                // TLS CA证书文件
	TLSSkipVerify  bool   `json:"tls_skip_verify,omitempty" yaml:"tls_skip_verify,omitempty" mapstructure:"tls_skip_verify"`    // 是否跳过TLS验证
	TLSServerName  string `json:"tls_server_name,omitempty" yaml:"tls_server_name,omitempty" mapstructure:"tls_server_name"`    // TLS服务器名称
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

type IPWhiteListConfig struct {
	Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用IP白名单
	IPs     []string `json:"ips,omitempty" yaml:"ips,omitempty" mapstructure:"ips"`             // IP白名单，优先级高于黑名单
}

func (i *IPWhiteListConfig) String() string {
	pretty, _ := json.MarshalPretty(i)
	return string(pretty)
}

type IPBlackListConfig struct {
	Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用IP黑名单
	IPs     []string `json:"ips,omitempty" yaml:"ips,omitempty" mapstructure:"ips"`             // IP黑名单
}

func (i *IPBlackListConfig) String() string {
	pretty, _ := json.MarshalPretty(i)
	return string(pretty)
}

type RateLimitConfig struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用限流
	RPS     int64  `json:"rps,omitempty" yaml:"rps,omitempty" mapstructure:"rps"`             // 每秒请求数限制
	Burst   int    `json:"burst,omitempty" yaml:"burst,omitempty" mapstructure:"burst"`       // 令牌桶突发大小
	Type    string `json:"type,omitempty" yaml:"type,omitempty" mapstructure:"type"`          // 限流类型，可选值：time, ulule, ants, custom, juju
}

func (r *RateLimitConfig) String() string {
	pretty, _ := json.MarshalPretty(r)
	return string(pretty)
}

// HttpServerConfig http服务配置
type HttpServerConfig struct {
	Enabled         bool               `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                                  // 是否启用
	Address         string             `json:"address,omitempty" yaml:"address,omitempty" mapstructure:"address"`                                  // 服务地址
	ContextPath     string             `json:"context_path,omitempty" yaml:"context_path,omitempty" mapstructure:"context_path"`                   // 服务上下文路径
	RunMode         string             `json:"run_mode,omitempty" yaml:"run_mode,omitempty" mapstructure:"run_mode" `                              // 服务运行模式
	TlsEnabled      bool               `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled" `                     // 是否启用TLS
	TlsGM           bool               `json:"tls_gm,omitempty" yaml:"tls_gm,omitempty" mapstructure:"tls_gm"`                                     // 是否启用国密TLS
	TlsGMSingleCert bool               `json:"tls_gm_single_cert,omitempty" yaml:"tls_gm_single_cert,omitempty" mapstructure:"tls_gm_single_cert"` // 国密TLS是否使用单证书模式，默认false（使用双证书模式：一个用于签名，一个用于加密）
	TlsCertFile     string             `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file" `               // TLS证书文件
	TlsKeyFile      string             `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                   // TLS私钥文件
	TlsRCACertFile  string             `json:"tls_rca_cert_file,omitempty" yaml:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"`    // TLS根证书文件
	Http2Enabled    bool               `json:"http2_enabled,omitempty" yaml:"http2_enabled,omitempty" mapstructure:"http2_enabled"`                // 是否启用HTTP/2
	Pprof           bool               `json:"pprof,omitempty" yaml:"pprof,omitempty" mapstructure:"pprof"`                                        // 是否启用pprof
	UploadDir       string             `json:"upload_dir,omitempty" yaml:"upload_dir,omitempty" mapstructure:"upload_dir"`                         // 文件上传目录
	IPWhiteList     *IPWhiteListConfig `json:"ip_white_list,omitempty" yaml:"ip_white_list,omitempty" mapstructure:"ip_white_list"`                // 白名单配置
	IPBlackList     *IPBlackListConfig `json:"ip_black_list,omitempty" yaml:"ip_black_list,omitempty" mapstructure:"ip_black_list"`                // 黑名单配置
	RateLimit       RateLimitConfig    `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty" mapstructure:"rate_limit"`                         // 限流配置
	Language        string             `json:"language,omitempty" yaml:"language,omitempty" mapstructure:"language"`                               // 默认语言设置，支持zh,en等 // 国际化配置
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
	TlsGM          bool   `json:"tls_gm,omitempty" yaml:"tls_gm,omitempty" mapstructure:"tls_gm"`                                  // 是否启用国密TLS
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
	TlsGM              bool   `json:"tls_gm,omitempty" yaml:"tls_gm,omitempty" mapstructure:"tls_gm"`                                                  // 是否启用国密TLS
	TlsCertFile        string `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file"`                             // TLS证书文件
	TlsKeyFile         string `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                                // TLS私钥文件
	TlsRCACertFile     string `json:"tls_rca_cert_file,omitempty" yaml:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"`                 // TLS根证书文件
	TlsServerName      string `json:"tls_server_name,omitempty" yaml:"tls_server_name,omitempty" mapstructure:"tls_server_name"`                       // TLS服务端名称
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

type ICEServerConfig struct {
	URL      string `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url"`                // ICE服务器地址
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username"` // 用户名
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"` // 密码
}

func (i *ICEServerConfig) String() string {
	pretty, _ := json.MarshalPretty(i)
	return string(pretty)
}

// WebRTCConfig WebRTC配置
// @return string WebRTCConfig的字符串表示
type WebRTCConfig struct {
	Enabled      bool               `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                      // 是否启用
	ICEServers   []*ICEServerConfig `json:"ice_servers,omitempty" yaml:"ice_servers,omitempty" mapstructure:"ice_servers"`          // ICE服务器列表（可同时包含STUN和TURN服务器）
	MinPort      int                `json:"min_port,omitempty" yaml:"min_port,omitempty" mapstructure:"min_port"`                   // 最小端口范围
	MaxPort      int                `json:"max_port,omitempty" yaml:"max_port,omitempty" mapstructure:"max_port"`                   // 最大端口范围
	LogInConsole bool               `json:"log_in_console,omitempty" yaml:"log_in_console,omitempty" mapstructure:"log_in_console"` // 是否在控制台打印日志
	// MaxMessageSize int      `json:"max_message_size,omitempty" yaml:"max_message_size,omitempty" mapstructure:"max_message_size"` // 最大消息大小
	// ListenAddr     string   `json:"listen_addr,omitempty" yaml:"listen_addr,omitempty" mapstructure:"listen_addr"`                // 监听地址
}

// String WebRTCConfig的字符串表示
func (w *WebRTCConfig) String() string {
	pretty, _ := json.MarshalPretty(w)
	return string(pretty)
}

// GeoIPConfig GeoIP配置
type GeoIPConfig struct {
	Enabled      bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                   // 是否启用
	DatabasePath string `json:"database_path,omitempty" yaml:"database_path,omitempty" mapstructure:"database_path"` // 本地数据库文件路径
	DownloadURL  string `json:"download_url,omitempty" yaml:"download_url,omitempty" mapstructure:"download_url"`    // 数据库下载地址
	AutoDownload bool   `json:"auto_download,omitempty" yaml:"auto_download,omitempty" mapstructure:"auto_download"` // 是否自动下载数据库
}

func (g *GeoIPConfig) GoString() string {
	return g.String()
}

// String GeoIPConfig的字符串表示
func (g *GeoIPConfig) String() string {
	pretty, _ := json.MarshalPretty(g)
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

type MFAGoogleConfig struct {
	Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty" mapstructure:"issuer"` // Google认证器的颁发者名称
}

func (m *MFAGoogleConfig) String() string {
	pretty, _ := json.MarshalPretty(m)
	return string(pretty)
}

type MFAMicrosoftConfig struct {
	TenantID     string `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty" mapstructure:"tenant_id"`             // Microsoft租户ID
	ClientID     string `json:"client_id,omitempty" yaml:"client_id,omitempty" mapstructure:"client_id"`             // Microsoft客户端ID
	ClientSecret string `json:"client_secret,omitempty" yaml:"client_secret,omitempty" mapstructure:"client_secret"` // Microsoft客户端密钥
}

func (m *MFAMicrosoftConfig) String() string {
	pretty, _ := json.MarshalPretty(m)
	return string(pretty)
}

// MFAConfig MFA配置
// @description 多因素认证配置
// @struct MFAConfig
type MFAConfig struct {
	Enabled         bool                `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                            // 是否启用
	DefaultProvider string              `json:"default_provider,omitempty" yaml:"default_provider,omitempty" mapstructure:"default_provider"` // 默认认证提供商 (google/microsoft)
	Google          *MFAGoogleConfig    `json:"google,omitempty" yaml:"google,omitempty" mapstructure:"google"`                               // Google认证器配置
	Microsoft       *MFAMicrosoftConfig `json:"microsoft,omitempty" yaml:"microsoft,omitempty" mapstructure:"microsoft"`                      // Microsoft认证器配置
}

// String MFAConfig的字符串表示
// @return string MFAConfig的字符串表示
func (m *MFAConfig) String() string {
	pretty, _ := json.MarshalPretty(m)
	return string(pretty)
}

type Tracer struct {
	Exporters      string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
	OTELProtocol   string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
	OTELEndpoint   string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
	OTELInsecure   bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
	ZipkinEndpoint string `json:"zipkin_endpoint,omitempty" yaml:"zipkin_endpoint,omitempty" mapstructure:"zipkin_endpoint"`
	FilePath       string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
}

func (t *Tracer) String() string {
	pretty, _ := json.MarshalPretty(t)
	return string(pretty)
}

type Logger struct {
	Exporters    string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
	OTELProtocol string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
	OTELEndpoint string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
	OTELInsecure bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
	FilePath     string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
}

func (l *Logger) String() string {
	pretty, _ := json.MarshalPretty(l)
	return string(pretty)
}

type Metrics struct {
	Exporters    string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
	OTELProtocol string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
	OTELEndpoint string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
	OTELInsecure bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
	FilePath     string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
}

func (m *Metrics) String() string {
	pretty, _ := json.MarshalPretty(m)
	return string(pretty)
}

// TracerConfig OTEL追踪配置
// @description 统一管理OpenTelemetry的tracer、logger和meter配置
// @type struct
// @field Enabled bool 是否启用OTEL
// @field ServiceName string 服务名称
// @field LogInConsole bool 是否在控制台打印日志
// 以下是tracer相关配置
// TracerConfig OTEL追踪配置
// @description 统一管理OpenTelemetry的tracer、logger和meter配置
type TracerConfig struct {
	Enabled      bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                      // 是否启用
	ServiceName  string   `json:"service_name,omitempty" yaml:"service_name,omitempty" mapstructure:"service_name"`       // 服务名称
	LogInConsole bool     `json:"log_in_console,omitempty" yaml:"log_in_console,omitempty" mapstructure:"log_in_console"` // 是否在控制台打印日志
	Tracer       *Tracer  `json:"tracer,omitempty" yaml:"tracer,omitempty" mapstructure:"tracer"`                         // Tracer配置
	Logger       *Logger  `json:"logger,omitempty" yaml:"logger,omitempty" mapstructure:"logger"`                         // Logger配置
	Metrics      *Metrics `json:"metrics,omitempty" yaml:"metrics,omitempty" mapstructure:"metrics"`                      // Meter配置
}

func (c TracerConfig) String() string {
	pretty, _ := json.MarshalPretty(c)
	return string(pretty)
}

type Address struct {
	Host     string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"` // host
	Port     int    `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port"` // port
	UserName string `json:"user_name,omitempty" yaml:"user_name,omitempty" mapstructure:"user_name"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"`
}

func (a *Address) String() string {
	pretty, _ := json.MarshalPretty(a)
	return string(pretty)
}

type IpfsClusterConfig struct {
	Enabled         bool       `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                            // 是否启用
	Strategy        string     `json:"strategy,omitempty" yaml:"strategy,omitempty" mapstructure:"strategy"`                         // 策略
	TimeOutInterval int        `json:"timeout_interval,omitempty" yaml:"timeout_interval,omitempty" mapstructure:"timeout_interval"` // 超时时间
	LogLevel        string     `json:"log_level,omitempty" yaml:"log_level,omitempty" mapstructure:"log_level"`                      // 日志级别
	ReTries         int        `json:"retries,omitempty" yaml:"retries,omitempty" mapstructure:"retries"`                            // 重试次数
	Addresses       []*Address `json:"addresses,omitempty" yaml:"addresses,omitempty" mapstructure:"addresses"`
}

func (c IpfsClusterConfig) String() string {
	pretty, _ := json.MarshalPretty(c)
	return string(pretty)
}

// RedisConfig Redis配置结构体
type RedisConfig struct {
	Enabled  bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`       // 是否启用
	Host     string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"`                // Redis服务器地址
	Port     int    `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port"`                // Redis服务器端口
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"`    // Redis密码
	DB       int    `json:"db,omitempty" yaml:"db,omitempty" mapstructure:"db"`                      // Redis数据库编号
	PoolSize int    `json:"pool_size,omitempty" yaml:"pool_size,omitempty" mapstructure:"pool_size"` // 连接池大小
}

// String 返回RedisConfig的字符串表示
// @return string RedisConfig的字符串表示
func (r *RedisConfig) String() string {
	pretty, _ := json.MarshalPretty(r)
	return string(pretty)
}

// KvDatabaseConfig KV数据库配置结构体
type KvDatabaseConfig struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用
	DbType  string `json:"db_type,omitempty" yaml:"db_type,omitempty" mapstructure:"db_type"` // 数据库类型 (leveldb, pebble, badger)
	DbPath  string `json:"db_path,omitempty" yaml:"db_path,omitempty" mapstructure:"db_path"` // 数据库文件路径
}

// String 返回KvDatabaseConfig的字符串表示
// @return string KvDatabaseConfig的字符串表示
func (k *KvDatabaseConfig) String() string {
	pretty, _ := json.MarshalPretty(k)
	return string(pretty)
}

// AIConfig AI配置结构体
type AIConfig struct {
	Enabled            bool              `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                                        // 是否启用
	APIKey             string            `json:"api_key,omitempty" yaml:"api_key,omitempty" mapstructure:"api_key"`                                        // API密钥
	APIEndpoint        string            `json:"api_endpoint,omitempty" yaml:"api_endpoint,omitempty" mapstructure:"api_endpoint"`                         // API端点
	DefaultModel       string            `json:"default_model,omitempty" yaml:"default_model,omitempty" mapstructure:"default_model"`                      // 默认模型
	Timeout            int               `json:"timeout,omitempty" yaml:"timeout,omitempty" mapstructure:"timeout"`                                        // 超时时间(秒)
	InsecureSkipVerify bool              `json:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty" mapstructure:"insecure_skip_verify"` // 是否跳过证书验证
	Headers            map[string]string `json:"headers,omitempty" yaml:"headers,omitempty" mapstructure:"headers"`                                        // 自定义请求头
	UseToonFormat      bool              `json:"use_toon_format,omitempty" yaml:"use_toon_format,omitempty" mapstructure:"use_toon_format"`                // 是否使用TOON格式减少token使用
}

// String 返回AIConfig的字符串表示
// @return string AIConfig的字符串表示
func (a *AIConfig) String() string {
	pretty, _ := json.MarshalPretty(a)
	return string(pretty)
}

// CertificateConfig 证书配置结构体
type CertificateConfig struct {
	Enabled     bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                // 是否启用证书功能
	DefaultAlgo string `json:"default_algo,omitempty" yaml:"default_algo,omitempty" mapstructure:"default_algo"` // 默认算法 (RSA/ECC/SM2)
	RootSubject string `json:"root_subject,omitempty" yaml:"root_subject,omitempty" mapstructure:"root_subject"` // 根证书主题信息，使用/分割(/OU=/O=/CN等)
	CertPath    string `json:"cert_path,omitempty" yaml:"cert_path,omitempty" mapstructure:"cert_path"`          // 证书存储路径
}

// String 返回CertificateConfig的字符串表示
// @return string CertificateConfig的字符串表示
func (c *CertificateConfig) String() string {
	pretty, _ := json.MarshalPretty(c)
	return string(pretty)
}

// Config 配置
type Config struct {
	GrpcConfig        *GrpcConfig        `json:"grpc_config,omitempty" yaml:"grpc_config,omitempty" mapstructure:"grpc"`                         // grpc配置
	LoggerConfig      *LoggerConfig      `json:"logger_config,omitempty" yaml:"logger_config,omitempty" mapstructure:"logger"`                   // logger配置
	HttpConfig        *HttpServerConfig  `json:"http_config,omitempty" yaml:"http_config,omitempty" mapstructure:"http"`                         // http配置
	Libp2pConfig      *Libp2pConfig      `json:"libp2p_config,omitempty" yaml:"libp2p_config,omitempty" mapstructure:"libp2p"`                   // libp2p配置
	DataSourceConfig  *DataSourceConfig  `json:"datasource_config,omitempty" yaml:"datasource_config,omitempty" mapstructure:"datasource"`       // 数据源配置
	DockerConfig      *DockerConfig      `json:"docker_config,omitempty" yaml:"docker_config,omitempty" mapstructure:"docker"`                   // docker配置
	IpfsConfig        *IpfsConfig        `json:"ipfs_config,omitempty" yaml:"ipfs_config,omitempty" mapstructure:"ipfs"`                         // IPFS配置
	CaptchaConfig     *CaptchaConfig     `json:"captcha_config,omitempty" yaml:"captcha_config,omitempty" mapstructure:"captcha"`                // 验证码配置
	EmailConfig       *EmailConfig       `json:"email_config,omitempty" yaml:"email_config,omitempty" mapstructure:"email"`                      // 邮件配置
	AntsPoolConfig    *AntsPoolConfig    `json:"ants_pool_config,omitempty" yaml:"ants_pool_config,omitempty" mapstructure:"ants_pool"`          // Ants线程池配置
	FabricCAConfig    *FabricCAConfig    `json:"fabric_ca_config,omitempty" yaml:"fabric_ca_config,omitempty" mapstructure:"fabric_ca"`          // Fabric CA配置
	WebRTCConfig      *WebRTCConfig      `json:"webrtc_config,omitempty" yaml:"webrtc_config,omitempty" mapstructure:"webrtc"`                   // WebRTC配置
	AuthzConfig       *AuthzConfig       `json:"authz_config,omitempty" yaml:"authz_config,omitempty" mapstructure:"authz"`                      // 权限控制配置
	MFAConfig         *MFAConfig         `json:"mfa_config,omitempty" yaml:"mfa_config,omitempty" mapstructure:"mfa"`                            // MFA配置
	TracerConfig      *TracerConfig      `json:"tracer_config,omitempty" yaml:"tracer_config,omitempty" mapstructure:"tracer"`                   // 链路追踪配置
	IpfsClusterConfig *IpfsClusterConfig `json:"ipfs_cluster_config,omitempty" yaml:"ipfs_cluster_config,omitempty" mapstructure:"ipfs_cluster"` // ipfs cluster配置
	RedisConfig       *RedisConfig       `json:"redis_config,omitempty" yaml:"redis_config,omitempty" mapstructure:"redis"`                      // Redis配置
	KvDatabaseConfig  *KvDatabaseConfig  `json:"kvdatabase_config,omitempty" yaml:"kvdatabase_config,omitempty" mapstructure:"kvdatabase"`       // KV数据库配置
	AIConfig          *AIConfig          `json:"ai_config,omitempty" yaml:"ai_config,omitempty" mapstructure:"ai"`                               // AI配置
	CertificateConfig *CertificateConfig `json:"certificate_config,omitempty" yaml:"certificate_config,omitempty" mapstructure:"certificate"`    // 证书配置
	GeoIPConfig       *GeoIPConfig       `json:"geoip_config,omitempty" yaml:"geoip_config,omitempty" mapstructure:"geoip"`                      // GeoIP配置

	// TunnyPoolConfig  *TunnyPoolConfig  `json:"tunny_pool_config,omitempty" yaml:"tunny_pool_config,omitempty" mapstructure:"tunny_pool"` // Tunny线程池配置
}

// String 返回配置的字符串表示
// @return string 配置的字符串表示
func (c *Config) String() string {
	pretty, _ := json.MarshalPretty(c)
	return string(pretty)
}
