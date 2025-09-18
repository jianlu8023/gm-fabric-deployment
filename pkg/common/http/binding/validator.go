package binding

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"reflect"
)

// https://segmentfault.com/a/1190000023725115 validator 使用 还可以加上i18n

// InitValidators 初始化自定义验证器
// 注册所有自定义的验证标签，确保在应用启动时被调用
func InitValidators() {
	// 获取gin默认使用的验证器引擎
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册pageNo验证器
		v.RegisterValidation("pageNo", validatePageNo)
		// 注册pageSize验证器
		v.RegisterValidation("pageSize", validatePageSize)
	}
}

// validatePageNo 验证页码是否合法
// 页码必须大于0
func validatePageNo(fl validator.FieldLevel) bool {
	// 获取字段值
	field := fl.Field()

	// 根据字段类型处理
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() > 0
	case reflect.Float32, reflect.Float64:
		return field.Float() > 0
	case reflect.String:
		// 如果是字符串类型，可能需要解析为数字
		// 这里简单处理，如果字符串不为空就认为有效
		// 实际应用中可能需要更严格的验证
		return field.String() != ""
	default:
		// 不支持的类型，返回false
		return false
	}
}

// validatePageSize 验证每页大小是否合法
// 每页大小必须大于0
func validatePageSize(fl validator.FieldLevel) bool {
	// 获取字段值
	field := fl.Field()

	// 根据字段类型处理
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() > 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() > 0
	case reflect.Float32, reflect.Float64:
		return field.Float() > 0
	case reflect.String:
		// 如果是字符串类型，可能需要解析为数字
		// 这里简单处理，如果字符串不为空就认为有效
		// 实际应用中可能需要更严格的验证
		return field.String() != ""
	default:
		// 不支持的类型，返回false
		return false
	}
}
