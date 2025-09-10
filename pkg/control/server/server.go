package server

import (
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/datasource"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/docker"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/grpc"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/http"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/libp2p"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"go.uber.org/zap"
	"sync"
)

type Control struct {
	dockerControl     *docker.Control
	configControl     *config.Control
	libp2pControl     *libp2p.Control
	grpcControl       *grpc.Control
	httpControl       *http.Control
	datasourceControl *datasource.Control
	loggerControl     *logger.Control
	logger            *zap.SugaredLogger
	once              sync.Once
}

func NewServerControl(dockerControl *docker.Control,
	configControl *config.Control,
	libp2pControl *libp2p.Control,
	grpcControl *grpc.Control,
	httpControl *http.Control,
	datasourceControl *datasource.Control,
	loggerControl *logger.Control,
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
	}
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		c.logger.Debugf("[control] starting up config server...")
		if c.configControl != nil {
			c.configControl.StartUp(failedFunc)
		}

		c.logger.Debugf("[control] starting up logger server...")
		if c.loggerControl != nil {
			c.loggerControl.StartUp(failedFunc)
		}

		c.logger.Debugf("[control] starting up datasource server...")
		if c.datasourceControl != nil {
			c.datasourceControl.StartUp(failedFunc)
		}

		c.logger.Debugf("[control] starting up grpc server...")
		if c.grpcControl != nil {
			c.grpcControl.StartUp(failedFunc)
		}

		c.logger.Debugf("[control] starting up libp2p server...")
		if c.libp2pControl != nil {
			c.libp2pControl.StartUp(failedFunc)
		}

		c.logger.Debugf("[control] starting up docker server...")
		if c.dockerControl != nil {
			c.dockerControl.StartUp(failedFunc)
		}

		c.logger.Debugf("[control] starting up http server...")
		if c.httpControl != nil {
			c.httpControl.RegisterRouter(http.NewRouter())

			c.httpControl.StartUp(failedFunc)
		}
		c.logger.Infof("[control] all server started up successfully...")
	})
}

func (c *Control) Shutdown() error {
	// 不会有问题的关闭
	c.logger.Debugf("[control] shutting down config server...")
	_ = c.configControl.Shutdown()
	c.logger.Debugf("[control] shutting down logger server...")
	_ = c.loggerControl.Shutdown()

	if c.datasourceControl != nil {
		c.logger.Debugf("[control] shutting down datasource server...")
		if err := c.datasourceControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown datasource server err: %v", err)
			return err
		}
	}

	if c.grpcControl != nil {
		c.logger.Debugf("[control] shutting down grpc server...")
		_ = c.grpcControl.Shutdown()
	}

	if c.libp2pControl != nil {
		c.logger.Debugf("[control] shutting down libp2p server...")
		if err := c.libp2pControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown libp2p server err: %v", err)
			return err
		}
	}

	if c.dockerControl != nil {
		c.logger.Debugf("[control] shutting down docker server...")
		if err := c.dockerControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown docker server err: %v", err)
			return err
		}
	}

	if c.httpControl != nil {
		c.logger.Debugf("[control] shutting down http server...")
		if err := c.httpControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown http server err: %v", err)
			return err
		}
	}
	return nil
}
