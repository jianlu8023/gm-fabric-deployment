// Package response 提供Web模块接口响应数据定义
//
// @Description 包括用户、验证码、系统、节点、Docker镜像、Docker网络、文件、gRPC、MFA、WebRTC等
// 各业务模块的响应数据结构体及其自定义JSON序列化方法。
package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// Response 基础响应结构体
//
// @description 所有响应数据结构体的基础结构体，提供统一的JSON序列化和字符串表示方法
// @struct
type Response struct{}

// MarshalJSON 自定义JSON序列化方法
//
// @description 通过别名机制避免递归调用，实现Response结构体的JSON序列化
// @return []byte JSON字节数组
// @return error 序列化错误信息
func (resp Response) MarshalJSON() ([]byte, error) {
	type Alias Response
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// GoString 实现GoStringer接口
//
// @description 返回响应数据的字符串表示，用于%#v格式化输出
// @return string 响应数据的字符串表示
func (resp Response) GoString() string {
	return resp.String()
}

// String 将响应数据转换为字符串表示
//
// @description 将Response结构体转换为JSON格式的字符串
// @return string 响应数据的JSON格式字符串
func (resp Response) String() string {
	marshalString, _ := json.MarshalString(resp)
	return marshalString
}

// NewResponse 创建基础响应对象
//
// @description 创建并返回一个新的基础Response实例
// @return Response 基础响应对象
func NewResponse() Response {
	return Response{}
}
