package ipfs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/ipfs/boxo/files"
	"github.com/ipfs/boxo/path"
	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	ipfsrpc "github.com/ipfs/kubo/client/rpc"
	"github.com/ipfs/kubo/core/coreiface/options"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/config"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/control/logger"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

// Control IPFS控制结构体

type Control struct {
	shell  *shell.Shell
	client *ipfsrpc.HttpApi
	config *config.IpfsConfig
	logger *zap.SugaredLogger
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

// NewIpfsControl 创建一个新的IPFS控制器
//
// @param ipfsConfig *config.IpfsConfig IPFS配置
// @param loggerControl *logger.Control 日志控制器
//
// @return *Control IPFS控制器
// @return error 新建过程中的错误
func NewIpfsControl(ipfsConfig *config.IpfsConfig, loggerControl *logger.Control) (*Control, error) {
	ipfsLogger := loggerControl.GenLogger(logger.ModuleIpfs)
	ipfsLogger.Infof("[control] starting new IPFS control...")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化IpfsControl
	ipfs := &Control{
		logger: ipfsLogger,
		ctx:    ctx,
		cancel: cancel,
		config: ipfsConfig,
	}

	// 初始化IPFS客户端
	if err := ipfs.initClient(); err != nil {
		ipfsLogger.Errorf("[control] IPFS initialization client failed: %v", err)
		cancel()
		return nil, err
	}

	ipfsLogger.Infof("[control] IPFS control started successfully")
	return ipfs, nil
}

// initClient 初始化IPFS客户端
//
// @description 初始化IPFS客户端连接
// @return error 初始化过程中的错误
func (ipfs *Control) initClient() error {
	ipfs.logger.Debugf("[control] initializing IPFS client...")

	// 如果配置了API地址，使用指定的API地址
	// apiAddr := "localhost:5001"
	// if ipfs.config.ApiAddress != "" {
	// 	apiAddr = ipfs.config.ApiAddress
	// }

	// 构建API URL
	// apiURL := fmt.Sprintf("http://%s/api/v0", apiAddr)

	// 创建IPFS HTTP客户端
	addr, err := multiaddr.NewMultiaddr(ipfs.config.ApiAddress)
	if err != nil {
		ipfs.logger.Errorf("[control] unable to parse API address: %v", err)
		return err
	}
	ipfs.logger.Debugf("[control] connecting to IPFS API at %s", addr)
	newApi, err := ipfsrpc.NewApi(addr)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to create IPFS client: %v", err)
		return err
	}

	ipfs.client = newApi
	return nil
}

// StartUp 启动IPFS服务
//
// @description 启动IPFS服务，测试连接是否正常
// @param failedFunc func(err error) 启动失败回调函数
func (ipfs *Control) StartUp(failedFunc func(err error)) {
	ipfs.once.Do(func() {
		ipfs.logger.Infof("[control] starting IPFS service...")

		// 测试连接
		if err := ipfs.testConnection(); err != nil {
			ipfs.logger.Errorf("[control] failed to connect to IPFS daemon: %v", err)
			failedFunc(err)
			return
		}

		ipfs.logger.Infof("[control] IPFS service started successfully")
	})
}

// Shutdown 关闭IPFS客户端连接
//
// @description 关闭IPFS客户端连接并释放资源
// @return error 关闭过程中的错误
func (ipfs *Control) Shutdown() error {
	ipfs.logger.Infof("[control] shutting down IPFS control...")
	ipfs.cancel()
	// IPFS HTTP客户端不需要显式关闭
	ipfs.logger.Infof("[control] IPFS control shutdown successfully")
	return nil
}

// testConnection 测试IPFS连接是否正常
//
// @description 测试与IPFS节点的连接是否正常
// @return error 测试过程中的错误
func (ipfs *Control) testConnection() error {
	ipfs.logger.Debugf("[control] testing IPFS connection...")

	// 获取IPFS版本信息来测试连接
	version, commit, err := ipfs.shell.Version()
	if err != nil {
		ipfs.logger.Errorf("[control] IPFS version check failed: %v", err)
		return err
	}

	ipfs.logger.Infof("[control] connected to IPFS node, version: %s commit: %v", version, commit)
	return nil
}

// UploadFile 上传文件到IPFS
//
// @param filePath string 要上传的文件路径
// @return string 文件在IPFS中的CID
// @return error 上传过程中的错误
func (ipfs *Control) UploadFile(filePath string) (string, error) {
	ipfs.logger.Infof("[control] uploading file to IPFS: %s", filePath)

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		ipfs.logger.Errorf("[control] file does not exist: %v", err)
		return "", err
	}

	// 检查是否是目录
	if fileInfo.IsDir() {
		ipfs.logger.Debugf("[control] uploading directory: %s", filePath)
		// 对于目录，递归上传
		return ipfs.uploadDirectory(filePath)
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to open file: %v", err)
		return "", err
	}
	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			ipfs.logger.Errorf("[control] failed to close file: %v", err)
		}
	}(file)

	// 上传文件到IPFS
	readerFile := files.NewReaderFile(file)
	immutablePath, err := ipfs.client.Unixfs().Add(
		ipfs.ctx,
		readerFile,
		options.Unixfs.CidVersion(1),
		options.Unixfs.Pin(true),
	)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to upload file: %v", err)
		return "", err
	}

	cidStr := immutablePath.String()
	ipfs.logger.Infof("[control] file uploaded successfully, CID: %s", cidStr)
	return cidStr, nil
}

// uploadDirectory 上传目录到IPFS
//
// @param dirPath string 要上传的目录路径
// @return string 目录在IPFS中的根CID
// @return error 上传过程中的错误
func (ipfs *Control) uploadDirectory(dirPath string) (string, error) {
	ipfs.logger.Debugf("[control] uploading directory recursively: %s", dirPath)

	// 构建一个目录添加操作
	addDir := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 忽略目录本身，只处理其中的文件
		if !info.IsDir() {
			ipfs.logger.Debugf("[control] uploading file in directory: %s", path)
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				if err := file.Close(); err != nil {
					ipfs.logger.Errorf("[control] failed to close file: %v", err)
				}
			}(file)

			// 上传文件
			readerFile := files.NewReaderFile(file)
			_, err = ipfs.client.Unixfs().Add(ipfs.ctx, readerFile,
				options.Unixfs.CidVersion(1),
				options.Unixfs.Pin(true),
			)
			return err
		}
		return nil
	}

	// 遍历目录并上传所有文件
	if err := filepath.Walk(dirPath, addDir); err != nil {
		ipfs.logger.Errorf("[control] failed to upload directory: %v", err)
		return "", err
	}

	// 注意：这个简化实现会返回最后上传文件的CID
	// 实际项目中可能需要使用MFS (Mutable File System) 来创建目录结构
	// 并返回整个目录的根CID
	ipfs.logger.Infof("[control] directory upload process completed: %s", dirPath)

	// 这里返回一个临时值，实际实现需要调整
	return "", fmt.Errorf("directory upload not fully implemented")
}

// DownloadFile 从IPFS下载文件
//
// @param cidStr string IPFS中的文件CID
// @param outputPath string 下载后的文件保存路径
// @return error 下载过程中的错误
func (ipfs *Control) DownloadFile(cidStr string, outputPath string) error {
	ipfs.logger.Infof("[control] downloading file from IPFS, CID: %s", cidStr)

	// 解析CID
	decode, err := cid.Decode(cidStr)
	if err != nil {
		ipfs.logger.Errorf("[control] invalid CID format: %v", err)
		return err
	}

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to create output file: %v", err)
		return err
	}
	defer func(outputFile *os.File) {
		err := outputFile.Close()
		if err != nil {
			ipfs.logger.Errorf("[control] failed to close output file: %v", err)
		}
	}(outputFile)

	// 从IPFS获取文件内容
	ipfsPath := path.FromCid(decode)
	reader, err := ipfs.client.Unixfs().Get(ipfs.ctx, ipfsPath)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to get file from IPFS: %v", err)
		return err
	}
	defer func(reader files.Node) {
		err := reader.Close()
		if err != nil {
			ipfs.logger.Errorf("[control] failed to close file: %v", err)
		}
	}(reader)

	// 写入文件
	toFile := files.ToFile(reader)
	size, err := io.Copy(outputFile, toFile)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to write to output file: %v", err)
		return err
	}

	ipfs.logger.Infof("[control] file downloaded successfully, size: %d bytes, saved to: %s", size, outputPath)
	return nil
}

// GetFileContent 获取IPFS文件内容
//
// @param cidStr string IPFS中的文件CID
// @return []byte 文件内容
// @return error 获取过程中的错误
func (ipfs *Control) GetFileContent(cidStr string) ([]byte, error) {
	ipfs.logger.Infof("[control] getting file content from IPFS, CID: %s", cidStr)

	// 解析CID
	decode, err := cid.Decode(cidStr)
	if err != nil {
		ipfs.logger.Errorf("[control] invalid CID format: %v", err)
		return nil, err
	}

	// 从IPFS获取文件内容
	ipfsPath := path.FromCid(decode)
	reader, err := ipfs.client.Unixfs().Get(ipfs.ctx, ipfsPath)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to get file from IPFS: %v", err)
		return nil, err
	}
	defer func(reader files.Node) {
		if err := reader.Close(); err != nil {
			ipfs.logger.Errorf("[control] failed to close file: %v", err)
		}
	}(reader)

	// 读取内容
	toFile := files.ToFile(reader)
	defer func(toFile files.File) {
		if err := toFile.Close(); err != nil {
			ipfs.logger.Errorf("[control] failed to close file: %v", err)
		}
	}(toFile)
	content, err := io.ReadAll(toFile)
	if err != nil {
		ipfs.logger.Errorf("[control] failed to read file content: %v", err)
		return nil, err
	}

	ipfs.logger.Infof("[control] file content retrieved successfully, size: %d bytes", len(content))
	return content, nil
}
