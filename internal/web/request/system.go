package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// SystemOverviewRequest 系统概览请求结构体
//
// @description 用于获取系统概览信息
// @struct
type SystemOverviewRequest struct {
}

// String 将系统概览请求参数转换为字符串表示
//
// @description 将SystemOverviewRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req SystemOverviewRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证系统概览请求参数是否合法
//
// @description 由于无参数，直接返回true
// @return bool 参数是否合法
func (req SystemOverviewRequest) IsLegal() bool {
	return true
}

// SystemInitStatusRequest 系统初始化状态请求结构体
//
// @description 用于获取系统初始化状态信息
// @struct
type SystemInitStatusRequest struct {
}

// String 将系统初始化状态请求参数转换为字符串表示
//
// @description 将SystemInitStatusRequest结构体转换为JSON格式的字符串
// @return string 请求参数的JSON格式字符串
func (req SystemInitStatusRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证系统初始化状态请求参数是否合法
//
// @description 由于无参数，直接返回true
// @return bool 参数是否合法
func (req SystemInitStatusRequest) IsLegal() bool {
	return true
}

