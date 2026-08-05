package config

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"reflect"
	"strings"
)

// marshalConfig 序列化配置结构体，显示所有字段包括false的bool值
//
// @description 通过反射遍历结构体所有字段，将其转换为map[string]interface{}后进行JSON序列化，
//
//	确保所有字段（包括false的bool、空字符串等零值字段）都能被正确显示
//
// @param v interface{} 需要序列化的结构体指针
// @return string JSON格式化字符串
func marshalConfig(v interface{}) string {
	data, err := json.MarshalPretty(structToMap(v))
	if err != nil {
		return ""
	}
	return string(data)
}

// structToMap 将结构体转换为map[string]interface{}
//
// @description 使用反射遍历结构体所有字段，忽略omitempty标签，确保所有字段都能被包含在结果中
// @param v interface{} 需要转换的结构体
// @return map[string]interface{} 转换后的map
func structToMap(v interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	if v == nil {
		return result
	}
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return result
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return result
	}
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonName := strings.Split(jsonTag, ",")[0]
		if jsonName == "-" || jsonName == "" {
			continue
		}
		// 跳过nil指针和nil切片/映射字段，避免打印null
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}
		if (field.Kind() == reflect.Slice || field.Kind() == reflect.Map) && field.IsNil() {
			continue
		}
		result[jsonName] = fieldValueToInterface(field)
	}
	return result
}

// fieldValueToInterface 将字段值转换为interface{}
//
// @description 处理不同类型的字段值，包括基本类型、嵌套结构体、指针、切片等
// @param field reflect.Value 字段值
// @return interface{} 转换后的值
func fieldValueToInterface(field reflect.Value) interface{} {
	switch field.Kind() {
	case reflect.Ptr:
		if field.IsNil() {
			return nil
		}
		return structToMap(field.Interface())
	case reflect.Struct:
		if field.CanAddr() {
			return structToMap(field.Addr().Interface())
		}
		return structToMap(field.Interface())
	case reflect.Slice, reflect.Array:
		if field.IsNil() {
			return nil
		}
		result := make([]interface{}, field.Len())
		for i := 0; i < field.Len(); i++ {
			elem := field.Index(i)
			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					result[i] = nil
				} else {
					result[i] = fieldValueToInterface(elem)
				}
			} else if elem.Kind() == reflect.Struct {
				if elem.CanAddr() {
					result[i] = structToMap(elem.Addr().Interface())
				} else {
					result[i] = structToMap(elem.Interface())
				}
			} else {
				result[i] = elem.Interface()
			}
		}
		return result
	case reflect.Map:
		if field.IsNil() {
			return nil
		}
		result := make(map[string]interface{})
		for _, key := range field.MapKeys() {
			result[toString(key)] = fieldValueToInterface(field.MapIndex(key))
		}
		return result
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return field.Interface()
	case reflect.Interface:
		if field.IsNil() {
			return nil
		}
		return fieldValueToInterface(field.Elem())
	default:
		return field.Interface()
	}
}

// toString 将reflect.Value转换为字符串
//
// @description 将map的key转换为字符串类型
// @param v reflect.Value 需要转换的值
// @return string 转换后的字符串
func toString(v reflect.Value) string {
	if v.Kind() == reflect.String {
		return v.String()
	}
	return ""
}

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
func (c *DockerConfig) String() string {
	return marshalConfig(c)
}

func (c *DockerConfig) GoString() string {
	return c.String()
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
	Charset        string `json:"charset,omitempty" yaml:"charset,omitempty" mapstructure:"charset"`                            // 数据库字符集，默认utf8mb4
	Timezone       string `json:"timezone,omitempty" yaml:"timezone,omitempty" mapstructure:"timezone"`                         // 数据库时区，默认Local(MySQL)或Asia/Shanghai(PostgreSQL)
	SSLMode        string `json:"ssl_mode,omitempty" yaml:"ssl_mode,omitempty" mapstructure:"ssl_mode"`                         // PostgreSQL SSL模式，默认disable
	ParseTime      bool   `json:"parse_time,omitempty" yaml:"parse_time,omitempty" mapstructure:"parse_time"`                   // 是否解析时间类型，默认true
}

// String 返回DataSourceConfig的字符串表示
// @return string 字符串表示
func (c *DataSourceConfig) String() string {
	return marshalConfig(c)
}

func (c *DataSourceConfig) GoString() string {
	return c.String()
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
func (c *LoggerConfig) String() string {
	return marshalConfig(c)
}

func (c *LoggerConfig) GoString() string {
	return c.String()
}

type IPWhiteListConfig struct {
	Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用IP白名单
	IPs     []string `json:"ips,omitempty" yaml:"ips,omitempty" mapstructure:"ips"`             // IP白名单，优先级高于黑名单
}

func (c *IPWhiteListConfig) String() string {
	return marshalConfig(c)
}
func (c *IPWhiteListConfig) GoString() string {
	return c.String()
}

type IPBlackListConfig struct {
	Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用IP黑名单
	IPs     []string `json:"ips,omitempty" yaml:"ips,omitempty" mapstructure:"ips"`             // IP黑名单
}

func (c *IPBlackListConfig) String() string {
	return marshalConfig(c)
}
func (c *IPBlackListConfig) GoString() string {
	return c.String()
}

// TrustedProxiesConfig 可信代理配置
//
// @description 配置可信代理网段，仅当请求的 RemoteAddr 命中该网段时，才信任 X-Forwarded-For / X-Real-IP 头
// @struct
type TrustedProxiesConfig struct {
	Enabled bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用可信代理
	IPs     []string `json:"ips,omitempty" yaml:"ips,omitempty" mapstructure:"ips"`             // 可信代理网段列表，支持精确 IP 和 CIDR 格式
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *TrustedProxiesConfig) String() string {
	return marshalConfig(c)
}

func (c *TrustedProxiesConfig) GoString() string {
	return c.String()
}

type RateLimitConfig struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用限流
	RPS     int64  `json:"rps,omitempty" yaml:"rps,omitempty" mapstructure:"rps"`             // 每秒请求数限制
	Burst   int    `json:"burst,omitempty" yaml:"burst,omitempty" mapstructure:"burst"`       // 令牌桶突发大小
	Type    string `json:"type,omitempty" yaml:"type,omitempty" mapstructure:"type"`          // 限流类型，可选值：time, ulule, ants, custom, juju
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *RateLimitConfig) String() string {
	return marshalConfig(c)
}
func (c *RateLimitConfig) GoString() string {
	return c.String()
}

// JWTConfig JWT认证配置
//
// @description 配置JWT令牌的签名密钥与会话有效期，密钥应通过配置文件注入而非硬编码在源码中
// @struct
type JWTConfig struct {
	Enabled    bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`             // 是否启用JWT认证
	Secret     string `json:"secret,omitempty" yaml:"secret,omitempty" mapstructure:"secret"`                // JWT签名密钥，禁止使用默认值上生产环境
	SessionTTL int    `json:"session_ttl,omitempty" yaml:"session_ttl,omitempty" mapstructure:"session_ttl"` // 会话有效期（秒），默认86400（24小时）
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *JWTConfig) String() string {
	return marshalConfig(c)
}
func (c *JWTConfig) GoString() string {
	return c.String()
}

// CORSConfig CORS跨域配置
//
// @description 配置CORS中间件的行为，遵循W3C CORS规范：禁止 AllowOrigins 含 "*" 与 AllowCredentials=true 同时使用
// @struct
type CORSConfig struct {
	Enabled          bool     `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                               // 是否启用CORS中间件
	AllowOrigins     []string `json:"allow_origins,omitempty" yaml:"allow_origins,omitempty" mapstructure:"allow_origins"`             // 允许的Origin白名单，含 "*" 时表示允许所有Origin（此时AllowCredentials必须为false）
	AllowMethods     []string `json:"allow_methods,omitempty" yaml:"allow_methods,omitempty" mapstructure:"allow_methods"`             // 允许的HTTP方法
	AllowHeaders     []string `json:"allow_headers,omitempty" yaml:"allow_headers,omitempty" mapstructure:"allow_headers"`             // 允许的请求头
	ExposeHeaders    []string `json:"expose_headers,omitempty" yaml:"expose_headers,omitempty" mapstructure:"expose_headers"`          // 允许暴露给浏览器的响应头
	AllowCredentials bool     `json:"allow_credentials,omitempty" yaml:"allow_credentials,omitempty" mapstructure:"allow_credentials"` // 是否允许携带Cookie等凭证，为true时AllowOrigins不可含"*"
	AllowWildcard    bool     `json:"allow_wildcard,omitempty" yaml:"allow_wildcard,omitempty" mapstructure:"allow_wildcard"`          // 是否启用子域名通配符匹配，如 https://*.example.com
	MaxAge           int      `json:"max_age,omitempty" yaml:"max_age,omitempty" mapstructure:"max_age"`                               // 预检请求缓存时间（秒），默认0表示不设置
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *CORSConfig) String() string {
	return marshalConfig(c)
}

func (c *CORSConfig) GoString() string {
	return c.String()
}

// RecoveryConfig panic恢复中间件配置
//
// @description 配置recovery中间件的行为，控制panic恢复后的日志输出、堆栈记录等
// @struct
type RecoveryConfig struct {
	Enabled       bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                      // 是否启用recovery中间件，默认true（建议始终启用）
	EnableStack   bool   `json:"enable_stack,omitempty" yaml:"enable_stack,omitempty" mapstructure:"enable_stack"`       // 是否在日志中输出goroutine堆栈，生产环境建议false避免日志膨胀
	LogLevel      string `json:"log_level,omitempty" yaml:"log_level,omitempty" mapstructure:"log_level"`                // panic日志级别：error | warn | fatal，默认error
	CustomMessage string `json:"custom_message,omitempty" yaml:"custom_message,omitempty" mapstructure:"custom_message"` // 自定义响应消息，为空时使用默认"Server Internal Error"
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *RecoveryConfig) String() string {
	return marshalConfig(c)
}

func (c *RecoveryConfig) GoString() string {
	return c.String()
}

// ResponseLogConfig 响应日志中间件配置
//
// @description 配置response日志中间件的行为，控制响应体捕获大小、日志级别等；默认不启用，需要时显式开启
// @struct
type ResponseLogConfig struct {
	Enabled      bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                      // 是否启用响应日志中间件，默认false（按需开启）
	MaxBodyBytes int    `json:"max_body_bytes,omitempty" yaml:"max_body_bytes,omitempty" mapstructure:"max_body_bytes"` // 响应体最大捕获字节数，超过部分仅记录长度不缓存，<=0 时使用默认 1024
	LogLevel     string `json:"log_level,omitempty" yaml:"log_level,omitempty" mapstructure:"log_level"`                // 日志级别：debug | info，默认debug
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *ResponseLogConfig) String() string {
	return marshalConfig(c)
}
func (c *ResponseLogConfig) GoString() string {
	return c.String()
}

// SessionConfig 会话存储配置
//
// @description 配置服务端会话的存储方式与生命周期，由 auth 包消费；StoreType=memory 时单机有效，redis 时跨实例共享
// @struct
type SessionConfig struct {
	Enabled           bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                                  // 是否启用服务端会话
	TTL               int    `json:"ttl,omitempty" yaml:"ttl,omitempty" mapstructure:"ttl"`                                              // 会话有效期（秒），<=0 时使用默认 86400
	CleanupInterval   int    `json:"cleanup_interval,omitempty" yaml:"cleanup_interval,omitempty" mapstructure:"cleanup_interval"`       // 过期清理间隔（秒），仅 memory 模式，<=0 时使用默认 3600
	StoreType         string `json:"store_type,omitempty" yaml:"store_type,omitempty" mapstructure:"store_type"`                         // 存储类型：memory | redis，默认 memory
	SlidingExpiration bool   `json:"sliding_expiration,omitempty" yaml:"sliding_expiration,omitempty" mapstructure:"sliding_expiration"` // 滑动过期（每次访问续期）vs 绝对过期
	SessionIDSource   string `json:"session_id_source,omitempty" yaml:"session_id_source,omitempty" mapstructure:"session_id_source"`    // sessionID 来源：auto | header | cookie | query，仅 session_only 模式用
	CookieName        string `json:"cookie_name,omitempty" yaml:"cookie_name,omitempty" mapstructure:"cookie_name"`                      // cookie 模式使用的名称，可选
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *SessionConfig) String() string {
	return marshalConfig(c)
}
func (c *SessionConfig) GoString() string {
	return c.String()
}

// AuthConfig 认证配置
//
// @description 认证总配置，编排 JWT 与 Session 两个子配置；Enabled=false 时不构造 Authenticator，所有路由按公开处理
// @struct
type AuthConfig struct {
	Enabled           bool           `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                                  // 总开关：是否启用认证
	TokenSource       string         `json:"token_source,omitempty" yaml:"token_source,omitempty" mapstructure:"token_source"`                   // token 来源：header | cookie | query | auto，默认 auto
	TokenHeader       string         `json:"token_header,omitempty" yaml:"token_header,omitempty" mapstructure:"token_header"`                   // token 头名称，默认 Authorization
	TokenPrefix       string         `json:"token_prefix,omitempty" yaml:"token_prefix,omitempty" mapstructure:"token_prefix"`                   // token 前缀，默认 Bearer
	SlidingExpiration bool           `json:"sliding_expiration,omitempty" yaml:"sliding_expiration,omitempty" mapstructure:"sliding_expiration"` // 滑动过期（每次访问续期）
	FreshWindow       int            `json:"fresh_window,omitempty" yaml:"fresh_window,omitempty" mapstructure:"fresh_window"`                   // 敏感操作二次校验窗口（秒），<=0 表示不启用
	OptionalPaths     []string       `json:"optional_paths,omitempty" yaml:"optional_paths,omitempty" mapstructure:"optional_paths"`             // 可选认证放行路径白名单
	JWT               *JWTConfig     `json:"jwt,omitempty" yaml:"jwt,omitempty" mapstructure:"jwt"`                                              // JWT 子配置
	Session           *SessionConfig `json:"session,omitempty" yaml:"session,omitempty" mapstructure:"session"`                                  // Session 子配置
}

// String 返回配置的字符串表示
//
// @return string 配置的字符串表示
func (c *AuthConfig) String() string {
	return marshalConfig(c)
}
func (c *AuthConfig) GoString() string {
	return c.String()
}

// HttpServerConfig http服务配置
type HttpServerConfig struct {
	Enabled         bool                  `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                                  // 是否启用
	Address         string                `json:"address,omitempty" yaml:"address,omitempty" mapstructure:"address"`                                  // 服务地址
	ContextPath     string                `json:"context_path,omitempty" yaml:"context_path,omitempty" mapstructure:"context_path"`                   // 服务上下文路径
	RunMode         string                `json:"run_mode,omitempty" yaml:"run_mode,omitempty" mapstructure:"run_mode" `                              // 服务运行模式
	TlsEnabled      bool                  `json:"tls_enabled,omitempty" yaml:"tls_enabled,omitempty" mapstructure:"tls_enabled" `                     // 是否启用TLS
	TlsGM           bool                  `json:"tls_gm,omitempty" yaml:"tls_gm,omitempty" mapstructure:"tls_gm"`                                     // 是否启用国密TLS
	TlsGMSingleCert bool                  `json:"tls_gm_single_cert,omitempty" yaml:"tls_gm_single_cert,omitempty" mapstructure:"tls_gm_single_cert"` // 国密TLS是否使用单证书模式，默认false（使用双证书模式：一个用于签名，一个用于加密）
	TlsCertFile     string                `json:"tls_cert_file,omitempty" yaml:"tls_cert_file,omitempty" mapstructure:"tls_cert_file" `               // TLS证书文件
	TlsKeyFile      string                `json:"tls_key_file,omitempty" yaml:"tls_key_file,omitempty" mapstructure:"tls_key_file"`                   // TLS私钥文件
	TlsRCACertFile  string                `json:"tls_rca_cert_file,omitempty" yaml:"tls_rca_cert_file,omitempty" mapstructure:"tls_rca_cert_file"`    // TLS根证书文件
	Http2Enabled    bool                  `json:"http2_enabled,omitempty" yaml:"http2_enabled,omitempty" mapstructure:"http2_enabled"`                // 是否启用HTTP/2
	Pprof           bool                  `json:"pprof,omitempty" yaml:"pprof,omitempty" mapstructure:"pprof"`                                        // 是否启用pprof
	UploadDir       string                `json:"upload_dir,omitempty" yaml:"upload_dir,omitempty" mapstructure:"upload_dir"`                         // 文件上传目录
	Recovery        *RecoveryConfig       `json:"recovery,omitempty" yaml:"recovery,omitempty" mapstructure:"recovery"`                               // panic恢复配置，建议始终启用
	ResponseLog     *ResponseLogConfig    `json:"response_log,omitempty" yaml:"response_log,omitempty" mapstructure:"response_log"`                   // 响应日志配置，默认不启用，需要时显式开启
	IPWhiteList     *IPWhiteListConfig    `json:"ip_white_list,omitempty" yaml:"ip_white_list,omitempty" mapstructure:"ip_white_list"`                // 白名单配置
	IPBlackList     *IPBlackListConfig    `json:"ip_black_list,omitempty" yaml:"ip_black_list,omitempty" mapstructure:"ip_black_list"`                // 黑名单配置
	TrustedProxies  *TrustedProxiesConfig `json:"trusted_proxies,omitempty" yaml:"trusted_proxies,omitempty" mapstructure:"trusted_proxies"`          // 可信代理配置，影响 X-Forwarded-For / X-Real-IP 的解析
	RateLimit       *RateLimitConfig      `json:"rate_limit,omitempty" yaml:"rate_limit,omitempty" mapstructure:"rate_limit"`                         // 限流配置
	Auth            *AuthConfig           `json:"auth,omitempty" yaml:"auth,omitempty" mapstructure:"auth"`                                           // 认证配置，编排 JWT 与 Session
	CORS            *CORSConfig           `json:"cors,omitempty" yaml:"cors,omitempty" mapstructure:"cors"`                                           // CORS跨域配置，遵循W3C规范：禁止 "*" 与 AllowCredentials 同时使用
	Language        string                `json:"language,omitempty" yaml:"language,omitempty" mapstructure:"language"`                               // 默认语言设置，支持zh,en等 // 国际化配置
}

// String 返回配置的字符串表示
// @return string 配置的字符串表示
func (c *HttpServerConfig) String() string {
	return marshalConfig(c)
}
func (c *HttpServerConfig) GoString() string {
	return c.String()
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
func (c *GrpcServerConfig) String() string {
	return marshalConfig(c)
}
func (c *GrpcServerConfig) GoString() string {
	return c.String()
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
func (c *GrpcClientConfig) String() string {
	return marshalConfig(c)
}
func (c *GrpcClientConfig) GoString() string {
	return c.String()
}

// GrpcConfig 配置Grpc
type GrpcConfig struct {
	Enabled bool              `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用
	Server  *GrpcServerConfig `json:"server,omitempty" yaml:"server,omitempty" mapstructure:"server"`    // 服务端配置
	Client  *GrpcClientConfig `json:"client,omitempty" yaml:"client,omitempty" mapstructure:"client"`    // 客户端配置
}

// String GrpcConfig的字符串表示
// @return string GrpcConfig的字符串表示
func (c *GrpcConfig) String() string {
	return marshalConfig(c)
}
func (c *GrpcConfig) GoString() string {
	return c.String()
}

// Identity 配置身份信息
type Identity struct {
	PeerID  string `json:"peer_id,omitempty" yaml:"peer_id,omitempty" mapstructure:"peer_id"`
	PrivKey string `json:"-" yaml:",omitempty" mapstructure:"privkey"`
}

// String 返回Identity的字符串表示
// @return string Identity的字符串表示
func (c *Identity) String() string {
	return marshalConfig(c)
}
func (c *Identity) GoString() string {
	return c.String()
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
func (c *Libp2pConfig) String() string {
	return marshalConfig(c)
}
func (c *Libp2pConfig) GoString() string {
	return c.String()
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
	return marshalConfig(c)
}
func (c *CaptchaConfig) GoString() string {
	return c.String()
}

// IpfsConfig IPFS配置
type IpfsConfig struct {
	Enabled        bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                         // 是否启用
	GatewayAddress string `json:"gateway_address,omitempty" yaml:"gateway_address,omitempty" mapstructure:"gateway_address"` // Gateway地址
	ApiAddress     string `json:"api_address,omitempty" yaml:"api_address,omitempty" mapstructure:"api_address"`             // API地址
}

// String IpfsConfig的字符串表示
// @return string IpfsConfig的字符串表示
func (c *IpfsConfig) String() string {
	return marshalConfig(c)
}
func (c *IpfsConfig) GoString() string {
	return c.String()
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
func (c *EmailConfig) String() string {
	return marshalConfig(c)
}
func (c *EmailConfig) GoString() string {
	return c.String()
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
func (c *AntsPoolConfig) String() string {
	return marshalConfig(c)
}
func (c *AntsPoolConfig) GoString() string {
	return c.String()
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

func (c *FabricCAConfig) String() string {
	return marshalConfig(c)
}
func (c *FabricCAConfig) GoString() string {
	return c.String()
}

type ICEServerConfig struct {
	URL      string `json:"url,omitempty" yaml:"url,omitempty" mapstructure:"url"`                // ICE服务器地址
	Username string `json:"username,omitempty" yaml:"username,omitempty" mapstructure:"username"` // 用户名
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"` // 密码
}

func (c *ICEServerConfig) String() string {
	return marshalConfig(c)
}
func (c *ICEServerConfig) GoString() string {
	return c.String()
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
func (c *WebRTCConfig) String() string {
	return marshalConfig(c)
}
func (c *WebRTCConfig) GoString() string {
	return c.String()
}

// GeoIPConfig GeoIP配置
type GeoIPConfig struct {
	Enabled      bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                   // 是否启用
	DatabasePath string `json:"database_path,omitempty" yaml:"database_path,omitempty" mapstructure:"database_path"` // 本地数据库文件路径
	DownloadURL  string `json:"download_url,omitempty" yaml:"download_url,omitempty" mapstructure:"download_url"`    // 数据库下载地址
	AutoDownload bool   `json:"auto_download,omitempty" yaml:"auto_download,omitempty" mapstructure:"auto_download"` // 是否自动下载数据库
}

func (c *GeoIPConfig) GoString() string {
	return c.String()
}

// String GeoIPConfig的字符串表示
func (c *GeoIPConfig) String() string {
	return marshalConfig(c)
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
func (c *AuthzConfig) String() string {
	return marshalConfig(c)
}
func (c *AuthzConfig) GoString() string {
	return c.String()
}

type MFAGoogleConfig struct {
	Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty" mapstructure:"issuer"` // Google认证器的颁发者名称
}

func (c *MFAGoogleConfig) String() string {
	return marshalConfig(c)
}
func (c *MFAGoogleConfig) GoString() string {
	return c.String()
}

type MFAMicrosoftConfig struct {
	TenantID     string `json:"tenant_id,omitempty" yaml:"tenant_id,omitempty" mapstructure:"tenant_id"`             // Microsoft租户ID
	ClientID     string `json:"client_id,omitempty" yaml:"client_id,omitempty" mapstructure:"client_id"`             // Microsoft客户端ID
	ClientSecret string `json:"client_secret,omitempty" yaml:"client_secret,omitempty" mapstructure:"client_secret"` // Microsoft客户端密钥
}

func (c *MFAMicrosoftConfig) String() string {
	return marshalConfig(c)
}
func (c *MFAMicrosoftConfig) GoString() string {
	return c.String()
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
func (c *MFAConfig) String() string {
	return marshalConfig(c)
}
func (c *MFAConfig) GoString() string {
	return c.String()
}

type Tracer struct {
	Exporters      string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
	OTELProtocol   string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
	OTELEndpoint   string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
	OTELInsecure   bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
	ZipkinEndpoint string `json:"zipkin_endpoint,omitempty" yaml:"zipkin_endpoint,omitempty" mapstructure:"zipkin_endpoint"`
	FilePath       string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
}

func (c *Tracer) String() string {
	return marshalConfig(c)
}
func (c *Tracer) GoString() string {
	return c.String()
}

type Logger struct {
	Exporters    string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
	OTELProtocol string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
	OTELEndpoint string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
	OTELInsecure bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
	FilePath     string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
}

func (c *Logger) String() string {
	return marshalConfig(c)
}
func (c *Logger) GoString() string {
	return c.String()
}

type Metrics struct {
	Exporters    string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
	OTELProtocol string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
	OTELEndpoint string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
	OTELInsecure bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
	FilePath     string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
}

func (c *Metrics) String() string {
	return marshalConfig(c)
}
func (c *Metrics) GoString() string {
	return c.String()
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

func (c *TracerConfig) String() string {
	return marshalConfig(&c)
}
func (c *TracerConfig) GoString() string {
	return c.String()
}

type Address struct {
	Host     string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host"` // host
	Port     int    `json:"port,omitempty" yaml:"port,omitempty" mapstructure:"port"` // port
	UserName string `json:"user_name,omitempty" yaml:"user_name,omitempty" mapstructure:"user_name"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" mapstructure:"password"`
}

func (c *Address) String() string {
	return marshalConfig(c)
}
func (c *Address) GoString() string {
	return c.String()
}

type IpfsClusterConfig struct {
	Enabled         bool       `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"`                            // 是否启用
	Strategy        string     `json:"strategy,omitempty" yaml:"strategy,omitempty" mapstructure:"strategy"`                         // 策略
	TimeOutInterval int        `json:"timeout_interval,omitempty" yaml:"timeout_interval,omitempty" mapstructure:"timeout_interval"` // 超时时间
	LogLevel        string     `json:"log_level,omitempty" yaml:"log_level,omitempty" mapstructure:"log_level"`                      // 日志级别
	ReTries         int        `json:"retries,omitempty" yaml:"retries,omitempty" mapstructure:"retries"`                            // 重试次数
	Addresses       []*Address `json:"addresses,omitempty" yaml:"addresses,omitempty" mapstructure:"addresses"`
}

func (c *IpfsClusterConfig) String() string {
	return marshalConfig(&c)
}
func (c *IpfsClusterConfig) GoString() string {
	return c.String()
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
func (c *RedisConfig) String() string {
	return marshalConfig(c)
}
func (c *RedisConfig) GoString() string {
	return c.String()
}

// KvDatabaseConfig KV数据库配置结构体
type KvDatabaseConfig struct {
	Enabled bool   `json:"enabled,omitempty" yaml:"enabled,omitempty" mapstructure:"enabled"` // 是否启用
	DbType  string `json:"db_type,omitempty" yaml:"db_type,omitempty" mapstructure:"db_type"` // 数据库类型 (leveldb, pebble, badger)
	DbPath  string `json:"db_path,omitempty" yaml:"db_path,omitempty" mapstructure:"db_path"` // 数据库文件路径
}

// String 返回KvDatabaseConfig的字符串表示
// @return string KvDatabaseConfig的字符串表示
func (c *KvDatabaseConfig) String() string {
	return marshalConfig(c)
}
func (c *KvDatabaseConfig) GoString() string {
	return c.String()
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
func (c *AIConfig) String() string {
	return marshalConfig(c)
}
func (c *AIConfig) GoString() string {
	return c.String()
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
	return marshalConfig(c)
}
func (c *CertificateConfig) GoString() string {
	return c.String()
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
	return marshalConfig(c)
}
func (c *Config) GoString() string {
	return c.String()
}
