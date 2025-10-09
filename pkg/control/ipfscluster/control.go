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
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/ipfs/thridparty/shell"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"go.uber.org/zap"
)

type Control struct {
	config *config.IpfsClusterConfig
	logger *zap.SugaredLogger
	ctx    context.Context
	cancel context.CancelFunc
	sdk    client.Client
	once   sync.Once
}

func NewIpfsClusterControl(ipfsClusterConfig *config.IpfsClusterConfig, loggerControl *logger.Control) (*Control, error) {
	if ipfsClusterConfig == nil {
		ipfsClusterConfig = getDefaultConfig()
	}
	if !ipfsClusterConfig.Enabled {
		// 未启用
		return nil, errors.New("ipfs cluster is not enabled")
	}
	ipfsClusterLogger := loggerControl.GenLogger(logger.ModuleIpfsCluster)

	ctx, cancel := context.WithCancel(context.Background())

	control := &Control{
		config: ipfsClusterConfig,
		logger: ipfsClusterLogger,
		ctx:    ctx,
		cancel: cancel,
	}
	return control, nil
}

func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		if c.config.Enabled {
			c.logger.Debugf("[control] starting up ipfs cluster control...")
			if err := c.initSDK(); err != nil {
				c.logger.Errorf("[control] failed to init ipfs cluster sdk, err: %s", err.Error())
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
			version, err := c.Version()
			if err != nil {
				c.logger.Errorf("[control] failed to get ipfs cluster version, err: %s", err.Error())
				if failedFunc != nil {
					failedFunc(err)
				}
				return
			}
			c.logger.Debugf("[control] ipfs cluster version: %s", version)
		}
	})
}

func (c *Control) Shutdown() error {
	c.logger.Debugf("[control] shutting down ipfs cluster control...")
	c.cancel()
	return nil
}

func (c *Control) initConfigs() []*client.Config {
	c.logger.Debugf("[control] init ipfs cluster clients configs...")
	configs := make([]*client.Config, 0, len(c.config.Addresses))

	for _, address := range c.config.Addresses {
		if stringer.IsBlank(address.Host) || address.Port == 0 {
			c.logger.Warnf("[control] invalid ipfs cluster address [%s:%d]", address.Host, address.Port)
			continue
		}
		configs = append(configs, &client.Config{
			Host:              address.Host,
			Port:              strconv.Itoa(address.Port),
			Username:          address.UserName,
			Password:          address.Password,
			Timeout:           time.Duration(c.config.TimeOutInterval) * time.Second,
			DisableKeepAlives: true,
			LogLevel:          "info",
		})
	}
	return configs
}

func (c *Control) initSDK() error {
	c.logger.Debugf("[control] init ipfs cluster clients...")
	configs := c.initConfigs()
	if len(configs) == 0 {
		c.logger.Errorf("[control] no valid ipfs cluster address")
		return errors.New("no valid ipfs cluster address")
	}

	switch StrategyType(strings.ToLower(c.config.Strategy)) {
	case RoundRobin:
		// 使用节点轮询的方式
		sdk, err := client.NewLBClient(&client.RoundRobin{}, configs, 5)
		if err != nil {
			c.logger.Errorf("[control] init ipfs cluster client failed: %s", err)
			return err
		}
		c.sdk = sdk
	case FailOver:
		// 使用节点故障转移的方式
		sdk, err := client.NewLBClient(&client.Failover{}, configs, 5)
		if err != nil {
			c.logger.Errorf("[control] init ipfs cluster client failed: %s", err)
			return err
		}
		c.sdk = sdk
	default:
		sdk, err := client.NewDefaultClient(configs[0])
		if err != nil {
			c.logger.Errorf("[control] init ipfs cluster client failed: %s", err)
			return err
		}
		c.sdk = sdk
	}

	return nil
}

func (c *Control) Version() (string, error) {
	c.logger.Debugf("[control] get ipfs cluster version...")
	version, err := c.sdk.Version(c.ctx)
	if err != nil {
		c.logger.Errorf("[control] get ipfs cluster version failed: %s", err)
		return "", err
	}
	return version.Version, nil
}

func (c *Control) AddDirectory(dir string, replicationMin, replicationMax int, expireAt time.Duration) ([]api.AddedOutput, error) {
	c.logger.Debugf("[control] add directory to ipfs cluster...")
	
	var results []api.AddedOutput
	var errors []error
	
	// 遍历目录，收集所有文件
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			c.logger.Errorf("[control] access path %s failed: %s", path, err)
			return err
		}
		
		// 如果是文件而不是目录，则添加到文件列表
		if !info.IsDir() {
			files = append(files, path)
		}
		
		return nil
	})
	
	if err != nil {
		c.logger.Errorf("[control] walk directory %s failed: %s", dir, err)
		return nil, err
	}
	
	c.logger.Debugf("[control] found %d files in directory %s", len(files), dir)
	
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
					c.logger.Errorf("[control] get relative path for %s failed: %s", filePath, err)
					errorChan <- err
					continue
				}
				
				// 调用 AddOneFile 方法添加单个文件
				result, err := c.AddOneFile(relPath, filePath, replicationMin, replicationMax, expireAt)
				if err != nil {
					c.logger.Errorf("[control] add file %s failed: %s", filePath, err)
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
			c.logger.Debugf("[control] added file successfully, cid: %s", result.Cid)
		case err := <-errorChan:
			errors = append(errors, err)
		case <-c.ctx.Done():
			c.logger.Errorf("[control] add directory timeout or cancelled")
			return results, c.ctx.Err()
		}
	}
	
	c.logger.Debugf("[control] add directory completed. Success: %d, Failures: %d", len(results), len(errors))
	
	// 如果有错误，返回部分结果和错误信息
	if len(errors) > 0 {
		// 可以根据需要决定是否返回错误
		// 这里选择记录错误但返回已成功添加的文件
		c.logger.Warnf("[control] add directory completed with %d errors", len(errors))
	}
	
	return results, nil
}

func (c *Control) AddOneFile(fileName, filePath string, replicationMin, replicationMax int, expireAt time.Duration) (api.AddedOutput, error) {
	c.logger.Debugf("[control] add one file to ipfs cluster...")
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
			c.logger.Errorf("[control] add one file to ipfs cluster failed: %s", err)
			errCh <- err
			return
		}
	}()

	select {
	case result := <-out:
		c.logger.Debugf("[control] add one file to ipfs cluster success, cid: %s", result.Cid)
		return result, nil
	case err := <-errCh:
		if err != nil {
			c.logger.Errorf("[control] add one file to ipfs cluster failed: %s", err)
			return api.AddedOutput{}, err
		}
		// 如果errCh中没有错误，继续等待结果
		select {
		case result := <-out:
			c.logger.Debugf("[control] add one file to ipfs cluster success, cid: %s", result.Cid)
			return result, nil
		case <-c.ctx.Done():
			c.logger.Errorf("[control] add one file to ipfs cluster timeout or cancelled")
			return api.AddedOutput{}, c.ctx.Err()
		}
	case <-c.ctx.Done():
		c.logger.Errorf("[control] add one file to ipfs cluster timeout or cancelled")
		return api.AddedOutput{}, c.ctx.Err()
	}
}

func (c *Control) CatOneFile(cid string) (io.Reader, error) {
	c.logger.Debugf("[control] cat one file from ipfs cluster, cid: %s", cid)
	sh := c.sdk.IPFS(c.ctx)
	c.logger.Debugf("[control] setting timeout for ipfs cluster cat command")
	sh.SetTimeout(time.Duration(c.config.TimeOutInterval) * time.Second)
	cat, err := shell.Cat(sh, c.ctx, cid, true)
	if err != nil {
		c.logger.Errorf("[control] cat one file from ipfs cluster failed: %s", err)
		return nil, err
	}
	return cat, nil
}

func (c *Control) DeleteOneFile(cid string) error {
	c.logger.Debugf("[control] delete one file from ipfs cluster, cid: %s", cid)
	deleteCid, err := api.DecodeCid(cid)
	if err != nil {
		c.logger.Errorf("[control] decode cid failed: %v", err)
		return err
	}

	_, err = c.sdk.Unpin(c.ctx, deleteCid)
	if err != nil {
		c.logger.Errorf("[control] unpin cid failed: %v", err)
		return err
	}
	return nil
}

func (c *Control) RepoGC(local bool) error {
	c.logger.Debugf("[control] repo gc...")
	_, err := c.sdk.RepoGC(c.ctx, local)
	if err != nil {
		c.logger.Errorf("[control] repo gc failed: %v", err)
		return err
	}

	// sh := shell.NewShell(targetNode)
	// err = sh.Request("repo/gc", cid).
	// 	Option("recursive", true).
	// 	Exec(ctx, nil)

	return nil
}

func (c *Control) DownloadOneFile(cid string, outPath string) error {
	c.logger.Debugf("[control] download one file from ipfs cluster, cid: %s, outPath: %s", cid, outPath)
	ipfs := c.sdk.IPFS(c.ctx)
	ipfs.SetTimeout(time.Duration(c.config.TimeOutInterval) * time.Second)
	if err := shell.Get(ipfs, c.ctx, cid, outPath, true); err != nil {
		c.logger.Errorf("[control] download one file from ipfs cluster failed: %v", err)
		return err
	}
	return nil
}
