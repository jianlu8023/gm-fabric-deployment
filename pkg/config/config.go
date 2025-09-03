package config

type GrpcServerConfig struct {
	Host string `yaml:"host"`
}

type GrpcClientConfig struct {
	Host string `yaml:"host"`
}

type GrpcConfig struct {
	Server *GrpcServerConfig `yaml:"server"`
	Client *GrpcClientConfig `yaml:"client"`
}
