package docker

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"go.uber.org/zap"
)

// Control Docker控制结构体
type Control struct {
	client *client.Client
	config *config.DockerConfig
	logger *zap.SugaredLogger
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDockerControl 创建一个新的Docker控制器
// @param dockerConfig *config.DockerConfig docker配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control docker控制器
// @return error 新建过程中的错误
func NewDockerControl(dockerConfig *config.DockerConfig, loggerControl *logger.Control) (*Control, error) {
	dockerLogger := loggerControl.GenLogger(logger.ModuleDocker)
	dockerLogger.Infof("[control] starting new docker control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化DockerControl
	dc := &Control{
		logger: dockerLogger,
		ctx:    ctx,
		cancel: cancel,
		config: dockerConfig,
	}

	// 初始化Docker客户端
	if err := dc.initClient(); err != nil {
		dockerLogger.Errorf("[control] docker initialization client failed: %v", err)
		cancel()
		return nil, err
	}

	dockerLogger.Infof("[control] docker control started successfully")
	return dc, nil
}

// initClient 初始化Docker客户端
func (dc *Control) initClient() error {
	dc.logger.Debugf("[control] initializing docker client...")

	// 设置Docker客户端选项
	opts := []client.Opt{
		client.WithAPIVersionNegotiation(),
	}

	// 如果指定了API版本
	if dc.config.APIVersion != "" {
		opts = append(opts, client.WithVersion(dc.config.APIVersion))
	}

	// 配置TLS连接
	if dc.config.TlsEnabled {
		dc.logger.Debugf("[control] configuring docker client with TLS...")
		opts = append(opts,
			client.WithTLSClientConfig(
				dc.config.TlsCertFile,
				dc.config.TlsKeyFile,
				dc.config.TlsCAFile,
			),
		)
	}

	// 设置Docker主机
	if dc.config.Host != "" {
		opts = append(opts, client.WithHost(dc.config.Host))
	}

	// 创建Docker客户端
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		dc.logger.Errorf("[control] failed to create docker client: %v", err)
		return err
	}

	dc.client = cli
	return nil
}

// StartUp 启动Docker服务（如果需要）
func (dc *Control) StartUp(failedFunc func(err error)) {
	// dc.logger.Infof("[control] docker service is already running")
	// 测试连接
	version, err := dc.client.ServerVersion(dc.ctx)
	if err != nil {
		dc.logger.Errorf("[control] failed to connect to docker daemon: %v", err)
		failedFunc(err)
		return
	}

	dc.logger.Infof("[control] connected to docker daemon, version: %s", version.Version)
}

// Shutdown 关闭Docker客户端连接
func (dc *Control) Shutdown() error {
	dc.logger.Infof("[control] shutting down docker control...")
	dc.cancel()
	if dc.client != nil {
		if err := dc.client.Close(); err != nil {
			dc.logger.Errorf("[control] failed to close docker client: %v", err)
			return err
		}
	}
	dc.logger.Infof("[control] docker control shutdown successfully")
	return nil
}

// PullImage 拉取Docker镜像
func (dc *Control) PullImage(imageName string) error {
	dc.logger.Infof("[control] pulling image: %s", imageName)

	// 创建拉取选项
	options := types.ImagePullOptions{
		All: true,
	}

	// 获取拉取镜像的输出流
	resp, err := dc.client.ImagePull(context.Background(), imageName, options)
	if err != nil {
		dc.logger.Errorf("[control] failed to pull image %s: %v", imageName, err)
		return err
	}
	defer resp.Close()

	// scanner := bufio.NewScanner(resp)
	// for scanner.Scan() {
	// 	line := scanner.Text()
	// 	var jm map[string]interface{}
	// 	if err := json.Unmarshal([]byte(line), &jm); err != nil {
	// 		dc.logger.Errorf("[control] failed to unmarshal json: %v", err)
	// 		continue
	// 	}
	// 	if errMsg, ok := jm["error"]; ok {
	// 		dc.logger.Errorf("[control] failed to pull image %s: %v", imageName, errMsg)
	// 		return errors.New(errMsg.(string))
	// 	}
	// 	status, hasStatus := jm["status"]
	// 	id, hasID := jm["id"]
	// 	progressDetail, hasProgressDetail := jm["progressDetail"]
	//
	// 	if hasID && hasStatus {
	// 		dc.logger.Infof("[control] pulling %s: %s: %v", imageName, id, status)
	// 	} else if hasStatus {
	// 		dc.logger.Infof("[control] pulling %s: %v", imageName, status)
	// 	}
	//
	// 	if hasProgressDetail {
	// 		dc.logger.Debugf("[control] pulling %s: progress detail: %v", imageName, progressDetail)
	// 	}
	// }
	// if err := scanner.Err(); err != nil {
	// 	dc.logger.Errorf("[control] error reading response: %v", err)
	// 	return err
	// }

	// 解析JSON响应
	decoder := json.NewDecoder(resp)
	// decoder := jsonmessage.NewDecoder(resp)
	for {
		var jm jsonmessage.JSONMessage
		if err := decoder.Decode(&jm); err != nil {
			if err == io.EOF {
				break
			}
			dc.logger.Errorf("[control] error decoding pull response: %v", err)
			return err
		}

		// if jm.Error != nil {
		// 	dc.logger.Errorf("[control] error pulling image: %v", jm.Error)
		// 	return jm.Error
		// }
		jm.Display(os.Stdout, false)
		// 输出进度信息
		// if jm.Progress != nil {
		// 	dc.logger.Debugf("[control] pulling %s: %s", imageName, jm.Progress.String())
		// }
	}

	dc.logger.Infof("[control] image %s pulled successfully", imageName)
	return nil
}

// CreateContainer 创建Docker容器
func (dc *Control) CreateContainer(containerName string, imageName string, config *container.Config, hostConfig *container.HostConfig) (string, error) {
	dc.logger.Infof("[control] creating container: %s with image: %s", containerName, imageName)

	// 创建容器选项
	// createOptions := types.ContainerCreateConfig{
	// 	Name: containerName,
	// }

	// 创建容器
	resp, err := dc.client.ContainerCreate(dc.ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		dc.logger.Errorf("[control] failed to create container %s: %v", containerName, err)
		return "", err
	}

	dc.logger.Infof("[control] container %s created successfully, ID: %s", containerName, resp.ID)
	return resp.ID, nil
	// return "nil", nil
}

// StartContainer 启动Docker容器
func (dc *Control) StartContainer(containerID string) error {
	dc.logger.Infof("[control] starting container: %s", containerID)

	// 启动容器
	if err := dc.client.ContainerStart(dc.ctx, containerID, types.ContainerStartOptions{}); err != nil {
		dc.logger.Errorf("[control] failed to start container %s: %v", containerID, err)
		return err
	}

	dc.logger.Infof("[control] container %s started successfully", containerID)
	return nil
}

// StopContainer 停止Docker容器
func (dc *Control) StopContainer(containerID string, timeout *time.Duration) error {
	dc.logger.Infof("[control] stopping container: %s", containerID)

	// 停止容器选项
	// opts := types.StopOptions{}
	// if timeout != nil {
	// 	opts.Timeout = timeout
	// }

	// 停止容器
	// if err := dc.client.ContainerStop(dc.ctx, containerID, opts); err != nil {
	// 	dc.logger.Errorf("[control] failed to stop container %s: %v", containerID, err)
	// 	return err
	// }

	dc.logger.Infof("[control] container %s stopped successfully", containerID)
	return nil
}

// RemoveContainer 删除Docker容器
func (dc *Control) RemoveContainer(containerID string, force bool) error {
	dc.logger.Infof("[control] removing container: %s", containerID)

	// 删除容器选项
	opts := types.ContainerRemoveOptions{
		Force: force,
	}

	// 删除容器
	if err := dc.client.ContainerRemove(dc.ctx, containerID, opts); err != nil {
		dc.logger.Errorf("[control] failed to remove container %s: %v", containerID, err)
		return err
	}

	dc.logger.Infof("[control] container %s removed successfully", containerID)
	return nil
}

func (dc *Control) ListNetworks(all bool) ([]types.NetworkResource, error) {
	dc.logger.Debugf("[control] listing networks,all: %v", all)
	// 列出网络选项
	opts := types.NetworkListOptions{
		Filters: filters.NewArgs(),
	}
	networkList, err := dc.client.NetworkList(dc.ctx, opts)
	if err != nil {
		dc.logger.Errorf("[control] failed to list networks: %v", err)
		return nil, err
	}
	return networkList, nil
}

// ListContainers 列出Docker容器
func (dc *Control) ListContainers(all bool) ([]types.Container, error) {
	dc.logger.Infof("[control] listing containers, all: %v", all)

	// 列出容器选项
	opts := types.ContainerListOptions{
		All: all,
	}

	// 列出容器
	containers, err := dc.client.ContainerList(dc.ctx, opts)
	if err != nil {
		dc.logger.Errorf("[control] failed to list containers: %v", err)
		return nil, err
	}

	dc.logger.Infof("[control] listed %d containers successfully", len(containers))
	return containers, nil
	// return nil, nil
}

// GetContainerStatus 获取Docker容器状态
func (dc *Control) GetContainerStatus(containerID string) (string, error) {
	dc.logger.Infof("[control] getting status for container: %s", containerID)

	// 查看容器详情
	inspect, err := dc.client.ContainerInspect(dc.ctx, containerID)
	if err != nil {
		dc.logger.Errorf("[control] failed to inspect container %s: %v", containerID, err)
		return "", err
	}

	status := inspect.State.Status
	dc.logger.Infof("[control] container %s status: %s", containerID, status)
	return status, nil
}

// ExecuteCommand 在容器中执行命令
func (dc *Control) ExecuteCommand(containerID string, cmd []string) (string, error) {
	dc.logger.Infof("[control] executing command in container %s: %v", containerID, cmd)

	// 执行命令选项
	execConfig := types.ExecConfig{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	// 创建执行
	resp, err := dc.client.ContainerExecCreate(dc.ctx, containerID, execConfig)
	if err != nil {
		dc.logger.Errorf("[control] failed to create exec in container %s: %v", containerID, err)
		return "", err
	}

	// 开始执行
	execResp, err := dc.client.ContainerExecAttach(dc.ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		dc.logger.Errorf("[control] failed to start exec in container %s: %v", containerID, err)
		return "", err
	}
	defer execResp.Close()

	// 读取执行结果
	output, err := io.ReadAll(execResp.Reader)
	if err != nil {
		dc.logger.Errorf("[control] failed to read exec output: %v", err)
		return "", err
	}

	dc.logger.Infof("[control] command executed successfully in container %s", containerID)
	return string(output), nil
}

func (dc *Control) ImageList() ([]types.ImageSummary, error) {
	opts := types.ImageListOptions{
		All: true,
	}
	imageList, err := dc.client.ImageList(dc.ctx, opts)
	if err != nil {
		dc.logger.Errorf("[control] failed to list images: %v", err)
		return nil, err
	}
	return imageList, nil
}
