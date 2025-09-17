package marshaler

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"_examples/pkg/str"
)

func Marshal(obj interface{}) ([]byte, error) {
	data := make(map[string]interface{})

	val := reflect.ValueOf(obj)
	typeOf := val.Type()

	switch typeOf.Kind() {
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			fieldType := typeOf.Field(i)
			// 判断结构体中的变量是否被导出
			if !str.CompareIgnoreCase("", fieldType.PkgPath) {
				continue
			}

			fieldName := fieldType.Name
			if str.CompareIgnoreCase("uid", fieldName) {
				continue
			}

			field := val.Field(i)

			jsonTag := fieldType.Tag.Get("json")

			if str.CompareIgnoreCase("", jsonTag) {
				jsonTag = strings.ToLower(fieldName[:1]) + fieldName[1:]
			}

			fieldValue := field.Interface()

			jsonName, options := splitJsonTag(jsonTag)

			if containsOmitEmpty(options) && isZero(field) {
				continue
			}

			maskTag := fieldType.Tag.Get("mask")
			if !str.CompareIgnoreCase("", maskTag) {
				switch maskTag {
				case "cid":
					fieldValue = maskCid(fieldValue.(string))
				case "uid":
					fieldValue = maskUid(fieldValue.(string))
				case "email":
					fieldValue = maskEmail(fieldValue.(string))
				case "address":
					fieldValue = maskAddress(fieldValue.(string))
				case "all":
					fallthrough
				default:
					fieldValue = maskAll(fieldValue.(string))
				}
			}
			data[jsonName] = fieldValue
		}
		return json.Marshal(data)
	case reflect.String:
		return []byte(val.String()), nil
	case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32:
		return json.Marshal(val.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return json.Marshal(val.Uint())
	case reflect.Float32, reflect.Float64:
		return json.Marshal(val.Float())
	case reflect.Bool:
		return json.Marshal(val.Bool())
	case reflect.Slice, reflect.Array:
		return json.Marshal(val.Interface())
	case reflect.Map:
		return json.Marshal(val.Interface())
	case reflect.Ptr:
		return json.Marshal(val.Pointer())
	case reflect.Interface:
		return json.Marshal(val.Interface())
	default:
		return nil, fmt.Errorf("unsupported type: %s", typeOf.Kind())
	}
}

func maskAddress(str string) interface{} {
	return mask(str, 1, 1)
}

func maskEmail(str string) string {

	parts := strings.Split(str, "@")
	if len(parts) != 2 {
		return str
	}
	userName := parts[0]
	domain := parts[1]
	return maskUserName(userName) + "@" + maskDomain(domain)
}

func maskDomain(domain string) string {
	parts := strings.Split(domain, ".")
	maskedDomainParts := make([]string, len(parts))

	for i, part := range parts {
		if i == 0 { // 只遮蔽主域名
			if len(part) <= 2 {
				maskedDomainParts[i] = strings.Repeat("*", len(part)) // 短域名全遮蔽
			} else {
				maskedDomainParts[i] = part[:1] + strings.Repeat("*", len(part)-1) // 保留首尾字符 + part[len(part)-1:]
				// maskedDomainParts[i] = part[:1] + strings.Repeat("*", len(part)-2) + part[len(part)-1:]
			}
		} else {
			maskedDomainParts[i] = part // 其他部分保持不变 (例如 com, net)
		}
	}

	return strings.Join(maskedDomainParts, ".")

}

func maskUserName(userName string) string {
	if len(userName) <= 2 {
		return strings.Repeat("*", len(userName))
	}
	// return userName[:1] + strings.Repeat("*", len(userName)-1) + userName[len(userName)-1:]
	return userName[:1] + strings.Repeat("*", len(userName)-1)
}

func maskAll(str string) string {
	return mask(str, 0, 0)
}

func maskUid(str string) string {
	return mask(str, 5, 5)
}

func maskCid(str string) string {
	return mask(str, 5, 5)
}

func mask(str string, pre int, suf int) string {
	if len(str) <= pre+suf {
		return strings.Repeat("*", len(str)-pre-suf)
	}
	return str[:pre] + strings.Repeat("*", (len(str)-pre-suf)) + str[len(str)-suf:]
}

// isZero 判断值是否为空
// @param value: 待判断的值
// @return bool: 是否为空
func isZero(value reflect.Value) bool {
	switch value.Kind() {

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.String:
		return str.CompareIgnoreCase("", value.String())
	case reflect.Ptr, reflect.Interface:
		return value.IsNil()
	case reflect.Slice, reflect.Array:
		return value.Len() == 0
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if !isZero(value.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// containsOmitEmpty 判断是否包含omitempty选项
// @param options: 待判断的选项
// @return bool: 是否包含
func containsOmitEmpty(options []string) bool {
	return slices.Contains(options, "omitempty")
}

// splitJsonTag 切割jsonTag
// @param tag: 待切割的tag内容
// @return jsonName: json字段名
// @return options: json字段选项
func splitJsonTag(tag string) (string, []string) {
	split := strings.Split(tag, ",")
	return split[0], split[1:]

}
