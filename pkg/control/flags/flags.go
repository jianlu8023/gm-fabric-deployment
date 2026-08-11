package flags

import (
	"fmt"
	"github.com/jianlu8023/go-tools/v2/pkg/colour"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
)

var _defaultConfig = "configs/default.yaml"

// Flags 存储所有命令行参数的结构体
// @description 用于统一管理和访问所有命令行参数
// @struct Flags
// @field ConfigPath 配置文件路径
// @field ConfigType 配置文件类型(dev, prod等)
// @field Debug 是否启用调试模式
// @field Version 是否显示版本信息
// @field OtherFlags 存储其他自定义命令行参数
// @field Parsed 是否已解析命令行参数
// @field parser go-flags解析器
type Flags struct {
	ConfigPath string            `long:"config" short:"c" description:"配置文件路径: configs/default.yaml" default:"configs/default.yaml"`
	ConfigType string            `long:"type" short:"t" description:"配置文件类型: dev 或 prod" default:""`
	Debug      bool              `long:"debug" short:"d" description:"启用调试模式"`
	Version    bool              `long:"version" short:"v" description:"显示版本信息"`
	OtherFlags map[string]string // 存储其他自定义命令行参数
	Parsed     bool              // 是否已解析命令行参数
	parser     *flags.Parser     // go-flags解析器
}

// newFlags 创建一个新的Flags实例
// @return *Flags 新创建的Flags实例
// @description 初始化Flags结构体并设置默认值
func newFlags() *Flags {
	// fmt.Printf("starting new Flags...\n")
	f := &Flags{
		ConfigPath: _defaultConfig, // 默认配置文件路径
		ConfigType: "",             // 默认不指定配置类型
		Debug:      false,          // 默认不启用调试模式
		Version:    false,          // 默认不显示版本信息
		OtherFlags: make(map[string]string),
		Parsed:     false,
	}

	// 创建go-flags解析器，添加IgnoreUnknown选项以忽略未知标志
	f.parser = flags.NewParser(f, flags.Default|flags.IgnoreUnknown|flags.HelpFlag)

	return f
}

// registerStandardFlags 注册标准命令行参数
// @description 为Flags实例注册内置的标准命令行参数
// 注意：go-flags使用结构体标签定义flag，此函数保留以保持API兼容性
func (f *Flags) registerStandardFlags() {
	// 无需实现，go-flags使用结构体标签自动注册flag
}

// Parse 解析命令行参数
// @param args 命令行参数数组，通常是os.Args[1:]
// @return error 解析过程中的错误
// @description 解析命令行参数并设置到Flags结构体中
func (f *Flags) Parse(args []string) error {
	// fmt.Printf("starting parse flags...\n")
	// 使用go-flags解析命令行参数
	remainingArgs, err := f.parser.ParseArgs(args)
	// 判断是否是用户请求输出help
	if err != nil {
		// 检查是否是帮助请求
		if flags.WroteHelp(err) {
			// 用户请求了帮助信息(-h或--help)，这不是真正的错误
			// 直接退出程序，不继续执行
			os.Exit(0)
		}
		return fmt.Errorf("解析命令行参数失败: %w", err)
	}

	f.Parsed = true

	// Debug模式下记录未知的flag
	if f.Debug && len(remainingArgs) > 0 {
		fmt.Printf("%v  [%v]  flags/glags.go:85  [flags/control]  warning: unknown flags (ignored by IgnoreUnknown): %v\n", time.Now().Format("2006-01-02 15:04:05.000"), colour.Yellow("WARN "), remainingArgs)
	}

	// 处理剩余的非flag参数
	// 支持 --key=value 和 --key value 两种格式
	for i := 0; i < len(remainingArgs); i++ {
		arg := remainingArgs[i]

		// 只处理以 -- 开头的参数
		if !strings.HasPrefix(arg, "--") {
			continue
		}

		key, value := "", ""
		consumedNext := false
		rest := arg[2:] // 去掉 --

		if idx := strings.Index(rest, "="); idx >= 0 {
			// --key=value 格式
			key = rest[:idx]
			value = rest[idx+1:]
		} else {
			// --key value 格式
			key = rest
			// 检查下一个参数是否是值（不是以 - 开头）
			if i+1 < len(remainingArgs) && !strings.HasPrefix(remainingArgs[i+1], "-") {
				value = remainingArgs[i+1]
				consumedNext = true
			}
		}

		// 验证key的合法性
		if key != "" && isValidFlagKey(key) {
			f.OtherFlags[key] = value
		}
		if consumedNext {
			i++ // 跳过已处理的下一个参数
		}
	}

	// 验证已知参数的合法性
	if err := f.validate(); err != nil {
		return err
	}

	return nil
}

// validate 验证命令行参数的合法性
// @return error 验证过程中的错误
// @description 验证命令行参数的合法性
func (f *Flags) validate() error {
	// ConfigType 允许任意值，由业务逻辑自行处理
	return nil
}

// isValidFlagKey 验证flag key的合法性
// @param key flag的key
// @return bool key是否合法
// @description 验证flag key的合法性，不允许空字符串和包含空格的key
func isValidFlagKey(key string) bool {
	return key != "" && !strings.Contains(key, " ")
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
	builder.WriteString("}")
	return builder.String()
}
