package binding

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// BindFormData 绑定表单数据
//
// @description 从HTTP请求中绑定表单数据到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindFormData(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.Form); err != nil {
		return err
	}
	return nil
}

// BindMultiPartForm 绑定多部分表单数据
//
// @description 从HTTP请求中绑定multipart/form-data类型的数据到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindMultiPartForm(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.FormMultipart); err != nil {
		return err
	}
	return nil
}

// BindQuery 绑定查询参数
//
// @description 从HTTP请求URL中绑定查询参数到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindQuery(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.Query); err != nil {
		return err
	}
	return nil
}

// BindJSON 绑定JSON数据
//
// @description 从HTTP请求体中绑定JSON格式的数据到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindJSON(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.JSON); err != nil {
		return err
	}
	return nil

}
