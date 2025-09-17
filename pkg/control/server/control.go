package server

import (
	"fmt"
	"github.com/jianlu8023/golang-example/pkg/control/ants"
	"github.com/jianlu8023/golang-example/pkg/control/fabricca"
	"github.com/jianlu8023/golang-example/version"
	"os"
	"sync"

	"github.com/jianlu8023/golang-example/pkg/control/flags"

	"github.com/jianlu8023/golang-example/pkg/control/websocket"

	"github.com/jianlu8023/golang-example/pkg/control/captcha"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/grpc"
	"github.com/jianlu8023/golang-example/pkg/control/http"
	"github.com/jianlu8023/golang-example/pkg/control/job"
	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	flagsControl      *flags.Control
	configControl     *config.Control
	loggerControl     *logger.Control
	grpcControl       *grpc.Control
	libp2pControl     *libp2p.Control
	dockerControl     *docker.Control
	datasourceControl *datasource.Control
	httpControl       *http.Control
	websocketControl  *websocket.Control
	jobControl        *job.Control
	captchaControl    *captcha.Control
	antsControl       *ants.Control
	fabricCAControl   *fabricca.Control
	logger            *zap.SugaredLogger
	once              sync.Once
	mutex             sync.RWMutex
}

// NewServerControlFromFile 从配置文件创建服务器控制器
// @return *Control 服务器控制器实例
// @return error 创建过程中可能产生的错误
func NewServerControlFromFile() (*Control, error) {
	control := new(Control)
	// control.mutex.Lock()
	// defer control.mutex.Unlock()

	flagsControl := flags.NewFlagsControl(version.Version)
	control.flagsControl = flagsControl
	flagsControl.StartUp(func(err error) {
		fmt.Printf("flgas control start up failed: %v\n", err)
		os.Exit(-1)
	})

	configControl := config.NewConfigControl(flagsControl)
	control.configControl = configControl
	configControl.StartUp(func(err error) {
		fmt.Printf("config control start up failed: %v\n", err)
		os.Exit(-2)
	})

	// 检查Logger配置
	if configControl.GetLoggerConfig() == nil {
		return nil, fmt.Errorf("not found logger config")
	}

	// 创建Logger控制器
	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	control.loggerControl = loggerControl
	control.logger = loggerControl.GenLogger(logger.ModuleServer)

	// 检查并创建GRPC控制器
	grpcConfig := configControl.GetGrpcConfig()
	if grpcConfig != nil && grpcConfig.Enabled {
		grpcControl, err := grpc.NewGrpcControl(grpcConfig, control.GetLoggerControl())
		if err != nil {
			return nil, err
		}
		control.grpcControl = grpcControl
	}

	// 检查并创建Libp2p控制器
	libp2pConfig := configControl.GetLibp2pConfig()
	if libp2pConfig != nil && libp2pConfig.Enabled {
		libp2pControl, err := libp2p.NewLibp2pControl(libp2pConfig, control.GetLoggerControl())
		if err != nil {
			return nil, err
		}
		control.libp2pControl = libp2pControl
	}

	// 检查并创建DataSource控制器
	dataSourceConfig := configControl.GetDataSourceConfig()
	if dataSourceConfig != nil && dataSourceConfig.Enabled {
		dataSourceControl, err := datasource.NewDataSourceControl(dataSourceConfig, control.GetLoggerControl())
		if err != nil {
			return nil, err
		}
		control.datasourceControl = dataSourceControl
	}

	antsPoolConfig := configControl.GetAntsPoolConfig()
	if antsPoolConfig != nil && antsPoolConfig.Enabled {
		antsPoolControl := ants.NewAntsPoolControl(antsPoolConfig, control.GetLoggerControl())
		control.antsControl = antsPoolControl
	}

	// if control.GetAntsPoolControl() == nil {
	// 	return nil, fmt.Errorf("ants pool control is nil")
	// }

	if control.GetAntsPoolControl() != nil {
		// 创建Job控制器
		jobControl := job.NewJobControl(control.GetLoggerControl(), control.GetAntsPoolControl())
		control.jobControl = jobControl
	}

	// 检查并创建Docker控制器
	dockerConfig := configControl.GetDockerConfig()
	if dockerConfig != nil && dockerConfig.Enabled {
		dockerControl, err := docker.NewDockerControl(dockerConfig, control.GetLoggerControl())
		if err != nil {
			return nil, err
		}
		control.dockerControl = dockerControl

		// 创建fabricca控制器
		fabricCAConfig := configControl.GetFabricCAConfig()
		if fabricCAConfig != nil {
			fabricCAControl, err := fabricca.NewFabricCAControl(fabricCAConfig, control.GetLoggerControl(), control.GetDockerControl())
			if err != nil {
				return nil, err
			}
			control.fabricCAControl = fabricCAControl
		}

	}

	// 检查并创建验证码控制器
	captchaConfig := configControl.GetCaptchaConfig()
	if captchaConfig != nil && captchaConfig.Enabled {
		captchaControl, err := captcha.NewCaptchaControl(captchaConfig, control.GetLoggerControl())
		if err != nil {
			return nil, err
		}
		control.captchaControl = captchaControl
	}

	// 检查并创建HTTP控制器
	webConfig := configControl.GetWebConfig()
	if webConfig != nil && webConfig.Enabled {
		webServerControl := http.NewWebServerControl(webConfig, control.GetLoggerControl())
		control.httpControl = webServerControl

		websocketControl := websocket.NewWebsocketControl(webConfig, control.GetLoggerControl())
		control.websocketControl = websocketControl
	}

	return control, nil
}

// GetDockerControl 获取Docker控制器
// @return *docker.Control Docker控制器实例
func (c *Control) GetDockerControl() *docker.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.dockerControl
}

// GetConfigControl 获取配置控制器
// @return *config.Control 配置控制器实例
func (c *Control) GetConfigControl() *config.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.configControl
}

// GetLibp2pControl 获取Libp2p控制器
// @return *libp2p.Control Libp2p控制器实例
func (c *Control) GetLibp2pControl() *libp2p.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.libp2pControl
}

// GetGrpcControl 获取gRPC控制器
// @return *grpc.Control gRPC控制器实例
func (c *Control) GetGrpcControl() *grpc.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.grpcControl
}

// GetHttpControl 获取HTTP控制器
// @return *http.Control HTTP控制器实例
func (c *Control) GetHttpControl() *http.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.httpControl
}

// GetDatasourceControl 获取数据源控制器
// @return *datasource.Control 数据源控制器实例
func (c *Control) GetDatasourceControl() *datasource.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.datasourceControl
}

// GetJobControl 获取作业控制器
// @return *job.Control 作业控制器实例
func (c *Control) GetJobControl() *job.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.jobControl
}

// GetLoggerControl 获取日志控制器
// @return *logger.Control 日志控制器实例
func (c *Control) GetLoggerControl() *logger.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.loggerControl
}

// GetWebsocketControl 获取WebSocket控制器
// @return *websocket.Control WebSocket控制器实例
func (c *Control) GetWebsocketControl() *websocket.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.websocketControl
}

// GetCaptchaControl 获取验证码控制器
// @return *captcha.Control 验证码控制器实例
func (c *Control) GetCaptchaControl() *captcha.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.captchaControl
}

// GetFlagsControl 获取flags控制器
// @return *flags.Control flags控制器
func (c *Control) GetFlagsControl() *flags.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flagsControl
}

func (c *Control) GetAntsPoolControl() *ants.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.antsControl
}

func (c *Control) GetFabricCAControl() *fabricca.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.fabricCAControl
}

// NewServerControl 创建服务器控制器
// @param dockerControl *docker.Control Docker控制器
// @param configControl *config.Control 配置控制器
// @param libp2pControl *libp2p.Control Libp2p控制器
// @param grpcControl *grpc.Control gRPC控制器
// @param httpControl *http.Control HTTP控制器
// @param datasourceControl *datasource.Control 数据源控制器
// @param loggerControl *logger.Control 日志控制器
// @param jobControl *job.Control 作业控制器
// @param websocketControl *websocket.Control WebSocket控制器
// @param flagsControl *flags.Control flags控制器
// @return *Control 服务器控制器实例
func NewServerControl(dockerControl *docker.Control,
	configControl *config.Control,
	libp2pControl *libp2p.Control,
	grpcControl *grpc.Control,
	httpControl *http.Control,
	datasourceControl *datasource.Control,
	loggerControl *logger.Control,
	jobControl *job.Control,
	websocketControl *websocket.Control,
	flagsControl *flags.Control,
	captchaControl *captcha.Control,
	antsControl *ants.Control,
	fabricCAControl *fabricca.Control,
) *Control {
	serverLogger := loggerControl.GenLogger(logger.ModuleServer)
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
		flagsControl:      flagsControl,
		captchaControl:    captchaControl,
		antsControl:       antsControl,
		fabricCAControl:   fabricCAControl,
	}
}

// StartUp 启动所有服务器组件
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {

		if c.flagsControl != nil {
			c.logger.Debugf("[control] starting up flags server...")
			c.flagsControl.StartUp(failedFunc)
		}

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

		if c.fabricCAControl != nil {
			c.logger.Debugf("[control] starting up fabric ca server...")
			c.fabricCAControl.StartUp(failedFunc)
		}

		if c.antsControl != nil {
			c.logger.Debugf("[control] starting up ants pool server...")
			c.antsControl.StartUp(failedFunc)
		}

		if c.jobControl != nil {
			c.logger.Debugf("[control] starting up job server...")
			c.jobControl.StartUp(failedFunc)
		}

		if c.captchaControl != nil {
			c.logger.Debugf("[control] starting up captcha server...")
			c.captchaControl.StartUp(failedFunc)
		}

		if c.websocketControl != nil {
			c.logger.Debugf("[control] starting up websocket server...")
			c.websocketControl.StartUp(failedFunc)
		}

		if c.httpControl != nil {
			c.logger.Debugf("[control] starting up http server...")
			// if runtime.GOOS == runtime.GOOS &&
			// 	runtime.GOARCH == runtime.GOARCH {
			// 	stack := make([]uintptr, 10)
			// 	length := runtime.Callers(1, stack)
			// 	frames := runtime.CallersFrames(stack[:length])
			// 	frame, _ := frames.Next()
			// 	if frame.Function == "main.main" {
			// 		fmt.Printf("this is th main project...\n\n\n\n\n")
			// 	} else {
			// 		fmt.Printf("this is not the main project...\n\n\n\n\n")
			// 	}
			// }
			c.httpControl.StartUp(failedFunc)
		}

		c.logger.Infof("[control] all server started up successfully...")
	})
}

// Shutdown 关闭所有服务器组件
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {

	if c.jobControl != nil {
		c.logger.Debugf("[control] shutting down job server...")
		_ = c.jobControl.Shutdown()
	}

	if c.antsControl != nil {
		c.logger.Debugf("[control] shutting down ants pool server...")
		_ = c.antsControl.Shutdown()
	}

	if c.websocketControl != nil {
		c.logger.Debugf("[control] shutting down websocket server...")
		_ = c.websocketControl.Shutdown()
	}

	if c.captchaControl != nil {
		c.logger.Debugf("[control] shutting down captcha server...")
		_ = c.captchaControl.Shutdown()
	}

	if c.httpControl != nil {
		c.logger.Debugf("[control] shutting down http server...")
		if err := c.httpControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown http server err: %v", err)
			return err
		}
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

	if c.fabricCAControl != nil {
		c.logger.Debugf("[control] shutting down fabric ca server...")
		if err := c.fabricCAControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown fabric ca server err: %v", err)
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

	if c.flagsControl != nil {
		c.logger.Debugf("[control] shutting down flags server...")
		_ = c.flagsControl.Shutdown()
	}

	return nil
}
