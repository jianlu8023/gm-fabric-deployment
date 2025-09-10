package binding

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// BindBody 绑定参数, 获取 json 数据
func BindBody(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindJSON(body); err != nil {
		return err
	}
	return nil
}

// BindFormData 绑定参数
func BindFormData(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.Form); err != nil {
		return err
	}
	return nil
}

// BindMultiPartForm 绑定参数, 获取 form-data 数据
// @param ctx: gin.Context
// @param body: BodyLegal
// @return error: 错误信息
func BindMultiPartForm(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.FormMultipart); err != nil {
		return err
	}
	return nil
}

// BindQuery 绑定参数, 获取 query 数据
// @param ctx: gin.Context
// @param body: BodyLegal
// @return error: 错误信息
func BindQuery(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.Query); err != nil {
		return err
	}
	return nil
}
