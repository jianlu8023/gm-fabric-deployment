package response

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// SystemInitStatusResponse 系统初始化状态响应
//
// @description 用于返回系统初始化状态的响应数据
// @struct
type SystemInitStatusResponse struct {
	Init bool `json:"init" yaml:"init"` // 初始化状态
}

// String 将SystemInitStatusResponse转换为字符串
//
// @description 将SystemInitStatusResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp SystemInitStatusResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// MarshalJSON 自定义JSON序列化方法
//
// @description 将SystemInitStatusResponse结构体转换为JSON格式
// @return []byte JSON字节数组
// @return error 错误信息
func (resp SystemInitStatusResponse) MarshalJSON() ([]byte, error) {
	type Alias SystemInitStatusResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// NewSystemInitStatusResponse 创建系统初始化状态响应对象
//
// @description 创建一个新的系统初始化状态响应对象
// @param init bool 初始化状态
// @return SystemInitStatusResponse 系统初始化状态响应对象
func NewSystemInitStatusResponse(init bool) SystemInitStatusResponse {
	return SystemInitStatusResponse{
		Init: init,
	}
}
