package docker

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/json"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

// https://docs.docker.com/engine/security/protect-access/

// Control Docker控制结构体
//
// @description 封装 Docker API 客户端的生命周期管理与常用操作，提供网络、镜像、容器、命令执行等高层接口
// @struct
type Control struct {
	mu           sync.RWMutex
	client       *client.Client
	config       *config.DockerConfig
	logger       *zap.SugaredLogger
	ctx          context.Context
	cancel       context.CancelFunc
	startOnce    sync.Once
	shutdownOnce sync.Once
	closed       bool
}

// NewDockerControl 创建一个新的Docker控制器
//
// @description 根据配置创建 Docker 控制器；即使未启用（Enabled=false）也返回非 nil 控制器，
//
//	其所有方法会统一返回 ErrNoAliveDockerClient，避免调用方判空遗漏导致 panic。
//
// @param dockerConfig *config.DockerConfig docker配置，为 nil 时使用默认配置
// @param loggerControl *logger.Control 日志控制器，为 nil 时使用默认 debug 级别日志
// @return *Control docker控制器，始终非 nil
func NewDockerControl(dockerConfig *config.DockerConfig, loggerControl *logger.Control) *Control {
	dockerConfig = check.IF[*config.DockerConfig](dockerConfig == nil,
		getDefaultConfig(),
		dockerConfig,
	)
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	// dockerLogger := loggerControl.GenLogger(logger.ModuleDocker)
	dockerLogger := loggerControl.GenLogger("")
	dockerLogger.Infof("[docker/control] starting new docker control, enabled=%v...", dockerConfig.Enabled)

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化DockerControl（始终返回非 nil 实例）
	dc := &Control{
		logger: dockerLogger,
		ctx:    ctx,
		cancel: cancel,
		config: dockerConfig,
	}

	dockerLogger.Infof("[docker/control] docker control started successfully")
	return dc
}

// timeoutCtx 根据 DefaultTimeout 派生一个带超时的子上下文
//
// @description 为每次 Docker API 调用提供超时保护，避免 daemon 无响应时无限阻塞
// @return context.Context 带超时的上下文
// @return context.CancelFunc 取消函数，调用方必须 defer 调用
func (dc *Control) timeoutCtx() (context.Context, context.CancelFunc) {
	timeout := time.Duration(dc.config.DefaultTimeout) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return context.WithTimeout(dc.ctx, timeout)
}

// ensureClient 检查 Docker 客户端是否可用
//
// @description 统一的客户端存活检查，调用前加读锁保证并发安全
// @return error 若客户端未初始化或已关闭则返回 ErrNoAliveDockerClient
func (dc *Control) ensureClient() error {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	if dc.client == nil || dc.closed {
		return ErrNoAliveDockerClient
	}
	return nil
}

// initClient 初始化Docker客户端
//
// @description 初始化Docker客户端连接，包含 TLS 配置校验与多证书数量一致性检查
// @return error 初始化过程中的错误
func (dc *Control) initClient() error {
	dc.logger.Debugf("[docker/control] initializing docker client...")

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
		dc.logger.Debugf("[docker/control] configuring docker client with TLS...")
		if len(dc.config.TlsCertFile) == 0 || len(dc.config.TlsKeyFile) == 0 {
			return fmt.Errorf("docker TLS启用时至少需要一对证书和密钥文件")
		}
		if len(dc.config.TlsCertFile) != len(dc.config.TlsKeyFile) {
			return fmt.Errorf("docker TLS证书与密钥数量不匹配: cert=%d, key=%d",
				len(dc.config.TlsCertFile), len(dc.config.TlsKeyFile))
		}
		if dc.config.TlsCAFile == "" {
			dc.logger.Warnf("[docker/control] TLS 启用但未配置 CA 文件，客户端将不校验服务端证书，存在安全风险")
		}
		// 目前 Docker Go SDK 的 WithTLSClientConfig 仅支持单对 cert/key，取第一对；
		// 若后续 SDK 支持多证书，可在此处扩展为遍历加载。
		opts = append(opts,
			client.WithTLSClientConfig(
				dc.config.TlsCAFile,
				dc.config.TlsCertFile[0],
				dc.config.TlsKeyFile[0],
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
		dc.logger.Errorf("[docker/control] failed to create docker client: %v", err)
		return err
	}

	dc.mu.Lock()
	dc.client = cli
	dc.mu.Unlock()
	return nil
}

// StartUp 启动Docker服务（如果需要）
//
// @description 启动Docker服务，初始化客户端并测试与 daemon 的连接；
//
//	若初始化或连接失败，会调用 failedFunc 并统一 cancel 上下文、清空客户端以便重试。
//
// @param failedFunc func(err error) 启动失败回调函数，可为 nil
func (dc *Control) StartUp(failedFunc func(err error)) {
	dc.startOnce.Do(func() {
		if !dc.config.Enabled {
			dc.logger.Debugf("[docker/control] docker control is disabled, skip startup")
			return
		}
		// 初始化Docker客户端
		if err := dc.initClient(); err != nil {
			dc.logger.Errorf("[docker/control] docker initialization client failed: %v", err)
			dc.cancel()
			dc.mu.Lock()
			dc.client = nil
			dc.mu.Unlock()
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}

		dc.logger.Debugf("[docker/control] docker service is already running...")
		// 测试连接（带超时）
		ctx, cancel := dc.timeoutCtx()
		defer cancel()
		version, err := dc.client.ServerVersion(ctx)
		if err != nil {
			dc.logger.Errorf("[docker/control] failed to connect to docker daemon: %v", err)
			dc.cancel()
			dc.mu.Lock()
			if dc.client != nil {
				_ = dc.client.Close()
				dc.client = nil
			}
			dc.mu.Unlock()
			if failedFunc != nil {
				failedFunc(err)
			}
			return
		}

		dc.logger.Infof("[docker/control] connected to docker daemon, version: %s", version.Version)
	})
}

// Shutdown 关闭Docker客户端连接
//
// @description 关闭Docker客户端连接并释放资源；
//
//	通过 shutdownOnce 保证幂等，重复调用不会重复 Close 也不会返回错误。
//
// @return error 关闭过程中的错误（首次关闭失败时返回）
func (dc *Control) Shutdown() error {
	var shutdownErr error
	dc.shutdownOnce.Do(func() {
		dc.logger.Infof("[docker/control] shutting down docker control...")
		dc.cancel()

		dc.mu.Lock()
		defer dc.mu.Unlock()
		if dc.client != nil {
			if err := dc.client.Close(); err != nil {
				dc.logger.Errorf("[docker/control] failed to close docker client: %v", err)
				shutdownErr = err
			}
			dc.client = nil
		}
		dc.closed = true
	})
	if shutdownErr != nil {
		return shutdownErr
	}
	if !dc.closed {
		// 非首次调用，直接视为成功
		return nil
	}
	dc.logger.Infof("[docker/control] docker control shutdown successfully")
	return nil
}

// ListNetworks 列出Docker网络
//
// @description 列出当前系统中满足过滤器条件的 Docker 网络
// @param networkListFilters ...func(args *[]filters.KeyValuePair) 网络列表过滤器选项
// @return []network.Summary 网络资源列表
// @return error 错误信息
func (dc *Control) ListNetworks(networkListFilters ...func(args *[]filters.KeyValuePair)) ([]network.Summary, error) {
	dc.logger.Debugf("[docker/control] listing networks...")
	if err := dc.ensureClient(); err != nil {
		return nil, err
	}

	// 列出网络选项
	ftArr := make([]filters.KeyValuePair, 0, 5)
	for _, opt := range networkListFilters {
		opt(&ftArr)
	}
	opts := network.ListOptions{
		Filters: filters.NewArgs(ftArr...),
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	networkList, err := cli.NetworkList(ctx, opts)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to list networks: %v", err)
		return nil, err
	}
	return networkList, nil
}

// GetNetwork 获取Docker网络
//
// @description 获取满足过滤器条件的第一个 Docker 网络，未找到时返回 ErrNotFound
// @param networkListFilters ...func(args *[]filters.KeyValuePair) 网络查询过滤器选项
// @return network.Summary 网络资源
// @return error 错误信息，未找到时为 ErrNotFound
func (dc *Control) GetNetwork(networkListFilters ...func(args *[]filters.KeyValuePair)) (network.Summary, error) {
	dc.logger.Debugf("[docker/control] getting network...")
	if err := dc.ensureClient(); err != nil {
		return network.Summary{}, err
	}

	listNetworks, err := dc.ListNetworks(
		networkListFilters...,
	)
	if err != nil {
		dc.logger.Errorf("[docker/control] get network failed: %v", err)
		return network.Summary{}, err
	}
	if len(listNetworks) == 0 {
		return network.Summary{}, fmt.Errorf("%w: no network found", ErrNotFound)
	}
	return listNetworks[0], nil
}

// CreateNetwork 创建网络
//
// @description 创建 Docker 网络；若网络已存在（409 Conflict）则直接返回已存在的网络信息，
//
//	避免"先查后建"的竞态问题；其他真实错误（daemon 连接失败等）会向上抛出。
//
// @param networkName string 网络名称
// @param driver string 网络驱动（如 bridge、overlay）
// @param networkCreateOpts ...func(create *network.CreateOptions) 网络创建选项
// @return network.Summary 网络资源
// @return error 错误信息
func (dc *Control) CreateNetwork(networkName string, driver string,
	networkCreateOpts ...func(create *network.CreateOptions),
) (network.Summary, error) {
	dc.logger.Debugf("[docker/control] creating net: %s, driver: %s", networkName, driver)
	if err := dc.ensureClient(); err != nil {
		return network.Summary{}, err
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
	for _, opt := range networkCreateOpts {
		opt(&networkOpts)
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	resp, err := cli.NetworkCreate(ctx, networkName, networkOpts)
	if err != nil {
		// 若网络已冲突（已存在），则查询并返回现有网络
		if errdefs.IsConflict(err) {
			dc.logger.Debugf("[docker/control] net %s already exists (conflict), returning existing one", networkName)
			return dc.GetNetwork(WithNetworkQueryName(networkName))
		}
		dc.logger.Errorf("[docker/control] failed to create net: %v", err)
		return network.Summary{}, err
	}
	dc.logger.Infof("[docker/control] net %v created successfully, ID %s", networkName, resp.ID)

	return dc.GetNetwork(WithNetworkQueryName(networkName), WithNetworkQueryID(resp.ID))
}

// RemoveNetwork 删除网络
//
// @description 删除指定 ID 的 Docker 网络
// @param networkId string 网络ID
// @return error 错误信息
func (dc *Control) RemoveNetwork(networkId string) error {
	dc.logger.Debugf("[docker/control] removing network: %s", networkId)

	if err := dc.ensureClient(); err != nil {
		return err
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	if err := cli.NetworkRemove(ctx, networkId); err != nil {
		dc.logger.Errorf("[docker/control] failed to remove network: %v", err)
		return err
	}
	dc.logger.Infof("[docker/control] network %s removed successfully", networkId)
	return nil
}

// InspectNetwork 查看网络详细信息
//
// @description 调用 NetworkInspect 获取网络的完整详细信息（包含子网、网关、连接的容器等），
//
//	与 GetNetwork 仅返回列表摘要（Summary）不同。
//
// @param networkId string 网络ID或网络名称
// @return network.Inspect 网络详细信息
// @return error 错误信息
func (dc *Control) InspectNetwork(networkId string) (network.Inspect, error) {
	dc.logger.Infof("[docker/control] inspecting network: %s", networkId)
	if err := dc.ensureClient(); err != nil {
		return network.Inspect{}, err
	}

	opts := network.InspectOptions{}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	inspect, err := cli.NetworkInspect(ctx, networkId, opts)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to inspect network: %v", err)
		return network.Inspect{}, err
	}
	return inspect, nil
}

// PullImage 拉取Docker镜像
//
// @description 拉取指定名称的 Docker 镜像，逐行解析进度并输出到日志；
//
//	使用控制器上下文（dc.ctx）派生超时上下文，保证控制器 Shutdown 后能及时取消。
//
// @param imageName string 镜像名称（含 tag）
// @param pullImageOpts ...func(options *image.PullOptions) 镜像拉取选项
// @return error 错误信息
func (dc *Control) PullImage(imageName string, pullImageOpts ...func(options *image.PullOptions)) error {
	dc.logger.Infof("[docker/control] pulling image: %s", imageName)
	if err := dc.ensureClient(); err != nil {
		return err
	}

	// 创建拉取选项
	options := image.PullOptions{
		All: false,
	}
	for _, opt := range pullImageOpts {
		opt(&options)
	}

	// 拉取通常耗时较长，使用控制器上下文，仅受 Shutdown 取消影响；
	// 如需更短超时，可在此派生 WithTimeout。
	ctx := dc.ctx

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 获取拉取镜像的输出流
	resp, err := cli.ImagePull(ctx, imageName, options)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to pull image %s: %v", imageName, err)
		return err
	}
	defer func(resp io.ReadCloser) {
		if closeErr := resp.Close(); closeErr != nil {
			dc.logger.Errorf("[docker/control] failed to close response: %v", closeErr)
		}
	}(resp)

	// 使用bufio.Scanner逐行处理响应
	scanner := bufio.NewScanner(resp)
	// 适当放宽单行上限，应对个别进度消息较长的情况
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		var jm jsonmessage.JSONMessage // 使用docker的JSONMessage结构
		if unmarshalErr := json.Unmarshal([]byte(line), &jm); unmarshalErr != nil {
			dc.logger.Errorf("[docker/control] failed to unmarshal json: %v, line: %s", unmarshalErr, line)
			continue
		}

		// 检查是否有错误
		if jm.Error != nil {
			dc.logger.Errorf("[docker/control] failed to pull image %s: %v", imageName, jm.Error)
			return jm.Error
		}

		// 输出状态信息
		if jm.ID != "" && jm.Status != "" {
			dc.logger.Infof("[docker/control] pulling %s: %s: %s", imageName, jm.ID, jm.Status)
		} else if jm.Status != "" {
			dc.logger.Infof("[docker/control] pulling %s: %s", imageName, jm.Status)
		}

		// 输出进度详情（调试用）
		if jm.Progress != nil {
			dc.logger.Debugf("[docker/control] pulling %s: progress: %s", imageName, jm.Progress.String())
		}
	}

	// 检查scanner是否有错误
	if scanErr := scanner.Err(); scanErr != nil {
		dc.logger.Errorf("[docker/control] error reading response: %v", scanErr)
		return scanErr
	}

	dc.logger.Infof("[docker/control] image %s pulled successfully", imageName)
	return nil
}

// RemoveImage 删除Docker镜像
//
// @description 删除指定的Docker镜像
// @param imageId string 镜像ID或镜像名称
// @param removeImageOpts ...func(options *image.RemoveOptions) 镜像删除选项
// @return []image.DeleteResponse 镜像删除响应项列表
// @return error 删除过程中的错误
func (dc *Control) RemoveImage(imageId string, removeImageOpts ...func(options *image.RemoveOptions)) ([]image.DeleteResponse, error) {
	dc.logger.Debugf("[docker/control] removing image: %s", imageId)
	if err := dc.ensureClient(); err != nil {
		return nil, err
	}

	// 删除镜像
	opts := image.RemoveOptions{}
	for _, opt := range removeImageOpts {
		opt(&opts)
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	resp, err := cli.ImageRemove(ctx, imageId, opts)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to remove image: %v", err)
		return nil, err
	}
	return resp, nil
}

// GetImage 获取Docker镜像
//
// @description 获取满足过滤器条件的第一个 Docker 镜像，未找到时返回 ErrNotFound
// @param imageListFilters ...func(args *[]filters.KeyValuePair) 镜像查询过滤器选项
// @return image.Summary 镜像摘要
// @return error 错误信息，未找到时为 ErrNotFound
func (dc *Control) GetImage(imageListFilters ...func(args *[]filters.KeyValuePair)) (image.Summary, error) {
	dc.logger.Debugf("[docker/control] getting image...")
	images, err := dc.ListImages(imageListFilters...)
	if err != nil {
		dc.logger.Errorf("[docker/control] getting image info failed: %v", err)
		return image.Summary{}, err
	}
	if len(images) == 0 {
		return image.Summary{}, fmt.Errorf("%w: no image found", ErrNotFound)
	}
	return images[0], nil
}

// ListImages 列出Docker镜像
//
// @description 列出当前系统中的Docker镜像
// @param imageListOpts ...func(args *[]filters.KeyValuePair) 镜像列表选项
// @return []image.Summary 镜像摘要列表
// @return error 列出过程中的错误
func (dc *Control) ListImages(imageListFilters ...func(args *[]filters.KeyValuePair)) ([]image.Summary, error) {
	dc.logger.Debugf("[docker/control] listing images...")
	if err := dc.ensureClient(); err != nil {
		return nil, err
	}

	// 列出网络选项
	ftArr := make([]filters.KeyValuePair, 0, 5)
	for _, opt := range imageListFilters {
		opt(&ftArr)
	}

	opts := image.ListOptions{
		All:     true,
		Filters: filters.NewArgs(ftArr...),
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	imageList, err := cli.ImageList(ctx, opts)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to list images: %v", err)
		return nil, err
	}
	return imageList, nil
}

// CreateContainer 创建Docker容器
//
// @description 创建新的Docker容器
// @param containerName string 容器名称
// @param imageName string 镜像名称
// @param config *container.Config 容器配置
// @param hostConfig *container.HostConfig 主机配置
// @param networkingConfig *network.NetworkingConfig 网络配置
// @return string 容器ID
// @return error 创建过程中的错误
func (dc *Control) CreateContainer(containerName string, imageName string,
	config *container.Config, hostConfig *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
) (string, error) {
	dc.logger.Infof("[docker/control] creating container: %s with image: %s", containerName, imageName)
	if err := dc.ensureClient(); err != nil {
		return "", err
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 创建容器
	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, containerName)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to create container %s: %v", containerName, err)
		return "", err
	}

	dc.logger.Infof("[docker/control] container %s created successfully, ID: %s", containerName, resp.ID)
	return resp.ID, nil
}

// StartContainer 启动Docker容器
//
// @description 启动指定的Docker容器
// @param containerID string 容器ID
// @return error 启动过程中的错误
func (dc *Control) StartContainer(containerID string) error {
	dc.logger.Infof("[docker/control] starting container: %s", containerID)
	if err := dc.ensureClient(); err != nil {
		return err
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 启动容器
	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		dc.logger.Errorf("[docker/control] failed to start container %s: %v", containerID, err)
		return err
	}

	dc.logger.Infof("[docker/control] container %s started successfully", containerID)
	return nil
}

// StopContainer 停止Docker容器
//
// @description 停止指定的Docker容器；timeout 为 0 时表示不指定超时（传 nil 给 Docker API，
//
//	使用 daemon 默认值），正数值为等待秒数，之后会被强制杀死。
//
// @param containerID string 容器ID
// @param timeout int 停止超时时间（秒）；0 表示使用 Docker daemon 默认值
// @return error 停止过程中的错误
func (dc *Control) StopContainer(containerID string, timeout int) error {
	dc.logger.Infof("[docker/control] stopping container: %s, timeout=%d", containerID, timeout)
	if err := dc.ensureClient(); err != nil {
		return err
	}

	// 停止容器容器选项
	opts := container.StopOptions{}
	if timeout > 0 {
		timeoutVal := timeout
		opts.Timeout = &timeoutVal
	}
	// timeout == 0 时传 nil，交由 Docker daemon 使用默认停止策略

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 停止容器
	if err := cli.ContainerStop(ctx, containerID, opts); err != nil {
		dc.logger.Errorf("[docker/control] failed to stop container %s: %v", containerID, err)
		return err
	}

	dc.logger.Infof("[docker/control] container %s stopped successfully", containerID)
	return nil
}

// RemoveContainer 删除Docker容器
//
// @description 删除指定的Docker容器
// @param containerID string 容器ID
// @param force bool 是否强制删除
// @return error 删除过程中的错误
func (dc *Control) RemoveContainer(containerID string, force bool) error {
	dc.logger.Infof("[docker/control] removing container: %s", containerID)

	if err := dc.ensureClient(); err != nil {
		return err
	}

	// 删除容器选项
	opts := container.RemoveOptions{
		Force: force,
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 删除容器
	if err := cli.ContainerRemove(ctx, containerID, opts); err != nil {
		dc.logger.Errorf("[docker/control] failed to remove container %s: %v", containerID, err)
		return err
	}

	dc.logger.Infof("[docker/control] container %s removed successfully", containerID)
	return nil
}

// ListContainers 列出Docker容器
//
// @description 列出当前系统中的Docker容器
// @param all bool 是否列出所有容器（包括已停止的）
// @param size bool 是否返回容器大小信息
// @param latest bool 是否仅返回最新创建的容器
// @param since string 仅返回该时间戳之后创建的容器
// @param before string 仅返回该时间戳之前创建的容器
// @param limit int 返回的最大容器数量，<=0 表示不限
// @param containerQueryFilters ...func(args *[]filters.KeyValuePair) 容器查询过滤器选项
// @return []container.Summary 容器列表
// @return error 列出过程中的错误
func (dc *Control) ListContainers(all bool, size bool, latest bool,
	since, before string, limit int, containerQueryFilters ...func(args *[]filters.KeyValuePair),
) ([]container.Summary, error) {
	dc.logger.Infof("[docker/control] listing containers, all: %v", all)
	if err := dc.ensureClient(); err != nil {
		return nil, err
	}

	ftArr := make([]filters.KeyValuePair, 0, 5)
	for _, opt := range containerQueryFilters {
		opt(&ftArr)
	}

	// 列出容器选项
	opts := container.ListOptions{
		Size:    size,
		All:     all,
		Latest:  latest,
		Since:   since,
		Before:  before,
		Limit:   limit,
		Filters: filters.NewArgs(ftArr...),
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 列出容器
	containers, err := cli.ContainerList(ctx, opts)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to list containers: %v", err)
		return nil, err
	}

	dc.logger.Infof("[docker/control] listed %d containers successfully", len(containers))
	return containers, nil
}

// GetContainer 获取Docker容器
//
// @description 根据容器名称获取容器摘要信息；未找到时返回 ErrNotFound，
//
//	调用方可通过 errors.Is(err, ErrNotFound) 判断。
//
// @param containerName string 容器名称
// @return container.Summary 容器摘要
// @return error 错误信息，未找到时为 ErrNotFound
func (dc *Control) GetContainer(containerName string) (container.Summary, error) {
	dc.logger.Infof("[docker/control] get container containerName: %v", containerName)
	if err := dc.ensureClient(); err != nil {
		return container.Summary{}, err
	}

	containers, err := dc.ListContainers(true, false, false, "", "", 1, WithContainerQueryName(containerName))
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to get container: %v", err)
		return container.Summary{}, err
	}
	if len(containers) == 0 {
		return container.Summary{}, fmt.Errorf("%w: container %s not found", ErrNotFound, containerName)
	}

	return containers[0], nil
}

// GetContainerStatus 获取Docker容器状态
//
// @description 获取指定Docker容器的运行状态
// @param containerID string 容器ID
// @return string 容器状态（created / running / paused / stopped / exited / dead 等）
// @return error 获取过程中的错误
func (dc *Control) GetContainerStatus(containerID string) (string, error) {
	dc.logger.Infof("[docker/control] getting status for container: %s", containerID)

	if err := dc.ensureClient(); err != nil {
		return "", err
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 查看容器详情
	inspect, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to inspect container %s: %v", containerID, err)
		return "", err
	}

	status := inspect.State.Status
	dc.logger.Infof("[docker/control] container %s status: %s", containerID, status)
	return status, nil
}

// ExecuteCommand 在容器中执行命令
//
// @description 在指定的Docker容器中执行命令；对 Docker exec attach 的多路复用流
//
//	使用 stdcopy.StdCopy 进行解复用，避免返回结果混入 8 字节帧头导致输出污染。
//	命令执行完成后会调用 ContainerExecInspect 检查退出码，非 0 时返回错误。
//
// @param containerID string 容器ID
// @param cmd []string 要执行的命令及参数
// @return string 命令标准输出内容（合并 stdout+stderr 的字符串）
// @return error 执行过程中的错误；若退出码非 0 会返回包含退出码的错误
func (dc *Control) ExecuteCommand(containerID string, cmd []string) (string, error) {
	dc.logger.Infof("[docker/control] executing command in container %s: %v", containerID, cmd)
	if err := dc.ensureClient(); err != nil {
		return "", err
	}

	// 执行命令选项（不分配 TTY，使用多路复用流）
	execConfig := container.ExecOptions{
		Cmd:          cmd,
		AttachStdout: true,
		AttachStderr: true,
	}

	ctx, cancel := dc.timeoutCtx()
	defer cancel()

	dc.mu.RLock()
	cli := dc.client
	dc.mu.RUnlock()

	// 创建执行
	resp, err := cli.ContainerExecCreate(ctx, containerID, execConfig)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to create exec in container %s: %v", containerID, err)
		return "", err
	}

	// Attach到执行
	execAttachOpts := container.ExecAttachOptions{}

	// 开始执行
	execResp, err := cli.ContainerExecAttach(ctx, resp.ID, execAttachOpts)
	if err != nil {
		dc.logger.Errorf("[docker/control] failed to start exec in container %s: %v", containerID, err)
		return "", err
	}
	defer execResp.Close()

	// 使用 stdcopy 解复用多路复用流（stdout/stderr 分离），避免混入 8 字节帧头
	var stdout, stderr bytes.Buffer
	_, copyErr := stdcopy.StdCopy(&stdout, &stderr, execResp.Reader)
	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		dc.logger.Errorf("[docker/control] failed to read exec output: %v", copyErr)
		return "", copyErr
	}

	// 检查 exec 退出码
	inspectCtx, inspectCancel := dc.timeoutCtx()
	defer inspectCancel()
	execInspect, inspectErr := cli.ContainerExecInspect(inspectCtx, resp.ID)
	if inspectErr != nil {
		dc.logger.Warnf("[docker/control] failed to inspect exec exit code: %v", inspectErr)
		// 非致命，继续返回已有输出
	} else if execInspect.ExitCode != 0 {
		errMsg := stderr.String()
		dc.logger.Errorf("[docker/control] command in container %s exited with code %d, stderr: %s",
			containerID, execInspect.ExitCode, errMsg)
		return stdout.String(), fmt.Errorf("exec command exited with code %d: %s", execInspect.ExitCode, errMsg)
	}

	dc.logger.Infof("[docker/control] command executed successfully in container %s", containerID)
	return stdout.String(), nil
}
