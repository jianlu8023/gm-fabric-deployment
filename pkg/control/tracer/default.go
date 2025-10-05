package tracer

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

const (
	ExporterOtlp             = "otlp"
	ExporterZipkin           = "zipkin"
	ExporterNone             = "none"
	ExporterFile             = "file"
	ExporterBlack            = ""
	ExporterStdout           = "stdout"
	ExporterOtlpProtocolGrpc = "grpc"
	ExporterOtlpProtocolHttp = "http/protobuf"
)

func getDefaultConfig() *config.TracerConfig {
	return &config.TracerConfig{
		Enabled:      true,
		ServiceName:  "golang-example",
		LogInConsole: false,
		Tracer: struct {
			Exporters      string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
			OTELProtocol   string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
			OTELEndpoint   string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
			OTELInsecure   bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
			ZipkinEndpoint string `json:"zipkin_endpoint,omitempty" yaml:"zipkin_endpoint,omitempty" mapstructure:"zipkin_endpoint"`
			FilePath       string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
		}{
			Exporters:      "none,file",
			OTELProtocol:   "grpc",
			OTELEndpoint:   "http://127.0.0.1:4317",
			OTELInsecure:   true,
			ZipkinEndpoint: "",
			FilePath:       "tracer.json",
		},
		Logger: struct {
			Exporters    string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
			OTELProtocol string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
			OTELEndpoint string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
			OTELInsecure bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
			FilePath     string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
		}{
			Exporters:    "none,file",
			OTELProtocol: "http/protobuf",
			OTELEndpoint: "http://127.0.0.1:4318",
			OTELInsecure: true,
			FilePath:     "logger.json",
		},
		Meter: struct {
			Exporters    string `json:"exporters,omitempty" yaml:"exporters,omitempty" mapstructure:"exporters"`
			OTELProtocol string `json:"otel_protocol,omitempty" yaml:"otel_protocol,omitempty" mapstructure:"otel_protocol"`
			OTELEndpoint string `json:"otel_endpoint,omitempty" yaml:"otel_endpoint,omitempty" mapstructure:"otel_endpoint"`
			OTELInsecure bool   `json:"otel_insecure,omitempty" yaml:"otel_insecure,omitempty" mapstructure:"otel_insecure"`
			FilePath     string `json:"file_path,omitempty" yaml:"file_path,omitempty" mapstructure:"file_path"`
		}{
			Exporters:    "none,file",
			OTELProtocol: "http/protobuf",
			OTELEndpoint: "http://127.0.0.1:4318",
			OTELInsecure: true,
			FilePath:     "meter.json",
		},
	}
}
