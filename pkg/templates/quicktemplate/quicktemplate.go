package quicktemplate

import (
	"fmt"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/templates/quicktemplate/templates"
)

// go install github.com/valyala/quicktemplate/qtc@latest
//go:generate qtc -dir=pkg/template/quicktemplate/templates

func QuickTemplateExample() {
	// 整型：{%d int %}，{%dl int64 %}，{%dul uint64 %}；
	// 浮点数：{%f float %}。还可以设置输出的精度，使用{%f.precision float %}。例如{%f.2 1.2345 %}输出1.23；
	// 字节切片（[]byte）：{%z bytes %}；
	// 字符串：{%q str %}或字节切片：{%qz bytes %}，引号转义为&quot;；
	// 字符串：{%j str %}或字节切片：{%jz bytes %}，没有引号；
	// URL 编码：{%u str %}，{%uz bytes %}；
	// {%v anything %}：输出等同于fmt.Sprintf("%v", anything)

	fmt.Println(templates.Types(1, 5.75, []byte{'a', 'b', 'c'}, "hello"))
}
