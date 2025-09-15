package binding

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"strings"
)

// InvalidValidationError 定义参数验证错误类型
//
// @description 当参数格式错误、类型不匹配等情况时返回此错误
// @struct
type InvalidValidationError struct {
	Err error
}

// NewInvalidValidationError 创建一个新的InvalidValidationError实例
//
// @description 根据原始错误创建一个InvalidValidationError实例
// @param err error 原始错误
// @return *InvalidValidationError 参数验证错误实例
func NewInvalidValidationError(err error) *InvalidValidationError {
	return &InvalidValidationError{
		Err: err,
	}
}

// Error 实现error接口
//
// @return string 错误信息
func (e *InvalidValidationError) Error() string {
	return "参数验证错误: " + e.Err.Error()
}

// Unwrap 实现errors.Unwrap接口
//
// @return error 被包装的原始错误
func (e *InvalidValidationError) Unwrap() error {
	return e.Err
}

// ValidationErrors 定义字段校验错误类型
//
// @description 当字段值不符合验证规则时返回此错误
// @struct
type ValidationErrors struct {
	Errors validator.ValidationErrors
}

// NewValidationErrors 创建一个新的ValidationErrors实例
//
// @description 根据validator/v10的ValidationErrors创建一个ValidationErrors实例
// @param errs validator.ValidationErrors 字段验证错误集合
// @return *ValidationErrors 字段校验错误实例
func NewValidationErrors(errs validator.ValidationErrors) *ValidationErrors {
	return &ValidationErrors{
		Errors: errs,
	}
}

// Error 实现error接口
//
// @return string 错误信息
func (e *ValidationErrors) Error() string {
	if e.Errors == nil {
		return "校验错误"
	}
	return e.Errors.Error()
}

// IsInvalidValidationError 判断错误是否是InvalidValidationError类型
//
// @description 检查给定的错误是否是InvalidValidationError类型的错误
// @param err error 需要检查的错误
// @return bool 如果是InvalidValidationError错误则返回true，否则返回false
//
//nolint:unused
func IsInvalidValidationError(err error) bool {
	var invalidErr *InvalidValidationError
	return errors.As(err, &invalidErr)
}

// IsValidationErrors 判断错误是否是ValidationErrors类型
//
// @description 检查给定的错误是否是ValidationErrors类型的错误
// @param err error 需要检查的错误
// @return bool 如果是ValidationErrors错误则返回true，否则返回false
func IsValidationErrors(err error) bool {
	var validationErrs *ValidationErrors
	if errors.As(err, &validationErrs) {
		return true
	}
	// 也检查原始的validator/v10.ValidationErrors类型
	var validatorErrs validator.ValidationErrors
	return errors.As(err, &validatorErrs)
}

// GetValidationErrorMessages 从验证错误中提取错误信息
//
// @description 从ValidationErrors或原始的validator/v10.ValidationErrors中提取人类可读的错误信息
// @param err error 验证错误
// @return map[string]string 字段名到错误信息的映射，格式为"【字段】违反什么策略"
func GetValidationErrorMessages(err error) map[string]string {
	messages := make(map[string]string)

	// 检查是否是ValidationErrors类型
	var validationErrs *ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, fieldErr := range validationErrs.Errors {
			// 获取字段名或form标签
			fieldName := fieldErr.Field()
			// 获取验证标签
			tag := fieldErr.Tag()
			// 构建格式为"【字段】违反什么策略"的错误消息
			message := "【" + fieldName + "】违反" + getValidationRuleDescription(tag)
			messages[fieldName] = message
		}
		return messages
	}

	// 检查是否是原始的validator/v10.ValidationErrors类型
	var validatorErrs validator.ValidationErrors
	if errors.As(err, &validatorErrs) {
		for _, fieldErr := range validatorErrs {
			// 获取字段名或form标签
			fieldName := fieldErr.Field()
			// 获取验证标签
			tag := fieldErr.Tag()
			// 构建格式为"【字段】违反什么策略"的错误消息
			message := "【" + fieldName + "】违反" + getValidationRuleDescription(tag)
			messages[fieldName] = message
		}
		return messages
	}

	return messages
}

// getValidationRuleDescription 将验证标签转换为人类可读的描述
//
// @description 将validator/v10的验证标签转换为易于理解的中文描述
// @param tag string 验证标签
// @return string 验证规则的中文描述
func getValidationRuleDescription(tag string) string {
	switch tag {
	case "required":
		return "必填策略"
	case "email":
		return "邮箱格式策略"
	case "url":
		return "URL格式策略"
	case "min":
		return "最小值策略"
	case "max":
		return "最大值策略"
	case "len":
		return "长度策略"
	case "eqfield":
		return "字段相等策略"
	case "nefield":
		return "字段不相等策略"
	case "oneof":
		return "枚举值策略"
	case "numeric":
		return "数值格式策略"
	case "alpha":
		return "字母格式策略"
	case "alphanum":
		return "字母数字格式策略"
	case "ip":
		return "IP地址格式策略"
	case "uuid":
		return "UUID格式策略"
	case "datetime":
		return "日期时间格式策略"
	default:
		return tag + "验证策略"
	}
}

// ValidatorError 翻译表单参数验证器出现的校验错误
func ValidatorError(c *gin.Context, err error) {
	if _, ok := err.(validator.ValidationErrors); ok {
		// wrongParam := validator_translation.RemoveTopStruct(errs.Translate(validator_translation.Trans))
		// ReturnJson(c, http.StatusBadRequest, 2001, " consts.ValidatorParamsCheckFailMsg", nil)
	} else {
		errStr := err.Error()
		// multipart:nextpart:eof 错误表示验证器需要一些参数，但是调用者没有提交任何参数
		if strings.ReplaceAll(strings.ToLower(errStr), " ", "") == "multipart:nextpart:eof" {
			// ReturnJson(c, http.StatusBadRequest, 2002, "consts.ValidatorParamsCheckFailMsg", gin.H{"tips": "my_errors.ErrorNotAllParamsIsBlank"})
		} else {
			// ReturnJson(c, http.StatusBadRequest, 2003, "consts.ValidatorParamsCheckFailMsg", gin.H{"tips": errStr})
		}
	}
	c.Abort()
}
