package config

import (
	"encoding/json"
)

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
