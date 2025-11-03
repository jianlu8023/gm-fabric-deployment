package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/jianlu8023/golang-example/pkg/control/ai"
	"github.com/jianlu8023/golang-example/pkg/control/ants"
	"github.com/jianlu8023/golang-example/pkg/control/authz"
	"github.com/jianlu8023/golang-example/pkg/control/captcha"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/control/docker"
	"github.com/jianlu8023/golang-example/pkg/control/flags"
	"github.com/jianlu8023/golang-example/pkg/control/grpc"
	"github.com/jianlu8023/golang-example/pkg/control/http"
	"github.com/jianlu8023/golang-example/pkg/control/ipfs"
	"github.com/jianlu8023/golang-example/pkg/control/ipfscluster"
	"github.com/jianlu8023/golang-example/pkg/control/job"
	"github.com/jianlu8023/golang-example/pkg/control/kvdatabase"
	"github.com/jianlu8023/golang-example/pkg/control/libp2p"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/jianlu8023/golang-example/pkg/control/mfa"
	"github.com/jianlu8023/golang-example/pkg/control/redis"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/pkg/control/webrtc"
	"github.com/jianlu8023/golang-example/pkg/control/websocket"
	"github.com/jianlu8023/golang-example/version"
	"go.uber.org/zap"
)

// Control 服务器控制器结构体
// @description 管理所有服务器组件的控制器，包含各种子控制器的引用
type Control struct {
	flagsControl       *flags.Control       // flags控制器
	configControl      *config.Control      // 配置控制器
	loggerControl      *logger.Control      // 日志控制器
	grpcControl        *grpc.Control        // gRPC控制器
	libp2pControl      *libp2p.Control      // Libp2p控制器
	dockerControl      *docker.Control      // Docker控制器
	datasourceControl  *datasource.Control  // 数据源控制器
	httpControl        *http.Control        // HTTP控制器
	websocketControl   *websocket.Control   // WebSocket控制器
	jobControl         *job.Control         // 作业控制器
	captchaControl     *captcha.Control     // 验证码控制器
	antsControl        *ants.Control        // Ants线程池控制器
	authzControl       *authz.Control       // 权限控制器
	ipfsControl        *ipfs.Control        // IPFS控制器
	mfaControl         *mfa.Control         // MFA控制器
	tracerControl      *tracer.Control      // Tracer追踪控制器
	ipfsClusterControl *ipfscluster.Control // IPFS集群控制器
	redisControl       *redis.Control       // Redis控制器
	kvDatabaseControl  *kvdatabase.Control  // KvDatabase控制器
	webRTCControl      *webrtc.Control      // WebRTC控制器
	aiControl          *ai.Control          // AI控制器
	logger             *zap.SugaredLogger   // 日志记录器
	once               sync.Once            // 确保StartUp只执行一次
	mutex              sync.RWMutex         // 读写锁，保护控制器访问
	ctx                context.Context      // 上下文
}

// NewServerControlFromFile 从配置文件创建服务器控制器
// @description 根据配置文件初始化所有启用的服务器组件控制器
// @return *Control 服务器控制器实例
// @return error 创建过程中可能产生的错误
func NewServerControlFromFile() (*Control, error) {
	control := new(Control)
	// control.mutex.Lock()
	// defer control.mutex.Unlock()
	control.ctx = context.Background()

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
	// if configControl.GetLoggerConfig() == nil {
	// 	return nil, fmt.Errorf("not found logger config")
	// }

	// 创建Logger控制器
	loggerControl := logger.NewLoggerControl(configControl.GetLoggerConfig())
	control.loggerControl = loggerControl
	control.logger = loggerControl.GenLogger(logger.ModuleServer)

	tracerConfig := configControl.GetTracerConfig()
	if tracerConfig != nil && tracerConfig.Enabled {
		tracerControl, err := tracer.NewTracerControl(tracerConfig, control.GetLoggerControl(), control.ctx)
		if err != nil {
			control.logger.Errorf("[control] create tracer control failed: %v", err)
			return nil, err
		}
		control.tracerControl = tracerControl
	}

	// 检查并创建GRPC控制器
	grpcConfig := configControl.GetGrpcConfig()
	if grpcConfig != nil && grpcConfig.Enabled {
		grpcControl, err := grpc.NewGrpcControl(grpcConfig, control.GetLoggerControl(), grpc.WithTracer(control.GetTracerControl()))
		if err != nil {
			control.logger.Errorf("[control] create grpc control failed: %v", err)
			return nil, err
		}
		control.grpcControl = grpcControl
	}

	// 检查并创建Libp2p控制器
	libp2pConfig := configControl.GetLibp2pConfig()
	if libp2pConfig != nil && libp2pConfig.Enabled {
		libp2pControl, err := libp2p.NewLibp2pControl(libp2pConfig, control.GetLoggerControl())
		if err != nil {
			control.logger.Errorf("[control] create libp2p control failed: %v", err)
			return nil, err
		}
		control.libp2pControl = libp2pControl
	}

	// 检查并创建DataSource控制器
	dataSourceConfig := configControl.GetDataSourceConfig()
	if dataSourceConfig != nil && dataSourceConfig.Enabled {
		dataSourceControl, err := datasource.NewDataSourceControl(dataSourceConfig, control.GetLoggerControl(), datasource.WithTracer(control.GetTracerControl()))
		if err != nil {
			control.logger.Errorf("[control] create data source control failed: %v", err)
			return nil, err
		}
		control.datasourceControl = dataSourceControl
	}

	redisConfig := configControl.GetRedisConfig()
	if redisConfig != nil && redisConfig.Enabled {
		redisControl := redis.NewRedisControl(redisConfig, control.GetLoggerControl(), redis.WithTracer(control.GetTracerControl()))
		control.redisControl = redisControl
	}

	ipfsConfig := configControl.GetIpfsConfig()
	if ipfsConfig != nil && ipfsConfig.Enabled {
		ipfsControl := ipfs.NewIpfsControl(ipfsConfig, control.GetLoggerControl())
		control.ipfsControl = ipfsControl
	}

	ipfsClusterConfig := configControl.GetIpfsClusterConfig()
	if ipfsClusterConfig != nil && ipfsClusterConfig.Enabled {
		ipfsClusterControl, err := ipfscluster.NewIpfsClusterControl(ipfsClusterConfig, control.GetLoggerControl())
		if err != nil {
			control.logger.Errorf("[control] create ipfs cluster control failed: %v", err)
			return nil, err
		}
		control.ipfsClusterControl = ipfsClusterControl
	}

	antsPoolConfig := configControl.GetAntsPoolConfig()
	if antsPoolConfig != nil && antsPoolConfig.Enabled {
		antsPoolControl := ants.NewAntsPoolControl(antsPoolConfig, control.GetLoggerControl())
		control.antsControl = antsPoolControl
		// 创建Job控制器
		jobControl := job.NewJobControl(control.GetLoggerControl(), control.GetAntsPoolControl())
		control.jobControl = jobControl
	}

	// 检查并创建Docker控制器
	dockerConfig := configControl.GetDockerConfig()
	if dockerConfig != nil && dockerConfig.Enabled {
		dockerControl := docker.NewDockerControl(dockerConfig, control.GetLoggerControl())
		control.dockerControl = dockerControl

	}

	// 检查并创建验证码控制器
	captchaConfig := configControl.GetCaptchaConfig()
	if captchaConfig != nil && captchaConfig.Enabled {
		captchaControl := captcha.NewCaptchaControl(captchaConfig, control.GetLoggerControl())
		control.captchaControl = captchaControl
	}

	mfaConfig := configControl.GetMFAConfig()
	if mfaConfig != nil && mfaConfig.Enabled {
		mfaControl, err := mfa.NewMFAControl(mfaConfig, control.GetLoggerControl())
		if err != nil {
			control.logger.Errorf("[control] create mfa control failed: %v", err)
			return nil, err
		}
		control.mfaControl = mfaControl
	}

	// 检查并创建AI控制器
	aiConfig := configControl.GetAIConfig()
	if aiConfig != nil && aiConfig.Enabled {
		aiControl, err := ai.NewAIControl(aiConfig, control.GetLoggerControl())
		if err != nil {
			control.logger.Errorf("[control] create ai control failed: %v", err)
			return nil, err
		} else {
			control.aiControl = aiControl
		}
	}

	// 检查并创建权限控制器
	authzConfig := configControl.GetAuthzConfig()
	if authzConfig != nil && authzConfig.Enabled {
		authzControl := authz.NewAuthzControl(authzConfig, control.GetLoggerControl())

		control.authzControl = authzControl
	}

	// 检查并创建KvDatabase控制器
	kvDatabaseConfig := configControl.GetKvDatabaseConfig()
	if kvDatabaseConfig != nil && kvDatabaseConfig.Enabled {
		kvDatabaseControl := kvdatabase.NewKvDatabaseControl(kvDatabaseConfig, control.GetLoggerControl())
		control.kvDatabaseControl = kvDatabaseControl
	}

	webRTCConfig := configControl.GetWebRTCConfig()
	if webRTCConfig != nil && webRTCConfig.Enabled {
		webRTCControl, err := webrtc.NewWebRTCControl(webRTCConfig, control.GetLoggerControl())
		if err != nil {
			control.logger.Errorf("[control] create webrtc control failed: %v", err)
			return nil, err
		}
		control.webRTCControl = webRTCControl
	}

	// 检查并创建HTTP控制器
	webConfig := configControl.GetWebConfig()
	if webConfig != nil && webConfig.Enabled {
		webServerControl, err := http.NewWebServerControl(webConfig,
			control.GetLoggerControl(),
			http.WithTracer(control.GetTracerControl()),
			http.WithDefaultStaticFiles(),
		)
		if err != nil {
			control.logger.Errorf("[control] create http control failed: %v", err)
			return nil, err
		}
		control.httpControl = webServerControl

		websocketControl := websocket.NewWebsocketControl(webConfig, control.GetLoggerControl())
		control.websocketControl = websocketControl
	}

	return control, nil
}

// GetDockerControl 获取Docker控制器
// @description 获取Docker控制器实例，如果未初始化则返回nil
// @return *docker.Control Docker控制器实例
func (c *Control) GetDockerControl() *docker.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.dockerControl
}

// GetConfigControl 获取配置控制器
// @description 获取配置控制器实例，如果未初始化则返回nil
// @return *config.Control 配置控制器实例
func (c *Control) GetConfigControl() *config.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.configControl
}

// GetLibp2pControl 获取Libp2p控制器
// @description 获取Libp2p控制器实例，如果未初始化则返回nil
// @return *libp2p.Control Libp2p控制器实例
func (c *Control) GetLibp2pControl() *libp2p.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.libp2pControl
}

// GetGrpcControl 获取gRPC控制器
// @description 获取gRPC控制器实例，如果未初始化则返回nil
// @return *grpc.Control gRPC控制器实例
func (c *Control) GetGrpcControl() *grpc.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.grpcControl
}

// GetHttpControl 获取HTTP控制器
// @description 获取HTTP控制器实例，如果未初始化则返回nil
// @return *http.Control HTTP控制器实例
func (c *Control) GetHttpControl() *http.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.httpControl
}

// GetDatasourceControl 获取数据源控制器
// @description 获取数据源控制器实例，如果未初始化则返回nil
// @return *datasource.Control 数据源控制器实例
func (c *Control) GetDatasourceControl() *datasource.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.datasourceControl
}

// GetJobControl 获取作业控制器
// @description 获取作业控制器实例，如果未初始化则返回nil
// @return *job.Control 作业控制器实例
func (c *Control) GetJobControl() *job.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.jobControl
}

// GetLoggerControl 获取日志控制器
// @description 获取日志控制器实例，如果未初始化则返回nil
// @return *logger.Control 日志控制器实例
func (c *Control) GetLoggerControl() *logger.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.loggerControl
}

// GetWebsocketControl 获取WebSocket控制器
// @description 获取WebSocket控制器实例，如果未初始化则返回nil
// @return *websocket.Control WebSocket控制器实例
func (c *Control) GetWebsocketControl() *websocket.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.websocketControl
}

// GetCaptchaControl 获取验证码控制器
// @description 获取验证码控制器实例，如果未初始化则返回nil
// @return *captcha.Control 验证码控制器实例
func (c *Control) GetCaptchaControl() *captcha.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.captchaControl
}

// GetFlagsControl 获取flags控制器
// @description 获取flags控制器实例，如果未初始化则返回nil
// @return *flags.Control flags控制器
func (c *Control) GetFlagsControl() *flags.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flagsControl
}

// GetAntsPoolControl 获取Ants线程池控制器
// @description 获取Ants线程池控制器实例，如果未初始化则返回nil
// @return *ants.Control Ants线程池控制器实例
func (c *Control) GetAntsPoolControl() *ants.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.antsControl
}

// GetAuthzControl 获取权限控制器
// @description 获取权限控制器实例，如果未初始化则返回nil
// @return *authz.Control 权限控制器实例
func (c *Control) GetAuthzControl() *authz.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.authzControl
}

// GetIpfsControl 获取IPFS控制器
// @description 获取IPFS控制器实例，如果未初始化则返回nil
// @return *ipfs.Control IPFS控制器实例
func (c *Control) GetIpfsControl() *ipfs.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ipfsControl
}

// GetIpfsClusterControl 获取IPFS集群控制器
// @description 获取IPFS集群控制器实例，如果未初始化则返回nil
// @return *ipfscluster.Control IPFS集群控制器实例
func (c *Control) GetIpfsClusterControl() *ipfscluster.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.ipfsClusterControl
}

// GetMFAControl 获取MFA控制器
// @description 获取MFA控制器实例，如果未初始化则返回nil
// @return *mfa.Control MFA控制器实例
func (c *Control) GetMFAControl() *mfa.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.mfaControl
}

// GetTracerControl 获取Tracer追踪控制器
// @description 获取Tracer追踪控制器实例，如果未初始化则返回nil
// @return *tracer.Control Tracer追踪控制器实例
func (c *Control) GetTracerControl() *tracer.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.tracerControl
}

// GetRedisControl 获取Redis控制器
// @description 获取Redis控制器实例，如果未初始化则返回nil
// @return *redis.Control Redis控制器实例
func (c *Control) GetRedisControl() *redis.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.redisControl
}

// GetKvDatabaseControl 获取KvDatabase控制器
// @description 获取KvDatabase控制器实例，如果未初始化则返回nil
// @return *kvdatabase.Control KvDatabase控制器实例
func (c *Control) GetKvDatabaseControl() *kvdatabase.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.kvDatabaseControl
}

func (c *Control) GetWebRTCControl() *webrtc.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.webRTCControl
}

// GetAIControl 获取AI控制器
// @return *ai.Control AI控制器
func (c *Control) GetAIControl() *ai.Control {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.aiControl
}

// NewServerControl 创建服务器控制器
// @description 使用预初始化的各个子控制器创建服务器控制器实例
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
// @param captchaControl *captcha.Control 验证码控制器
// @param antsControl *ants.Control Ants线程池控制器
// @param authzControl *authz.Control 权限控制器
// @param ipfsControl *ipfs.Control IPFS控制器
// @param mfaControl *mfa.Control MFA控制器
// @param tracerControl *tracer.Control Tracer追踪控制器
// @param ipfsClusterControl *ipfscluster.Control IPFS集群控制器
// @param redisControl *redis.Control Redis控制器
// @param kvDatabaseControl *kvdatabase.Control KvDatabase控制器
// @param webRTCControl *webrtc.Control WebRTC控制器
// @param aiControl *ai.Control AI控制器
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
	authzControl *authz.Control,
	ipfsControl *ipfs.Control,
	mfaControl *mfa.Control,
	tracerControl *tracer.Control,
	ipfsClusterControl *ipfscluster.Control,
	redisControl *redis.Control,
	kvDatabaseControl *kvdatabase.Control,
	webRTCControl *webrtc.Control,
	aiControl *ai.Control,
) *Control {
	serverLogger := loggerControl.GenLogger(logger.ModuleServer)
	serverLogger.Infof("[control] starting new server control...")
	return &Control{
		logger:             serverLogger,
		ctx:                context.Background(),
		dockerControl:      dockerControl,
		configControl:      configControl,
		libp2pControl:      libp2pControl,
		grpcControl:        grpcControl,
		httpControl:        httpControl,
		datasourceControl:  datasourceControl,
		loggerControl:      loggerControl,
		jobControl:         jobControl,
		websocketControl:   websocketControl,
		flagsControl:       flagsControl,
		captchaControl:     captchaControl,
		antsControl:        antsControl,
		authzControl:       authzControl,
		ipfsControl:        ipfsControl,
		mfaControl:         mfaControl,
		tracerControl:      tracerControl,
		ipfsClusterControl: ipfsClusterControl,
		redisControl:       redisControl,
		kvDatabaseControl:  kvDatabaseControl,
		webRTCControl:      webRTCControl,
		aiControl:          aiControl,
	}
}

// StartUp 启动所有服务器组件
// @description 按照依赖顺序启动所有已初始化的服务器组件
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

		if c.tracerControl != nil {
			c.logger.Debugf("[control] starting up tracer server...")
			c.tracerControl.StartUp(failedFunc)
		}

		if c.datasourceControl != nil {
			c.logger.Debugf("[control] starting up datasource server...")
			c.datasourceControl.StartUp(failedFunc)
		}

		if c.redisControl != nil {
			c.logger.Debugf("[control] starting up redis server...")
			c.redisControl.StartUp(failedFunc)
		}

		if c.kvDatabaseControl != nil {
			c.logger.Debugf("[control] starting up kvdatabase server...")
			c.kvDatabaseControl.StartUp(failedFunc)
		}

		if c.grpcControl != nil {
			c.logger.Debugf("[control] starting up grpc server...")
			c.grpcControl.StartUp(failedFunc)
		}

		if c.libp2pControl != nil {
			c.logger.Debugf("[control] starting up libp2p server...")
			c.libp2pControl.StartUp(failedFunc)
		}
		if c.ipfsControl != nil {
			c.logger.Debugf("[control] starting up ipfs server...")
			c.ipfsControl.StartUp(failedFunc)
		}

		if c.ipfsClusterControl != nil {
			c.logger.Debugf("[control] starting up ipfs cluster server...")
			c.ipfsClusterControl.StartUp(failedFunc)
		}

		if c.dockerControl != nil {
			c.logger.Debugf("[control] starting up docker server...")
			c.dockerControl.StartUp(failedFunc)
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

		if c.authzControl != nil {
			c.logger.Debugf("[control] starting up authz server...")
			c.authzControl.StartUp(failedFunc)
		}
		if c.mfaControl != nil {
			c.logger.Debugf("[control] starting up mfa server...")
			c.mfaControl.StartUp(failedFunc)
		}

		if c.websocketControl != nil {
			c.logger.Debugf("[control] starting up websocket server...")
			c.websocketControl.StartUp(failedFunc)
		}

		if c.webRTCControl != nil {
			c.logger.Debugf("[control] starting up webrtc server...")
			c.webRTCControl.StartUp(failedFunc)
		}

		if c.aiControl != nil {
			c.logger.Debugf("[control] starting up ai server...")
			c.aiControl.StartUp(failedFunc)
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
// @description 按照依赖顺序关闭所有已初始化的服务器组件
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	var errs []error

	if c.jobControl != nil {
		c.logger.Debugf("[control] shutting down job server...")
		if err := c.jobControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown job server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.antsControl != nil {
		c.logger.Debugf("[control] shutting down ants pool server...")
		if err := c.antsControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown ants pool server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.websocketControl != nil {
		c.logger.Debugf("[control] shutting down websocket server...")
		if err := c.websocketControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown websocket server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.captchaControl != nil {
		c.logger.Debugf("[control] shutting down captcha server...")
		if err := c.captchaControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown captcha server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.authzControl != nil {
		c.logger.Debugf("[control] shutting down authz server...")
		if err := c.authzControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown authz server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.mfaControl != nil {
		c.logger.Debugf("[control] shutting down mfa server...")
		if err := c.mfaControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown mfa server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.httpControl != nil {
		c.logger.Debugf("[control] shutting down http server...")
		if err := c.httpControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown http server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.libp2pControl != nil {
		c.logger.Debugf("[control] shutting down libp2p server...")
		if err := c.libp2pControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown libp2p server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.grpcControl != nil {
		c.logger.Debugf("[control] shutting down grpc server...")
		if err := c.grpcControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown grpc server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.ipfsControl != nil {
		c.logger.Debugf("[control] shutting down ipfs server...")
		if err := c.ipfsControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown ipfs server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.ipfsClusterControl != nil {
		c.logger.Debugf("[control] shutting down ipfs cluster server...")
		if err := c.ipfsClusterControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown ipfs cluster server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.dockerControl != nil {
		c.logger.Debugf("[control] shutting down docker server...")
		if err := c.dockerControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown docker server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.datasourceControl != nil {
		c.logger.Debugf("[control] shutting down datasource server...")
		if err := c.datasourceControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown datasource server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.redisControl != nil {
		c.logger.Debugf("[control] shutting down redis server...")
		if err := c.redisControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown redis server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.kvDatabaseControl != nil {
		c.logger.Debugf("[control] shutting down kvdatabase server...")
		if err := c.kvDatabaseControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown kvdatabase server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.webRTCControl != nil {
		c.logger.Debugf("[control] shutting down webrtc server...")
		if err := c.webRTCControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown webrtc server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.aiControl != nil {
		c.logger.Debugf("[control] shutting down ai server...")
		if err := c.aiControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown ai server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.tracerControl != nil {
		c.logger.Debugf("[control] shutting down tracer server...")
		if err := c.tracerControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown tracer server err: %v", err)
			errs = append(errs, err)
		}
	}

	// 不会有问题的关闭
	if c.loggerControl != nil {
		c.logger.Debugf("[control] shutting down logger server...")
		if err := c.loggerControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown logger server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.configControl != nil {
		c.logger.Debugf("[control] shutting down config server...")
		if err := c.configControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown config server err: %v", err)
			errs = append(errs, err)
		}
	}

	if c.flagsControl != nil {
		c.logger.Debugf("[control] shutting down flags server...")
		if err := c.flagsControl.Shutdown(); err != nil {
			c.logger.Errorf("[control] shutdown flags server err: %v", err)
			errs = append(errs, err)
		}
	}

	// 如果有错误，返回所有错误
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
