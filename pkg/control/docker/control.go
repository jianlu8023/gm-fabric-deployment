package docker

import (
	"bufio"
	"context"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"io"
	"sync"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"go.uber.org/zap"
)

// https://docs.docker.com/engine/security/protect-access/

// Control Docker控制结构体
type Control struct {
	client *client.Client
	config *config.DockerConfig
	logger *zap.SugaredLogger
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
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
// @description 初始化Docker客户端连接
// @return error 初始化过程中的错误
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
				dc.config.TlsCAFile,
				dc.config.TlsCertFile,
				dc.config.TlsKeyFile,
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
// @description 启动Docker服务，测试连接是否正常
// @param failedFunc func(err error) 启动失败回调函数
func (dc *Control) StartUp(failedFunc func(err error)) {
	dc.once.Do(func() {
		dc.logger.Debugf("[control] docker service is already running...")
		// 测试连接
		version, err := dc.client.ServerVersion(dc.ctx)
		if err != nil {
			dc.logger.Errorf("[control] failed to connect to docker daemon: %v", err)
			failedFunc(err)
			return
		}

		dc.logger.Infof("[control] connected to docker daemon, version: %s", version.Version)
	})
}

// Shutdown 关闭Docker客户端连接
// @description 关闭Docker客户端连接并释放资源
// @return error 关闭过程中的错误
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

// ListNetworks 列出Docker网络
//
// @param networkListOpts func(args *[]filters.KeyValuePair) 选项
//
// @return []network.Summary 网络资源列表
// @return error 错误信息
func (dc *Control) ListNetworks(networkListOpts ...func(args *[]filters.KeyValuePair)) ([]network.Summary, error) {
	dc.logger.Debugf("[control] listing networks...")
	if dc.client == nil {
		return nil, ErrNoAliveDockerClient
	}

	// 列出网络选项
	ftArr := make([]filters.KeyValuePair, 0, 5)
	for _, opt := range networkListOpts {
		opt(&ftArr)
	}
	opts := network.ListOptions{
		Filters: filters.NewArgs(ftArr...),
	}

	networkList, err := dc.client.NetworkList(dc.ctx, opts)
	if err != nil {
		dc.logger.Errorf("[control] failed to list networks: %v", err)
		return nil, err
	}
	return networkList, nil
}

// GetNetwork 获取Docker网络
//
// @param networkName 网络名称
// @param networkId 网络ID
//
// @return network.Summary 网络资源
// @return error 错误信息
func (dc *Control) GetNetwork(networkListOpts ...func(args *[]filters.KeyValuePair)) (network.Summary, error) {
	dc.logger.Debugf("[control] getting network...")
	if dc.client == nil {
		return network.Summary{}, ErrNoAliveDockerClient
	}

	listNetworks, err := dc.ListNetworks(
		networkListOpts...,
	)
	if err != nil {
		dc.logger.Errorf("[control] get network failed: %v", err)
		return network.Summary{}, err
	}
	return listNetworks[0], nil
}

// CreateNetwork 创建网络
//
// @param networkName 网络名称
// @param driver 网络驱动
// @param opts func(create *network.CreateOptions) 网络创建选项
//
// @return network.Summary 网络资源
// @return error 错误信息
func (dc *Control) CreateNetwork(networkName string, driver string, opts ...func(create *network.CreateOptions)) (network.Summary, error) {
	dc.logger.Debugf("[control] creating net: %s, driver: %s", networkName, driver)
	if dc.client == nil {
		return network.Summary{}, ErrNoAliveDockerClient
	}

	dc.logger.Debugf("[control] check net %s exists...", networkName)
	if net, err := dc.GetNetwork(WithNetworkName(networkName)); err == nil {
		// net exists
		dc.logger.Debugf("[control] net %s exists", networkName)
		return net, err
	} else {
		dc.logger.Debugf("[control] net %s need creating...", networkName)
	}

	ipv6Default := false
	networkOpts := network.CreateOptions{
		// 驱动类型
		Driver: driver,
		// 默认不启用ipv6
		EnableIPv6: &ipv6Default,
		// 默认创建桥接网络
		// CheckDuplicate: true,
	}
	for _, opt := range opts {
		opt(&networkOpts)
	}

	resp, err := dc.client.NetworkCreate(dc.ctx, networkName, networkOpts)
	if err != nil {
		dc.logger.Errorf("[control] failed to create net: %v", err)
		return network.Summary{}, err
	}
	dc.logger.Infof("[control] net %v created sucesfully, ID %s", networkName, resp.ID)

	return dc.GetNetwork(WithNetworkName(networkName), WithNetworkID(resp.ID))
}

// RemoveNetwork 删除网络
//
// @param networkId 网络ID
//
// @returns error 错误信息
func (dc *Control) RemoveNetwork(networkId string) error {
	dc.logger.Debugf("[control] removing network: %s", networkId)

	if dc.client == nil {
		return ErrNoAliveDockerClient
	}

	if err := dc.client.NetworkRemove(dc.ctx, networkId); err != nil {
		dc.logger.Errorf("[control] failed to remove network: %v", err)
		return err
	}
	dc.logger.Infof("[control] network %s removed sucesfully", networkId)
	return nil
}

// PullImage 拉取Docker镜像
//
// @param imageName 镜像名称
//
// @returns error 错误信息
func (dc *Control) PullImage(imageName string, pullImageOpts ...func(options *image.PullOptions)) error {
	dc.logger.Infof("[control] pulling image: %s", imageName)
	if dc.client == nil {
		return ErrNoAliveDockerClient
	}

	// 创建拉取选项
	options := image.PullOptions{
		All: true,
	}
	for _, opt := range pullImageOpts {
		opt(&options)
	}

	// 获取拉取镜像的输出流
	resp, err := dc.client.ImagePull(context.Background(), imageName, options)
	if err != nil {
		dc.logger.Errorf("[control] failed to pull image %s: %v", imageName, err)
		return err
	}
	defer func(resp io.ReadCloser) {
		if err := resp.Close(); err != nil {
			dc.logger.Errorf("[control] failed to close response: %v", err)
		}
	}(resp)

	// 使用bufio.Scanner逐行处理响应
	scanner := bufio.NewScanner(resp)
	for scanner.Scan() {
		line := scanner.Text()
		var jm jsonmessage.JSONMessage // 使用docker的JSONMessage结构
		if err := json.Unmarshal([]byte(line), &jm); err != nil {
			dc.logger.Errorf("[control] failed to unmarshal json: %v, line: %s", err, line)
			continue
		}

		// 检查是否有错误
		if jm.Error != nil {
			dc.logger.Errorf("[control] failed to pull image %s: %v", imageName, jm.Error)
			return jm.Error
		}

		// 输出状态信息
		if jm.ID != "" && jm.Status != "" {
			dc.logger.Infof("[control] pulling %s: %s: %s", imageName, jm.ID, jm.Status)
		} else if jm.Status != "" {
			dc.logger.Infof("[control] pulling %s: %s", imageName, jm.Status)
		}

		// 输出进度详情（调试用）
		if jm.Progress != nil {
			dc.logger.Debugf("[control] pulling %s: progress: %s", imageName, jm.Progress.String())
		}
	}

	// 检查scanner是否有错误
	if err := scanner.Err(); err != nil {
		dc.logger.Errorf("[control] error reading response: %v", err)
		return err
	}

	dc.logger.Infof("[control] image %s pulled successfully", imageName)
	return nil
}

// RemoveImage 删除Docker镜像
// @description 删除指定的Docker镜像
// @param imageId string 镜像ID
// @param removeImageOpts ...func(options *image.RemoveOptions) 镜像删除选项
// @return []image.DeleteResponse 镜像删除响应项列表
// @return error 删除过程中的错误
func (dc *Control) RemoveImage(imageId string, removeImageOpts ...func(options *image.RemoveOptions)) ([]image.DeleteResponse, error) {
	dc.logger.Debugf("[control] removing image: %s", imageId)
	if dc.client == nil {
		return nil, ErrNoAliveDockerClient
	}

	// 删除镜像
	opts := image.RemoveOptions{}

	for _, opt := range removeImageOpts {
		opt(&opts)
	}

	resp, err := dc.client.ImageRemove(dc.ctx, imageId, opts)
	if err != nil {
		dc.logger.Errorf("[control] failed to remove image: %s", err)
		return nil, err
	}
	return resp, nil
}

func (dc *Control) GetImage(imageListOpts ...func(args *[]filters.KeyValuePair)) (image.Summary, error) {
	dc.logger.Debugf("[control] getting image...")
	images, err := dc.ListImages(imageListOpts...)
	if err != nil {
		dc.logger.Errorf("[control] getting image info failed: %v", err)
		return image.Summary{}, err
	}
	return images[0], nil
}

// ListImages 列出Docker镜像
// @description 列出当前系统中的Docker镜像
// @param imageListOpts ...func(args *[]filters.KeyValuePair) 镜像列表选项
// @return []image.Summary 镜像摘要列表
// @return error 列出过程中的错误
func (dc *Control) ListImages(imageListOpts ...func(args *[]filters.KeyValuePair)) ([]image.Summary, error) {
	dc.logger.Debugf("[control] listing images...")
	if dc.client == nil {
		return nil, ErrNoAliveDockerClient
	}

	// 列出网络选项
	ftArr := make([]filters.KeyValuePair, 0, 5)
	for _, opt := range imageListOpts {
		opt(&ftArr)
	}

	opts := image.ListOptions{
		All:     true,
		Filters: filters.NewArgs(ftArr...),
	}
	imageList, err := dc.client.ImageList(dc.ctx, opts)
	if err != nil {
		dc.logger.Errorf("[control] failed to list images: %v", err)
		return nil, err
	}
	return imageList, nil
}

// CreateContainer 创建Docker容器
// @description 创建新的Docker容器
// @param containerName string 容器名称
// @param imageName string 镜像名称
// @param config *container.Config 容器配置
// @param hostConfig *container.HostConfig 主机配置
// @return string 容器ID
// @return error 创建过程中的错误
func (dc *Control) CreateContainer(containerName string, imageName string, config *container.Config, hostConfig *container.HostConfig) (string, error) {
	dc.logger.Infof("[control] creating container: %s with image: %s", containerName, imageName)
	if dc.client == nil {
		return "", ErrNoAliveDockerClient
	}

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
// @description 启动指定的Docker容器
// @param containerID string 容器ID
// @return error 启动过程中的错误
func (dc *Control) StartContainer(containerID string) error {
	dc.logger.Infof("[control] starting container: %s", containerID)
	if dc.client == nil {
		return ErrNoAliveDockerClient
	}

	// 启动容器
	if err := dc.client.ContainerStart(dc.ctx, containerID, container.StartOptions{}); err != nil {
		dc.logger.Errorf("[control] failed to start container %s: %v", containerID, err)
		return err
	}

	dc.logger.Infof("[control] container %s started successfully", containerID)
	return nil
}

// StopContainer 停止Docker容器
// @description 停止指定的Docker容器
// @param containerID string 容器ID
// @param timeout int 停止超时时间
// @return error 停止过程中的错误
func (dc *Control) StopContainer(containerID string, timeout int) error {
	dc.logger.Infof("[control] stopping container: %s", containerID)

	// 停止容器容器选项
	opts := container.StopOptions{
		Timeout: &timeout,
	}

	// 停止容器
	if err := dc.client.ContainerStop(dc.ctx, containerID, opts); err != nil {
		dc.logger.Errorf("[control] failed to stop container %s: %v", containerID, err)
		return err
	}

	dc.logger.Infof("[control] container %s stopped successfully", containerID)
	return nil
}

// RemoveContainer 删除Docker容器
// @description 删除指定的Docker容器
// @param containerID string 容器ID
// @param force bool 是否强制删除
// @return error 删除过程中的错误
func (dc *Control) RemoveContainer(containerID string, force bool) error {
	dc.logger.Infof("[control] removing container: %s", containerID)

	if dc.client == nil {
		return ErrNoAliveDockerClient
	}

	// 删除容器选项
	opts := container.RemoveOptions{
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

// ListContainers 列出Docker容器
// @description 列出当前系统中的Docker容器
// @param all bool 是否列出所有容器（包括已停止的）
// @return []container.Summary 容器列表
// @return error 列出过程中的错误
func (dc *Control) ListContainers(all bool) ([]container.Summary, error) {
	dc.logger.Infof("[control] listing containers, all: %v", all)
	if dc.client == nil {
		return nil, ErrNoAliveDockerClient
	}

	// 列出容器选项
	opts := container.ListOptions{
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
// @description 获取指定Docker容器的运行状态
// @param containerID string 容器ID
// @return string 容器状态
// @return error 获取过程中的错误
func (dc *Control) GetContainerStatus(containerID string) (string, error) {
	dc.logger.Infof("[control] getting status for container: %s", containerID)

	if dc.client == nil {
		return "", ErrNoAliveDockerClient
	}

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
// @description 在指定的Docker容器中执行命令
// @param containerID string 容器ID
// @param cmd []string 要执行的命令
// @return string 命令执行结果
// @return error 执行过程中的错误
func (dc *Control) ExecuteCommand(containerID string, cmd []string) (string, error) {
	dc.logger.Infof("[control] executing command in container %s: %v", containerID, cmd)
	if dc.client == nil {
		return "", ErrNoAliveDockerClient
	}
	// 执行命令选项
	execConfig := container.ExecOptions{
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

	// Attach到执行
	execAttachOpts := container.ExecAttachOptions{}

	// 开始执行
	execResp, err := dc.client.ContainerExecAttach(dc.ctx, resp.ID, execAttachOpts)
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
