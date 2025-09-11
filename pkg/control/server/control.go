package server

import (
	"sync"

	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/websocket"

	"github.com/jianlu8023/gm-fabric-deployment/internal/web/router"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/job"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	dockerControl     *docker.Control
	configControl     *config.Control
	libp2pControl     *libp2p.Control
	grpcControl       *grpc.Control
	httpControl       *http.Control
	datasourceControl *datasource.Control
	loggerControl     *logger.Control
	jobControl        *job.Control
	websocketControl  *websocket.Control
	logger            *zap.SugaredLogger
	once              sync.Once
}

func NewServerControlFromFile() (*Control, error) {
	control := new(Control)

	configControl, err := config.NewConfigControl()
	if err != nil {
		return nil, err
	}

	control.configControl = configControl

	// 检查Logger配置
	if configControl.GetLoggerConfig() == nil {
		return nil, ErrNoLoggerConfig
	}

	// 创建Logger控制器
	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	control.loggerControl = loggerControl
	control.logger = loggerControl.GenLogger("server")

	// 检查并创建GRPC控制器
	grpcConfig := configControl.GetGrpcConfig()
	if grpcConfig != nil && grpcConfig.Enabled {
		grpcControl, err := grpc.NewGrpcControl(grpcConfig, loggerControl)
		if err != nil {
			return nil, err
		}
		control.grpcControl = grpcControl
	}

	// 检查并创建Libp2p控制器
	libp2pConfig := configControl.GetLibp2pConfig()
	if libp2pConfig != nil && libp2pConfig.Enabled {
		libp2pControl, err := libp2p.NewLibp2pControl(libp2pConfig, loggerControl)
		if err != nil {
			return nil, err
		}
		control.libp2pControl = libp2pControl
	}

	// 检查并创建DataSource控制器
	dataSourceConfig := configControl.GetDataSourceConfig()
	if dataSourceConfig != nil && dataSourceConfig.Enabled {
		dataSourceControl, err := datasource.NewDataSourceControl(dataSourceConfig, loggerControl)
		if err != nil {
			return nil, err
		}
		control.datasourceControl = dataSourceControl
	}

	// 创建Job控制器
	jobControl := job.NewJobControl(loggerControl)
	control.jobControl = jobControl

	// 检查并创建Docker控制器
	dockerConfig := configControl.GetDockerConfig()
	if dockerConfig != nil && dockerConfig.Enabled {
		dockerControl, err := docker.NewDockerControl(dockerConfig, loggerControl)
		if err != nil {
			return nil, err
		}
		control.dockerControl = dockerControl
	}

	// 检查并创建HTTP控制器
	webConfig := configControl.GetWebConfig()
	if webConfig != nil && webConfig.Enabled {
		webServerControl := http.NewWebServerControl(webConfig, loggerControl)
		control.httpControl = webServerControl

		websocketControl := websocket.NewWebsocketControl(webConfig, loggerControl)
		control.websocketControl = websocketControl
	}

	return control, nil
}

func (c *Control) GetDockerControl() *docker.Control {
	return c.dockerControl
}

func (c *Control) GetConfigControl() *config.Control {
	return c.configControl
}

func (c *Control) GetLibp2pControl() *libp2p.Control {
	return c.libp2pControl
}

func (c *Control) GetGrpcControl() *grpc.Control {
	return c.grpcControl
}

func (c *Control) GetHttpControl() *http.Control {
	return c.httpControl
}

func (c *Control) GetDatasourceControl() *datasource.Control {
	return c.datasourceControl
}

func (c *Control) GetJobControl() *job.Control {
	return c.jobControl
}

func (c *Control) GetLoggerControl() *logger.Control {
	return c.loggerControl
}

func (c *Control) GetWebsocketControl() *websocket.Control {
	return c.websocketControl
}

func NewServerControl(dockerControl *docker.Control,
	configControl *config.Control,
	libp2pControl *libp2p.Control,
	grpcControl *grpc.Control,
	httpControl *http.Control,
	datasourceControl *datasource.Control,
	loggerControl *logger.Control,
	jobControl *job.Control,
	websocketControl *websocket.Control,
) *Control {
	serverLogger := loggerControl.GenLogger("server")
	serverLogger.Infof("[control] starting new server control...")
	return &Control{
		logger:            serverLogger,
		dockerControl:     dockerControl,
		configControl:     configControl,
		libp2pControl:     libp2pControl,
		grpcControl:       grpcControl,
		httpControl:       httpControl,
		datasourceControl: datasourceControl,
		loggerControl:     loggerControl,
		jobControl:        jobControl,
		websocketControl:  websocketControl,
	}
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.configControl != nil {
			c.logger.Debugf("[control] starting up config server...")
			c.configControl.StartUp(failedFunc)
		}

		if c.loggerControl != nil {
			c.logger.Debugf("[control] starting up logger server...")
			c.loggerControl.StartUp(failedFunc)
		}

		if c.datasourceControl != nil {
			c.logger.Debugf("[control] starting up datasource server...")
			c.datasourceControl.StartUp(failedFunc)
		}

		if c.grpcControl != nil {
			c.logger.Debugf("[control] starting up grpc server...")
			c.grpcControl.StartUp(failedFunc)
		}

		if c.libp2pControl != nil {
			c.logger.Debugf("[control] starting up libp2p server...")
			c.libp2pControl.StartUp(failedFunc)
		}

		if c.dockerControl != nil {
			c.logger.Debugf("[control] starting up docker server...")
			c.dockerControl.StartUp(failedFunc)
		}

		if c.jobControl != nil {
			c.logger.Debugf("[control] starting up job server...")
			c.jobControl.StartUp(failedFunc)
		}

		if c.httpControl != nil {
			c.logger.Debugf("[control] starting up http server...")
			c.httpControl.RegisterRouter(router.NewRouter(
				c.loggerControl,
				c.libp2pControl,
				c.grpcControl,
				c.dockerControl,
				c.datasourceControl,
				c.websocketControl,
				c.httpControl,
			))
			c.httpControl.StartUp(failedFunc)
		}

		if c.websocketControl != nil {
			c.logger.Debugf("[control] starting up websocket server...")
			c.websocketControl.StartUp(failedFunc)
		}

		c.logger.Infof("[control] all server started up successfully...")
	})
}

func (c *Control) Shutdown() error {
	if c.httpControl != nil {
		c.logger.Debugf("[control] shutting down http server...")
		if err := c.httpControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown http server err: %v", err)
			return err
		}
	}

	if c.jobControl != nil {
		c.logger.Debugf("[control] shutting down job server...")
		_ = c.jobControl.Shutdown()
	}

	if c.libp2pControl != nil {
		c.logger.Debugf("[control] shutting down libp2p server...")
		if err := c.libp2pControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown libp2p server err: %v", err)
			return err
		}
	}

	if c.grpcControl != nil {
		c.logger.Debugf("[control] shutting down grpc server...")
		_ = c.grpcControl.Shutdown()
	}

	if c.dockerControl != nil {
		c.logger.Debugf("[control] shutting down docker server...")
		if err := c.dockerControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown docker server err: %v", err)
			return err
		}
	}

	if c.datasourceControl != nil {
		c.logger.Debugf("[control] shutting down datasource server...")
		if err := c.datasourceControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown datasource server err: %v", err)
			return err
		}
	}

	// 不会有问题的关闭
	if c.loggerControl != nil {
		c.logger.Debugf("[control] shutting down logger server...")
		_ = c.loggerControl.Shutdown()
	}

	if c.configControl != nil {
		c.logger.Debugf("[control] shutting down config server...")
		_ = c.configControl.Shutdown()
	}

	return nil
}
