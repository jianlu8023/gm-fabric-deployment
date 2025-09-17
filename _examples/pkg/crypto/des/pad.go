package des

import (
	"bytes"
	"crypto/des"
	"fmt"
)

const blockSize = des.BlockSize // DES 块大小为 8 字节

// PKCS5Padding 添加 PKCS#5 填充
func PKCS5Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// PKCS5UnPadding 移除 PKCS#5 填充
func PKCS5UnPadding(src []byte) ([]byte, error) {
	length := len(src)
	if length == 0 {
		return nil, fmt.Errorf("输入为空")
	}
	unpadding := int(src[length-1])
	if unpadding > length || unpadding > blockSize {
		return nil, fmt.Errorf("无效的填充")
	}
	return src[:length-unpadding], nil
}
