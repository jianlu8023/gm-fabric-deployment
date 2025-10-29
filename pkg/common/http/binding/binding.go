package binding

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// BindFormData 绑定表单数据
//
// @description 从HTTP请求中绑定表单数据到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindFormData(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.Form); err != nil {
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
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
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
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
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
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
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
	}
	return nil
}

// BindXML 绑定XML数据
//
// @description 从HTTP请求体中绑定XML格式的数据到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindXML(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.XML); err != nil {
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
	}
	return nil
}

// BindURI 绑定URI参数
//
// @description 从HTTP请求URI中绑定参数到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindURI(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindUri(body); err != nil {
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
	}
	return nil
}

// BindHeader 绑定Header参数
//
// @description 从HTTP请求Header中绑定参数到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindHeader(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.Header); err != nil {
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
	}
	return nil
}

// BindYAML 绑定YAML数据
//
// @description 从HTTP请求体中绑定YAML格式的数据到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindYAML(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.YAML); err != nil {
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
	}
	return nil
}

// BindTOML 绑定TOML数据
//
// @description 从HTTP请求体中绑定TOML格式的数据到结构体
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindTOML(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBindWith(body, binding.TOML); err != nil {
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
	}
	return nil
}

// BindAuto 根据Content-Type自动选择绑定方式
//
// @description 根据请求的Content-Type头自动选择合适的绑定方式，支持JSON、XML、Form等格式
// @param ctx *gin.Context Gin上下文
// @param body BodyLegal 实现BodyLegal接口的结构体指针
// @return error 绑定过程中发生的错误，成功则返回nil
func BindAuto(ctx *gin.Context, body BodyLegal) error {
	if err := ctx.ShouldBind(body); err != nil {
		// 检查是否是validator.ValidationErrors错误
		if IsValidationErrors(err) {
			return NewValidationErrors(err.(validator.ValidationErrors))
		}

		// 其他错误作为参数验证错误处理
		return NewInvalidValidationError(err)
	}
	return nil
}
