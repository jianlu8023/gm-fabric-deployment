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
		Tracer: &config.Tracer{
			Exporters:      "none",
			OTELProtocol:   "grpc",
			OTELEndpoint:   "http://127.0.0.1:4317",
			OTELInsecure:   true,
			ZipkinEndpoint: "",
			FilePath:       "tracer.json",
		},
		Logger: &config.Logger{
			Exporters:    "none",
			OTELProtocol: "http/protobuf",
			OTELEndpoint: "http://127.0.0.1:4318",
			OTELInsecure: true,
			FilePath:     "logger.json",
		},
		Metrics: &config.Metrics{
			Exporters:    "none",
			OTELProtocol: "http/protobuf",
			OTELEndpoint: "http://127.0.0.1:4318",
			OTELInsecure: true,
			FilePath:     "metrics.json",
		},
	}
}
