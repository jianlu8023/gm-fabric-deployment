package flags

import (
	"fmt"
	"os"
	"sync"
)

// Control 命令行参数控制器
// @description 管理命令行参数的解析、获取和生命周期
// @struct Control
// @field flags 命令行参数实例
// @field once 确保参数只被解析一次的同步原语
// @field mutex 保证并发安全的互斥锁
// @field version 应用程序版本号
type Control struct {
	flags   *Flags
	once    sync.Once
	mutex   sync.RWMutex
	version string
}

// NewFlagsControl 创建一个新的FlagsControl实例
// @param version 应用程序版本号
// @return *Control 新创建的FlagsControl实例
// @description 初始化FlagsControl结构体并设置版本号
func NewFlagsControl(version string) *Control {
	fmt.Printf("starting new flags control...\n")
	return &Control{
		flags:   newFlags(),
		version: version,
	}
}

// StartUp 启动服务
// @param failedFunc 解析失败时的回调函数
func (c *Control) StartUp(failedFunc func(err error)) {
	c.once.Do(func() {
		fmt.Printf("starting up flags server...\n")
		c.mutex.Lock()
		parseErr := c.flags.Parse(os.Args[1:])
		if parseErr != nil {
			c.PrintHelp()
			// fmt.Printf("parse error: %v\n", parseErr)
			failedFunc(parseErr)
			return
		}
		// fmt.Printf("flags parsed successfully...\n")
		c.mutex.Unlock()

		if c.IsDebugMode() {
			fmt.Printf("flags: %s\n", c.flags.String())
		}
		// fmt.Printf("flags control StartUp method finished...\n")
	})
}

// Shutdown 关闭命令行参数控制器
// @return error 关闭过程中的错误
// @description 关闭命令行参数控制器（目前为无操作）
func (c *Control) Shutdown() error {
	// 目前没有需要关闭的资源
	// 这里可以添加清理操作，如关闭文件等
	fmt.Printf("shutting down flags server...\n")
	return nil
}

// GetConfigPath 获取配置文件路径
// @return string 配置文件路径
// @description 获取命令行参数中指定的配置文件路径
func (c *Control) GetConfigPath() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.ConfigPath
}

// GetConfigType 获取配置文件类型
// @return string 配置文件类型
// @description 获取命令行参数中指定的配置文件类型
func (c *Control) GetConfigType() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.ConfigType
}

// GetLogLevel 获取日志级别
// @return string 日志级别
// @description 获取命令行参数中指定的日志级别
// func (c *Control) GetLogLevel() string {
// 	c.mutex.RLock()
// 	defer c.mutex.RUnlock()
// 	return c.flags.LogLevel
// }

// GetLogFile 获取日志文件路径
// @return string 日志文件路径
// @description 获取命令行参数中指定的日志文件路径
// func (c *Control) GetLogFile() string {
// 	c.mutex.RLock()
// 	defer c.mutex.RUnlock()
// 	return c.flags.LogFile
// }

// IsDebugMode 检查是否启用调试模式
// @return bool 是否启用调试模式
// @description 检查命令行参数是否指定了启用调试模式
func (c *Control) IsDebugMode() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.Debug
}

// IsVersionRequested 检查是否请求显示版本信息
// @return bool 是否请求显示版本信息
// @description 检查命令行参数是否指定了显示版本信息
func (c *Control) IsVersionRequested() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.Version
}

// GetVersion 获取应用程序版本号
// @return string 应用程序版本号
// @description 获取当前应用程序的版本号
func (c *Control) GetVersion() string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.version
}

// GetString 获取字符串类型的命令行参数
// @param name 参数名称
// @param defaultValue 默认值
// @return string 参数值，如果不存在则返回默认值
// @description 安全地获取字符串类型的命令行参数
func (c *Control) GetString(name string, defaultValue string) string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.GetString(name, defaultValue)
}

// GetBool 获取布尔类型的命令行参数
// @param name 参数名称
// @param defaultValue 默认值
// @return bool 参数值，如果不存在则返回默认值
// @description 安全地获取布尔类型的命令行参数
func (c *Control) GetBool(name string, defaultValue bool) bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.GetBool(name, defaultValue)
}

// GetFlags 获取底层的Flags实例
// @return *Flags 底层的Flags实例
// @description 获取底层的Flags实例，用于高级操作
func (c *Control) GetFlags() *Flags {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags
}

// PrintVersion 打印版本信息
// @description 打印应用程序的版本信息到标准输出
func (c *Control) PrintVersion() {
	fmt.Printf("版本: %s\n", c.version)
}

// PrintHelp 打印帮助信息
// @description 打印命令行参数的帮助信息到标准输出
func (c *Control) PrintHelp() {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	fmt.Println("使用方法:")
	c.flags.parser.WriteHelp(os.Stdout)
}

// IsParsed 检查命令行参数是否已解析
// @return bool 是否已解析
// @description 检查命令行参数是否已经被解析过
func (c *Control) IsParsed() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.IsParsed()
}

// GetAllFlags 获取所有命令行参数
// @return map[string]interface{} 包含所有参数的映射
// @description 获取所有已解析的命令行参数，包括标准参数和自定义参数
func (c *Control) GetAllFlags() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.flags.GetAllFlags()
}
