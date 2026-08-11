package ipfscluster

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	"github.com/jianlu8023/go-tools/v2/pkg/check"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/ipfs/thridparty/shell"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

// Control IPFS集群控制器结构体
// @description 管理IPFS集群客户端连接和操作的控制器
type Control struct {
	config *config.IpfsClusterConfig // IPFS集群配置
	logger *zap.SugaredLogger        // 日志记录器
	ctx    context.Context           // 上下文
	cancel context.CancelFunc        // 取消函数
	sdk    client.Client             // IPFS集群客户端
	once   sync.Once                 // 确保StartUp只执行一次
}

// NewIpfsClusterControl 创建IPFS集群控制器
// @description 根据配置创建一个新的IPFS集群控制器实例
// @param ipfsClusterConfig *config.IpfsClusterConfig IPFS集群配置
// @param loggerControl *logger.Control 日志控制器
// @return *Control IPFS集群控制器实例
// @return error 创建过程中可能产生的错误
func NewIpfsClusterControl(ipfsClusterConfig *config.IpfsClusterConfig, loggerControl *logger.Control) (*Control, error) {
	ipfsClusterConfig = check.IF[*config.IpfsClusterConfig](ipfsClusterConfig == nil,
		getDefaultConfig(),
		ipfsClusterConfig,
	)
	if !ipfsClusterConfig.Enabled {
		// 未启用
		return nil, errors.New("ipfs cluster is not enabled")
	}
	loggerControl = check.IF[*logger.Control](loggerControl == nil,
		logger.NewLoggerControl(&config.LoggerConfig{
			DefaultLogLevel: "debug",
			PrintFormat:     "console",
		}),
		loggerControl,
	)
	// ipfsClusterLogger := loggerControl.GenLogger(logger.ModuleIpfsCluster)
	ipfsClusterLogger := loggerControl.GenLogger("")

	ctx, cancel := context.WithCancel(context.Background())

	control := &Control{
		config: ipfsClusterConfig,
		logger: ipfsClusterLogger,
		ctx:    ctx,
		cancel: cancel,
	}
	return control, nil
}

// StartUp 启动IPFS集群控制器
// @description 初始化SDK并获取IPFS集群版本信息
// @param failedFunc func(err error) 启动失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Debugf("[ipfscluster/control] starting up ipfs cluster control...")
			if err := c.initSDK(); err != nil {
				c.logger.Errorf("[ipfscluster/control] failed to init ipfs cluster sdk, err: %s", err.Error())
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
			version, err := c.Version()
			if err != nil {
				c.logger.Errorf("[ipfscluster/control] failed to get ipfs cluster version, err: %s", err.Error())
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
			c.logger.Debugf("[ipfscluster/control] ipfs cluster version: %s", version)
		}
	})
}

// Shutdown 关闭IPFS集群控制器
// @description 关闭IPFS集群控制器并释放资源
// @return error 关闭过程中可能产生的错误
func (c *Control) Shutdown() error {
	c.logger.Debugf("[ipfscluster/control] shutting down ipfs cluster control...")
	c.cancel()
	return nil
}

// initConfigs 初始化IPFS集群客户端配置
// @description 根据配置信息初始化IPFS集群客户端配置列表
// @return []*client.Config IPFS集群客户端配置列表
func (c *Control) initConfigs() []*client.Config {
	c.logger.Debugf("[ipfscluster/control] init ipfs cluster clients configs...")
	configs := make([]*client.Config, 0, len(c.config.Addresses))

	for _, address := range c.config.Addresses {
		if stringer.IsBlank(address.Host) || address.Port == 0 {
			c.logger.Warnf("[ipfscluster/control] invalid ipfs cluster address [%s:%d]", address.Host, address.Port)
			continue
		}
		configs = append(configs, &client.Config{
			Host:              address.Host,
			Port:              strconv.Itoa(address.Port),
			Username:          address.UserName,
			Password:          address.Password,
			Timeout:           time.Duration(c.config.TimeOutInterval) * time.Second,
			DisableKeepAlives: true,
			LogLevel:          c.config.LogLevel,
		})
	}
	return configs
}

// initSDK 初始化IPFS集群SDK客户端
// @description 根据配置和负载均衡策略初始化IPFS集群SDK客户端
// @return error 初始化过程中可能产生的错误
func (c *Control) initSDK() error {
	c.logger.Debugf("[ipfscluster/control] init ipfs cluster clients...")
	configs := c.initConfigs()
	if len(configs) == 0 {
		c.logger.Errorf("[ipfscluster/control] no valid ipfs cluster address")
		return errors.New("no valid ipfs cluster address")
	}

	switch strategyType(strings.ToLower(c.config.Strategy)) {
	case roundRobin:
		// 使用节点轮询的方式
		sdk, err := client.NewLBClient(&client.RoundRobin{}, configs, c.config.ReTries)
		if err != nil {
			c.logger.Errorf("[ipfscluster/control] init ipfs cluster client failed: %s", err)
			return err
		}
		c.sdk = sdk
	case failOver:
		// 使用节点故障转移的方式
		sdk, err := client.NewLBClient(&client.Failover{}, configs, c.config.ReTries)
		if err != nil {
			c.logger.Errorf("[ipfscluster/control] init ipfs cluster client failed: %s", err)
			return err
		}
		c.sdk = sdk
	default:
		c.logger.Warnf("[ipfscluster/control] unknown strategy type: %s, use address[0] create default sdk...", c.config.Strategy)
		sdk, err := client.NewDefaultClient(configs[0])
		if err != nil {
			c.logger.Errorf("[ipfscluster/control] init ipfs cluster client failed: %s", err)
			return err
		}
		c.sdk = sdk
	}

	return nil
}

// Version 获取IPFS集群版本信息
// @description 获取IPFS集群的版本信息
// @return string IPFS集群版本号
// @return error 获取过程中可能产生的错误
func (c *Control) Version() (string, error) {
	c.logger.Debugf("[ipfscluster/control] get ipfs cluster version...")
	version, err := c.sdk.Version(c.ctx)
	if err != nil {
		c.logger.Errorf("[ipfscluster/control] get ipfs cluster version failed: %s", err)
		return "", err
	}
	return version.Version, nil
}

// AddDirectory 添加目录到IPFS集群
// @description 将指定目录下的所有文件添加到IPFS集群中
// @param dir string 目录路径
// @param replicationMin int 最小复制因子
// @param replicationMax int 最大复制因子
// @param expireAt time.Duration 过期时间
// @return []api.AddedOutput 添加结果列表
// @return error 添加过程中可能产生的错误
func (c *Control) AddDirectory(dir string, replicationMin, replicationMax int, expireAt time.Duration) ([]api.AddedOutput, error) {
	c.logger.Debugf("[ipfscluster/control] add directory to ipfs cluster...")

	var results []api.AddedOutput
	var errors []error

	// 遍历目录，收集所有文件
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.logger.Errorf("[ipfscluster/control] access path %s failed: %s", path, err)
			return err
		}

		// 如果是文件而不是目录，则添加到文件列表
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		c.logger.Errorf("[ipfscluster/control] walk directory %s failed: %s", dir, err)
		return nil, err
	}

	c.logger.Debugf("[ipfscluster/control] found %d files in directory %s", len(files), dir)

	// 使用带缓冲的通道来控制并发数
	const maxWorkers = 5
	fileChan := make(chan string, len(files))
	resultChan := make(chan api.AddedOutput, len(files))
	errorChan := make(chan error, len(files))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				// 获取相对路径作为文件名
				relPath, err := filepath.Rel(dir, filePath)
				if err != nil {
					c.logger.Errorf("[ipfscluster/control] get relative path for %s failed: %s", filePath, err)
					errorChan <- err
					continue
				}

				// 调用 AddOneFile 方法添加单个文件
				result, err := c.AddOneFile(relPath, filePath, replicationMin, replicationMax, expireAt)
				if err != nil {
					c.logger.Errorf("[ipfscluster/control] add file %s failed: %s", filePath, err)
					errorChan <- err
					continue
				}

				resultChan <- result
			}
		}()
	}

	// 发送文件到工作协程
	go func() {
		defer close(fileChan)
		for _, file := range files {
			fileChan <- file
		}
	}()

	// 关闭结果和错误通道
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// 收集结果和错误
	for i := 0; i < len(files); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
			c.logger.Debugf("[ipfscluster/control] added file successfully, cid: %s", result.Cid)
		case err := <-errorChan:
			errors = append(errors, err)
		case <-c.ctx.Done():
			c.logger.Errorf("[ipfscluster/control] add directory timeout or cancelled")
			return results, c.ctx.Err()
		}
	}

	c.logger.Debugf("[ipfscluster/control] add directory completed. Success: %d, Failures: %d", len(results), len(errors))

	// 如果有错误，返回部分结果和错误信息
	if len(errors) > 0 {
		// 可以根据需要决定是否返回错误
		// 这里选择记录错误但返回已成功添加的文件
		c.logger.Warnf("[ipfscluster/control] add directory completed with %d errors", len(errors))
	}

	return results, nil
}

// AddOneFile 添加单个文件到IPFS集群
// @description 将指定文件添加到IPFS集群中
// @param fileName string 文件名
// @param filePath string 文件路径
// @param replicationMin int 最小复制因子
// @param replicationMax int 最大复制因子
// @param expireAt time.Duration 过期时间
// @return api.AddedOutput 添加结果
// @return error 添加过程中可能产生的错误
func (c *Control) AddOneFile(fileName, filePath string, replicationMin, replicationMax int, expireAt time.Duration) (api.AddedOutput, error) {
	c.logger.Debugf("[ipfscluster/control] add one file to ipfs cluster...")
	defaultParams := api.DefaultAddParams()
	defaultParams.CidVersion = 1
	defaultParams.ExpireAt = time.Now().Add(expireAt)
	if replicationMin <= 0 {
		replicationMin = 1
	}
	if replicationMax <= 0 {
		replicationMax = 1
	}
	defaultParams.ReplicationFactorMin = replicationMin
	defaultParams.ReplicationFactorMax = replicationMax
	defaultParams.Metadata = map[string]string{
		"fileName":   fileName,
		"uploadTime": time.Now().Format(timeFormat),
	}
	defaultParams.Name = fileName
	out := make(chan api.AddedOutput, 1)
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)
		defer close(out)
		err := c.sdk.Add(c.ctx, []string{filePath}, defaultParams, out)
		if err != nil {
			c.logger.Errorf("[ipfscluster/control] add one file to ipfs cluster failed: %s", err)
			errCh <- err
			return
		}
	}()

	select {
	case result := <-out:
		c.logger.Debugf("[ipfscluster/control] add one file to ipfs cluster success, cid: %s", result.Cid)
		return result, nil
	case err := <-errCh:
		if err != nil {
			c.logger.Errorf("[ipfscluster/control] add one file to ipfs cluster failed: %s", err)
			return api.AddedOutput{}, err
		}
		// 如果errCh中没有错误，继续等待结果
		select {
		case result := <-out:
			c.logger.Debugf("[ipfscluster/control] add one file to ipfs cluster success, cid: %s", result.Cid)
			return result, nil
		case <-c.ctx.Done():
			c.logger.Errorf("[ipfscluster/control] add one file to ipfs cluster timeout or cancelled")
			return api.AddedOutput{}, c.ctx.Err()
		}
	case <-c.ctx.Done():
		c.logger.Errorf("[ipfscluster/control] add one file to ipfs cluster timeout or cancelled")
		return api.AddedOutput{}, c.ctx.Err()
	}
}

// CatOneFile 获取IPFS集群中的文件内容
// @description 根据CID获取IPFS集群中的文件内容
// @param cid string 文件的CID
// @return io.Reader 文件内容读取器
// @return error 获取过程中可能产生的错误
func (c *Control) CatOneFile(cid string) (io.Reader, error) {
	c.logger.Debugf("[ipfscluster/control] cat one file from ipfs cluster, cid: %s", cid)
	sh := c.sdk.IPFS(c.ctx)
	c.logger.Debugf("[ipfscluster/control] setting timeout for ipfs cluster cat command")
	sh.SetTimeout(time.Duration(c.config.TimeOutInterval) * time.Second)
	cat, err := shell.Cat(sh, c.ctx, cid, true)
	if err != nil {
		c.logger.Errorf("[ipfscluster/control] cat one file from ipfs cluster failed: %s", err)
		return nil, err
	}
	return cat, nil
}

// DeleteOneFile 从IPFS集群中删除文件
// @description 根据CID从IPFS集群中删除文件
// @param cid string 文件的CID
// @return error 删除过程中可能产生的错误
func (c *Control) DeleteOneFile(cid string) error {
	c.logger.Debugf("[ipfscluster/control] delete one file from ipfs cluster, cid: %s", cid)
	deleteCid, err := api.DecodeCid(cid)
	if err != nil {
		c.logger.Errorf("[ipfscluster/control] decode cid failed: %v", err)
		return err
	}

	_, err = c.sdk.Unpin(c.ctx, deleteCid)
	if err != nil {
		c.logger.Errorf("[ipfscluster/control] unpin cid failed: %v", err)
		return err
	}
	return nil
}

// RepoGC 执行仓库垃圾回收
// @description 执行IPFS集群节点的仓库垃圾回收操作
// @param local bool 是否只在本地执行
// @return error 执行过程中可能产生的错误
func (c *Control) RepoGC(local bool) error {
	c.logger.Debugf("[ipfscluster/control] repo gc...")
	_, err := c.sdk.RepoGC(c.ctx, local)
	if err != nil {
		c.logger.Errorf("[ipfscluster/control] repo gc failed: %v", err)
		return err
	}

	// sh := shell.NewShell(targetNode)
	// err = sh.Request("repo/gc", cid).
	// 	Option("recursive", true).
	// 	Exec(ctx, nil)

	return nil
}

// DownloadOneFile 从IPFS集群下载文件
// @description 根据CID从IPFS集群下载文件到指定路径
// @param cid string 文件的CID
// @param outPath string 输出文件路径
// @return error 下载过程中可能产生的错误
func (c *Control) DownloadOneFile(cid string, outPath string) error {
	c.logger.Debugf("[ipfscluster/control] download one file from ipfs cluster, cid: %s, outPath: %s", cid, outPath)
	ipfs := c.sdk.IPFS(c.ctx)
	ipfs.SetTimeout(time.Duration(c.config.TimeOutInterval) * time.Second)
	if err := shell.Get(ipfs, c.ctx, cid, outPath, true); err != nil {
		c.logger.Errorf("[ipfscluster/control] download one file from ipfs cluster failed: %v", err)
		return err
	}
	return nil
}
