package binding

// BodyLegal 请求体合法性接口
//
// @description 定义所有请求体需要实现的合法性验证接口
// @interface
type BodyLegal interface {
	// IsLegal 检查请求体是否合法
	//
	// @return bool 请求体是否合法
	IsLegal() bool
}
