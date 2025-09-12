package flags

import (
	"flag"
	"fmt"
	"strings"
)

var _defaultConfig = "configs/default.yaml"

// Flags 存储所有命令行参数的结构体
// @description 用于统一管理和访问所有命令行参数
// @struct Flags
// @field ConfigPath 配置文件路径
// @field ConfigType 配置文件类型(dev, prod等)
// @field LogLevel 日志级别
// @field LogFile 日志文件路径
// @field Debug 是否启用调试模式
// @field Version 是否显示版本信息
// @field OtherFlags 存储其他自定义命令行参数
// @field Parsed 是否已解析命令行参数
// @field mutex 用于保证并发安全的互斥锁
// @field flagSet 自定义的flag集合，避免使用全局flag

type Flags struct {
	ConfigPath string // 配置文件路径
	ConfigType string // 配置文件类型(dev, prod等)
	// LogLevel    string            // 日志级别
	// LogFile     string            // 日志文件路径
	Debug      bool              // 是否启用调试模式
	Version    bool              // 是否显示版本信息
	OtherFlags map[string]string // 存储其他自定义命令行参数
	Parsed     bool              // 是否已解析命令行参数
	flagSet    *flag.FlagSet     // 自定义的flag集合，避免使用全局flag
}

// newFlags 创建一个新的Flags实例
// @return *Flags 新创建的Flags实例
// @description 初始化Flags结构体并设置默认值
func newFlags() *Flags {
	// fmt.Printf("starting new Flags...\n")
	flagSet := flag.NewFlagSet("gm-fabric-deployment", flag.ExitOnError)

	f := &Flags{
		ConfigPath: _defaultConfig, // 默认配置文件路径
		ConfigType: "",             // 默认不指定配置类型
		// LogLevel:   "info",                 // 默认日志级别为info
		// LogFile:    "./logs/app.log",       // 默认日志文件路径
		Debug:      false, // 默认不启用调试模式
		Version:    false, // 默认不显示版本信息
		OtherFlags: make(map[string]string),
		Parsed:     false,
		flagSet:    flagSet,
	}

	// 注册标准命令行参数
	// fmt.Printf("starting register standard flags...\n")
	f.registerStandardFlags()

	return f
}

// registerStandardFlags 注册标准命令行参数到flagSet
// @description 为Flags实例注册内置的标准命令行参数
func (f *Flags) registerStandardFlags() {
	f.flagSet.StringVar(&f.ConfigPath, "config", f.ConfigPath, "配置文件路径: configs/default.yaml")
	f.flagSet.StringVar(&f.ConfigType, "type", f.ConfigType, "配置文件类型: dev 或 prod")
	// f.flagSet.StringVar(&f.LogLevel, "loglevel", f.LogLevel, "日志级别: debug, info, warn, error, fatal")
	// f.flagSet.StringVar(&f.LogFile, "logfile", f.LogFile, "日志文件路径: ./logs/app.log")
	f.flagSet.BoolVar(&f.Debug, "debug", f.Debug, "启用调试模式")
	f.flagSet.BoolVar(&f.Version, "version", f.Version, "显示版本信息")
}

// Parse 解析命令行参数
// @param args 命令行参数数组，通常是os.Args[1:]
// @return error 解析过程中的错误
// @description 解析命令行参数并设置到Flags结构体中
func (f *Flags) Parse(args []string) error {
	// fmt.Printf("starting parse flags...\n")
	if err := f.flagSet.Parse(args); err != nil {
		return fmt.Errorf("解析命令行参数失败: %w", err)
	}

	f.Parsed = true

	// 处理非flag参数
	remainingArgs := f.flagSet.Args()
	for i := 0; i < len(remainingArgs); i++ {
		arg := remainingArgs[i]
		// 处理格式为--key=value的参数
		if strings.HasPrefix(arg, "--") {
			parts := strings.SplitN(arg[2:], "=", 2)
			if len(parts) == 2 {
				f.OtherFlags[parts[0]] = parts[1]
			} else {
				// 如果没有值，设置为空字符串
				f.OtherFlags[parts[0]] = ""
			}
		}
	}
	return nil
}

// GetString 获取字符串类型的命令行参数
// @param name 参数名称
// @param defaultValue 默认值
// @return string 参数值，如果不存在则返回默认值
// @description 安全地获取字符串类型的命令行参数
func (f *Flags) GetString(name string, defaultValue string) string {
	if val, exists := f.OtherFlags[name]; exists {
		return val
	}

	// 检查标准参数
	switch name {
	case "config":
		return f.ConfigPath
	case "type":
		return f.ConfigType
	// case "loglevel":
	// 	return f.LogLevel
	// case "logfile":
	// 	return f.LogFile
	default:
		return defaultValue
	}
}

// GetBool 获取布尔类型的命令行参数
// @param name 参数名称
// @param defaultValue 默认值
// @return bool 参数值，如果不存在则返回默认值
// @description 安全地获取布尔类型的命令行参数
func (f *Flags) GetBool(name string, defaultValue bool) bool {
	if val, exists := f.OtherFlags[name]; exists {
		// 将字符串转换为布尔值
		return strings.ToLower(val) == "true" || val == "1" || strings.ToLower(val) == "yes"
	}

	// 检查标准参数
	switch name {
	case "debug":
		return f.Debug
	case "version":
		return f.Version
	default:
		return defaultValue
	}
}

// IsParsed 检查命令行参数是否已解析
// @return bool 是否已解析
// @description 检查命令行参数是否已经被解析过
func (f *Flags) IsParsed() bool {
	return f.Parsed
}

// GetAllFlags 获取所有命令行参数
// @return map[string]interface{} 包含所有参数的映射
// @description 获取所有已解析的命令行参数，包括标准参数和自定义参数
func (f *Flags) GetAllFlags() map[string]interface{} {
	flagsMap := make(map[string]interface{})

	// 添加标准参数
	flagsMap["config"] = f.ConfigPath
	flagsMap["type"] = f.ConfigType
	// flagsMap["loglevel"] = f.LogLevel
	// flagsMap["logfile"] = f.LogFile
	flagsMap["debug"] = f.Debug
	flagsMap["version"] = f.Version

	// 添加自定义参数
	for key, value := range f.OtherFlags {
		flagsMap[key] = value
	}

	return flagsMap
}

// String 返回Flags的字符串表示
// @return string Flags的字符串表示
// @description 返回Flags结构体的可读字符串表示，用于调试和日志记录
func (f *Flags) String() string {
	var builder strings.Builder
	builder.WriteString("Flags{")
	builder.WriteString(fmt.Sprintf("ConfigPath='%s', ", f.ConfigPath))
	builder.WriteString(fmt.Sprintf("ConfigType='%s', ", f.ConfigType))
	// builder.WriteString(fmt.Sprintf("LogLevel='%s', ", f.LogLevel))
	// builder.WriteString(fmt.Sprintf("LogFile='%s', ", f.LogFile))
	builder.WriteString(fmt.Sprintf("Debug=%v, ", f.Debug))
	builder.WriteString(fmt.Sprintf("Version=%v, ", f.Version))
	builder.WriteString("OtherFlags={")
	first := true
	for k, v := range f.OtherFlags {
		if !first {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%s='%s'", k, v))
		first = false
	}
	builder.WriteString("}")
	builder.WriteString(", Parsed=")
	builder.WriteString(fmt.Sprintf("%v", f.Parsed))
	builder.WriteString(")")
	return builder.String()
}
