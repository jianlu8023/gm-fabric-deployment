package aes

import (
	"bytes"
	"crypto/aes"
	"fmt"
)

const blockSize = aes.BlockSize // AES 块大小为 16 字节

// PKCS7Padding 添加 PKCS#7 填充
func PKCS7Padding(src []byte, blockSize int) []byte {
	padding := blockSize - len(src)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(src, padtext...)
}

// PKCS7UnPadding 移除 PKCS#7 填充
func PKCS7UnPadding(src []byte) ([]byte, error) {
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
