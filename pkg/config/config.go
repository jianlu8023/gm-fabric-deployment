package config

import (
	"encoding/json"
)

type LoggerConfig struct {
	DefaultLogLevel string            `json:"default_log_level" yaml:"default_log_level" mapstrcuture:"default_log_level"`
	PrintFormat     string            `yaml:"print_format" json:"print_format" mapstrcuture:"print_format"`
	FilePath        string            `json:"file_path" yaml:"file_path" mapstructure:"file_path"`
	MaxAge          int               `json:"max_age" yaml:"max_age" mapstructure:"max_age"`
	RotationTime    int               `json:"rotation_time" yaml:"rotation_time" mapstructure:"rotation_time"`
	LoggerLevel     map[string]string `yaml:"logger_level" yaml:"logger_level" mapstrcuture:"logger_level"`
}

type HttpServerConfig struct {
	Address     string `yaml:"address" mapstructure:"address"`
	ContextPath string `json:"context_path" mapstructure:"context_path"`
	RunMode     string `yaml:"run_mode" mapstructure:"run_mode"`
	TlsEnabled  bool   `yaml:"tls_enabled" mapstructure:"tls_enabled"`
	TlsCertFile string `yaml:"tls_cert_file" mapstructure:"tls_cert_file"`
	TlsKeyFile  string `yaml:"tls_key_file" mapstructure:"tls_key_file"`
}

func (h HttpServerConfig) String() string {
	bytes, _ := json.MarshalIndent(h, " ", "  ")
	return string(bytes)
}

type GrpcServerConfig struct {
	Host           string `mapstructure:"host" yaml:"host"`
	MaxRecvMsgSize int    `mapstructure:"max_recv_msg_size" yaml:"max_recv_msg_size"`
	MaxSendMsgSize int    `mapstructure:"max_send_msg_size" yaml:"max_send_msg_size"`
	ChunkSize      int    `mapstructure:"chunk_size" yaml:"chunk_size"`
	TlsEnabled     bool   `yaml:"tls_enabled" mapstructure:"tls_enabled"`
	TlsCertFile    string `yaml:"tls_cert_file" mapstructure:"tls_cert_file"`
	TlsKeyFile     string `yaml:"tls_key_file" mapstructure:"tls_key_file"`
	TlsRCACertFile string `yaml:"tls_rca_cert_file" mapstructure:"tls_rca_cert_file"`
}

func (g GrpcServerConfig) String() string {
	bytes, _ := json.MarshalIndent(g, " ", "  ")
	return string(bytes)
}

type GrpcClientConfig struct {
	Host               string `yaml:"host" mapstructure:"host"`
	MaxCallRecvMsgSize int    `mapstructure:"max_call_recv_msg_size" yaml:"max_call_recv_msg_size"`
	MaxCallSendMsgSize int    `mapstructure:"max_call_send_msg_size" yaml:"max_call_send_msg_size"`
	ChunkSize          int    `mapstructure:"chunk_size" yaml:"chunk_size"`
	CallTimeout        int    `mapstructure:"call_timeout" yaml:"call_timeout"`
	TlsEnabled         bool   `yaml:"tls_enabled" mapstructure:"tls_enabled"`
	TlsCertFile        string `yaml:"tls_cert_file" mapstructure:"tls_cert_file"`
	TlsKeyFile         string `yaml:"tls_key_file" mapstructure:"tls_key_file"`
	TlsRCACertFile     string `yaml:"tls_rca_cert_file" mapstructure:"tls_rca_cert_file"`
}

func (g GrpcClientConfig) String() string {
	bytes, _ := json.MarshalIndent(g, " ", "  ")
	return string(bytes)
}

type GrpcConfig struct {
	Server *GrpcServerConfig `yaml:"server"`
	Client *GrpcClientConfig `yaml:"client"`
}

func (g GrpcConfig) String() string {
	bytes, _ := json.MarshalIndent(g, " ", "  ")
	return string(bytes)
}
