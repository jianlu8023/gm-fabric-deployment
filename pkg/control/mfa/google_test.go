package mfa

import (
	"fmt"
	"testing"
)

func TestHand(t *testing.T) {
	// 快速生成TestCase文件
	// 需要go get golang.org/x/tools/imports go get github.com/cweill/gotests支持
	//
	// 操作步骤:
	//
	// 光标移动到Method/Function上
	// Command/Control+Shift+T
	g := &googleAuth{
		Secret:       "HXDMVJECJJWSRB3HWIZR4IFUGFTMXBOZ",
		Digits:       6,
		ExpireSecond: 30,
	}
	qr := g.Qr("example.org", "example")
	fmt.Printf("%v\n", qr)
}
