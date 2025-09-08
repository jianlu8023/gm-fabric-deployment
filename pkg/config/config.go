package config

import (
	"encoding/json"
)

type LoggerConfig struct {
	DefaultLogLevel string            `json:"default_log_level,omitempty" yaml:"default_log_level,omitempty" mapstructure:"default_log_level"`
	StackLogLevel   string            `json:"stack_log_level,omitempty" yaml:"stack_log_level,omitempty" mapstructure:"stack_log_level"`
	PrintFormat     string            `json:"print_format,omitempty" yaml:"print_format,omitempty" mapstructure:"print_format"`
	FilePath        string            `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
	MaxAge          int               `json:"max_age,omitempty" yaml:"max_age,omitempty" mapstructure:"max_age"`
	RotationTime    int               `json:"rotation_time,omitempty" yaml:"rotation_time,omitempty" mapstructure:"rotation_time"`
	LoggerLevel     map[string]string `json:"logger_level,omitempty" yaml:"logger_level,omitempty" mapstructure:"logger_level"`
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
	GrpcConfig   *GrpcConfig       `json:"grpc_config,omitempty" yaml:"grpc_config,omitempty" mapstructure:"grpc"`
	LoggerConfig *LoggerConfig     `json:"logger_config,omitempty" yaml:"logger_config,omitempty" mapstructure:"logger"`
	HttpConfig   *HttpServerConfig `json:"http_config,omitempty" yaml:"http_config,omitempty" mapstructure:"http"`
	Libp2pConfig *Libp2pConfig     `json:"libp2p_config,omitempty" yaml:"libp2p_config,omitempty" mapstructure:"libp2p"`
}

func (c Config) String() string {
	bytes, _ := json.MarshalIndent(c, "", " ")
	return string(bytes)
}
