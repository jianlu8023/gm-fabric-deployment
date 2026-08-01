// Package request 提供Web模块接口请求参数定义
//
// @Description 包括用户、验证码、系统、节点、Docker镜像、Docker网络、文件、gRPC、MFA、WebRTC、WebSocket等
// 各业务模块的请求参数结构体及其参数验证方法。
package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// Request 基础请求结构体
//
// @description 所有请求参数结构体的基础结构体，提供统一的合法性验证和字符串表示方法
// @struct
type Request struct {
}

// IsLegal 验证请求参数是否合法
//
// @description 基础实现直接返回true，由具体请求结构体重写以实现各自的参数验证逻辑
// @return bool 参数是否合法
func (req Request) IsLegal() bool {
	return true
}

// GoString 实现GoStringer接口
//
// @description 返回请求参数的字符串表示，用于%#v格式化输出
// @return string 请求参数的字符串表示
func (req Request) GoString() string {
	return req.String()
}

// String 将请求参数转换为字符串表示
//
// @description 将Request结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req Request) String() string {
	marshalString, _ := json.MarshalString(req)
	return marshalString
}
