// Package {{ .PackageName }}
//
//
package {{ .PackageName }}


// Param 请求参数
type Param struct {
}

// String 返回json字符串
// @return string: json字符串
func (p Param) String() string {
	if jsonStr, err := json.ToJSON(p); err != nil {
		return ""
	} else {
		return jsonStr
	}
}
