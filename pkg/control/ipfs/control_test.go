package ipfs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jianlu8023/golang-example/pkg/control/config"
	"github.com/jianlu8023/golang-example/pkg/control/logger"
	"github.com/stretchr/testify/assert"
)

// TestNewIpfsControl 测试创建IPFS控制器
func TestNewIpfsControl(t *testing.T) {
	// 创建日志控制器
	loggerControl := logger.NewLoggerControl(&config.LoggerConfig{
		DefaultLogLevel: "debug",
		StackLogLevel:   "fatal",
		PrintFormat:     "console",
		FilePath:        "logs/app.log",
		MaxAge:          7,
		RotationTime:    1,
		LoggerLevel: map[string]string{
			"Ipfs": "debug",
		},
	})

	defer loggerControl.Shutdown()

	// 创建IPFS配置
	ipfsConfig := &config.IpfsConfig{
		Enabled:        true,
		ApiAddress:     "/ip4/127.0.0.1/tcp/5001",
		GatewayAddress: "/ip4/127.0.0.1/tcp/8080",
	}

	// 创建IPFS控制器
	ipfsControl, err := NewIpfsControl(ipfsConfig, loggerControl)
	assert.NoError(t, err)
	defer func() {
		shutdownErr := ipfsControl.Shutdown()
		assert.NoError(t, shutdownErr)
	}()

	assert.NotNil(t, ipfsControl)
}

// TestIpfsUploadDownload 测试IPFS上传和下载功能
// 注意：此测试需要本地运行IPFS守护进程
func TestIpfsUploadDownload(t *testing.T) {
	// 检查是否在CI环境中运行，如果是则跳过测试
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping IPFS integration test in CI environment")
	}

	// 创建日志控制器
	loggerControl := logger.NewLoggerControl(&config.LoggerConfig{
		DefaultLogLevel: "debug",
		StackLogLevel:   "fatal",
		PrintFormat:     "console",
		FilePath:        "logs/app.log",
		MaxAge:          7,
		RotationTime:    1,
		LoggerLevel: map[string]string{
			"Ipfs": "debug",
		},
	})
	defer loggerControl.Shutdown()

	// 创建IPFS配置
	ipfsConfig := &config.IpfsConfig{
		Enabled:    true,
		ApiAddress: "localhost:5001",
	}

	// 创建IPFS控制器
	ipfsControl, err := NewIpfsControl(ipfsConfig, loggerControl)
	assert.NoError(t, err)
	defer func() {
		shutdownErr := ipfsControl.Shutdown()
		assert.NoError(t, shutdownErr)
	}()

	// 启动IPFS服务
	hasFailed := false
	ipfsControl.StartUp(func(err error) {
		hasFailed = true
		t.Logf("IPFS service failed to start: %v", err)
	})

	// 如果启动失败，跳过测试
	if hasFailed {
		t.Skip("Skipping IPFS upload/download test because IPFS service failed to start")
	}

	// 创建临时测试文件
	tempDir, err := os.MkdirTemp("", "ipfs-test*")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// 创建测试文件
	testFilePath := filepath.Join(tempDir, "test.txt")
	testContent := []byte("This is a test file for IPFS control\n")
	err = os.WriteFile(testFilePath, testContent, os.FileMode(0o644))
	assert.NoError(t, err)

	// 上传文件
	cid, err := ipfsControl.UploadFile(testFilePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, cid)
	t.Logf("File uploaded successfully, CID: %s", cid)

	// 下载文件
	downloadPath := filepath.Join(tempDir, "downloaded.txt")
	err = ipfsControl.DownloadFile(cid, downloadPath)
	assert.NoError(t, err)

	// 读取下载的文件
	downloadedContent, err := os.ReadFile(downloadPath)
	assert.NoError(t, err)

	// 验证内容是否一致
	assert.Equal(t, testContent, downloadedContent)
	t.Logf("File downloaded successfully, content matches original")

	// 测试获取文件内容
	content, err := ipfsControl.GetFileContent(cid)
	assert.NoError(t, err)
	assert.Equal(t, testContent, content)
	t.Logf("File content retrieved successfully")
}

// TestIpfsControl_StartUpShutdown 测试IPFS服务的启动和关闭
func TestIpfsControl_StartUpShutdown(t *testing.T) {
	// 创建日志控制器
	loggerControl := logger.NewLoggerControl(&config.LoggerConfig{
		DefaultLogLevel: "debug",
		StackLogLevel:   "fatal",
		PrintFormat:     "console",
		FilePath:        "logs/app.log",
		MaxAge:          7,
		RotationTime:    1,
		LoggerLevel: map[string]string{
			"Ipfs": "debug",
		},
	})
	defer loggerControl.Shutdown()

	// 创建IPFS配置
	ipfsConfig := &config.IpfsConfig{
		Enabled:    true,
		ApiAddress: "localhost:5001",
	}

	// 创建IPFS控制器
	ipfsControl, err := NewIpfsControl(ipfsConfig, loggerControl)
	assert.NoError(t, err)

	// 启动IPFS服务
	ipfsControl.StartUp(func(err error) {
		t.Logf("IPFS service failed to start: %v", err)
	})

	// 等待一会儿
	time.Sleep(1 * time.Second)

	// 关闭IPFS服务
	err = ipfsControl.Shutdown()
	assert.NoError(t, err)
}
