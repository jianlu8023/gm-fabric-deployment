package config

import (
	"encoding/json"
)

type LoggerConfig struct {
	DefaultLogLevel string            `json:"default_log_level" yaml:"default_log_level" mapstructure:"default_log_level"`
	PrintFormat     string            `json:"print_format" yaml:"print_format" mapstructure:"print_format"`
	FilePath        string            `json:"file_path" yaml:"file_path" mapstructure:"file_path"`
	MaxAge          int               `json:"max_age" yaml:"max_age" mapstructure:"max_age"`
	RotationTime    int               `json:"rotation_time" yaml:"rotation_time" mapstructure:"rotation_time"`
	LoggerLevel     map[string]string `json:"logger_level" yaml:"logger_level" mapstructure:"logger_level"`
}

type HttpServerConfig struct {
	Address     string `yaml:"address" json:"address" mapstructure:"address" `
	ContextPath string `yaml:"context_path" json:"context_path" mapstructure:"context_path"`
	RunMode     string `yaml:"run_mode" json:"run_mode" mapstructure:"run_mode" `
	TlsEnabled  bool   `yaml:"tls_enabled" json:"tls_enabled" mapstructure:"tls_enabled" `
	TlsCertFile string `yaml:"tls_cert_file" json:"tls_cert_file" mapstructure:"tls_cert_file" `
	TlsKeyFile  string `yaml:"tls_key_file" json:"tls_key_file" mapstructure:"tls_key_file"`
}

func (h HttpServerConfig) String() string {
	bytes, _ := json.MarshalIndent(h, " ", "  ")
	return string(bytes)
}

type GrpcServerConfig struct {
	Host           string `yaml:"host" json:"host" mapstructure:"host" `
	MaxRecvMsgSize int    `yaml:"max_recv_msg_size" json:"max_recv_msg_size" mapstructure:"max_recv_msg_size"`
	MaxSendMsgSize int    `yaml:"max_send_msg_size" json:"max_send_msg_size" mapstructure:"max_send_msg_size"`
	ChunkSize      int    `yaml:"chunk_size" json:"chunk_size" mapstructure:"chunk_size"`
	TlsEnabled     bool   `yaml:"tls_enabled" json:"tls_enabled" mapstructure:"tls_enabled"`
	TlsCertFile    string `yaml:"tls_cert_file" json:"tls_cert_file" mapstructure:"tls_cert_file"`
	TlsKeyFile     string `yaml:"tls_key_file" json:"tls_key_file" mapstructure:"tls_key_file"`
	TlsRCACertFile string `yaml:"tls_rca_cert_file" json:"tls_rca_cert_file" mapstructure:"tls_rca_cert_file"`
}

func (g GrpcServerConfig) String() string {
	bytes, _ := json.MarshalIndent(g, " ", "  ")
	return string(bytes)
}

type GrpcClientConfig struct {
	Host               string `yaml:"host" json:"host" mapstructure:"host"`
	MaxCallRecvMsgSize int    `yaml:"max_call_recv_msg_size" json:"max_call_recv_msg_size" mapstructure:"max_call_recv_msg_size"`
	MaxCallSendMsgSize int    `yaml:"max_call_send_msg_size" json:"max_call_send_msg_size" mapstructure:"max_call_send_msg_size" `
	ChunkSize          int    `yaml:"chunk_size" json:"chunk_size" mapstructure:"chunk_size" `
	CallTimeout        int    `yaml:"call_timeout" json:"call_timeout" mapstructure:"call_timeout" `
	TlsEnabled         bool   `yaml:"tls_enabled" json:"tls_enabled" mapstructure:"tls_enabled"`
	TlsCertFile        string `yaml:"tls_cert_file" json:"tls_cert_file" mapstructure:"tls_cert_file"`
	TlsKeyFile         string `yaml:"tls_key_file" json:"tls_key_file" mapstructure:"tls_key_file"`
	TlsRCACertFile     string `yaml:"tls_rca_cert_file" json:"tls_rca_cert_file" mapstructure:"tls_rca_cert_file"`
}

func (g GrpcClientConfig) String() string {
	bytes, _ := json.MarshalIndent(g, " ", "  ")
	return string(bytes)
}

type GrpcConfig struct {
	Server *GrpcServerConfig `yaml:"server" json:"server" mapstructure:"server"`
	Client *GrpcClientConfig `yaml:"client" json:"client" mapstructure:"client"`
}

func (g GrpcConfig) String() string {
	bytes, _ := json.MarshalIndent(g, " ", "  ")
	return string(bytes)
}

type Config struct {
	GrpcConfig   *GrpcConfig       `json:"grpc_config" yaml:"grpc_config" mapstructure:"grpc"`
	LoggerConfig *LoggerConfig     `json:"logger_config" yaml:"logger_config" mapstructure:"logger"`
	HttpConfig   *HttpServerConfig `json:"http_config" yaml:"http_config" mapstructure:"http"`
}

func (c Config) String() string {
	bytes, _ := json.MarshalIndent(c, " ", "  ")
	return string(bytes)
}
