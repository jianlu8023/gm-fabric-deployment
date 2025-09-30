package tracer

import (
	"github.com/jianlu8023/golang-example/pkg/control/config"
)

func getDefaultConfig() *config.TracerConfig {
	return &config.TracerConfig{
		Enabled:                true,
		ExporterServiceName:    "golang-example",
		ExporterFilePath:       "tracer.json",
		ExporterOTELInsecure:   true,
		ExporterOTELProtocol:   "grpc",
		TracesExporter:         "file",
		ExporterZipkinEndpoint: "",
		ExporterOTELEndpoint:   "http://127.0.0.1:4317",
	}
}
