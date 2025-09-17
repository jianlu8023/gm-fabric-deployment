package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// EncryptCBC 使用AES CBC模式加密
func EncryptCBC(key, plaintext string) (string, error) {

	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("无效的密钥: %v", err)
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	mode := cipher.NewCBCEncrypter(block, iv)

	// 将明文转换为字节切片并进行填充
	plaintextBytes := []byte(plaintext)
	paddedPlaintext := PKCS7Padding(plaintextBytes, blockSize)

	// 创建一个缓冲区来存储加密后的数据
	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// 将 IV 和加密后的数据拼接在一起
	encryptedData := append(iv, ciphertext...)

	// 将结果转换为十六进制字符串
	encryptedHex := hex.EncodeToString(encryptedData)

	return encryptedHex, nil
}

// DecryptCBC 使用AES CBC模式解密
func DecryptCBC(key, encryptedHex string) (string, error) {
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		return "", fmt.Errorf("无效的密钥: %v", err)
	}

	encryptedData, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("无效的加密数据: %v", err)
	}

	if len(encryptedData) < blockSize {
		return "", fmt.Errorf("加密数据太短")
	}

	iv := encryptedData[:blockSize]
	ciphertext := encryptedData[blockSize:]

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// 创建一个缓冲区来存储解密后的数据
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 移除填充
	unpaddedPlaintext, err := PKCS7UnPadding(plaintext)
	if err != nil {
		return "", err
	}

	return string(unpaddedPlaintext), nil
}

// EncryptCBCFile 使用AES CBC模式加密文件
func EncryptCBCFile(key, inputFile, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("打开输入文件错误: %v", err)
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("创建输出文件错误: %v", err)
		return err
	}
	defer outFile.Close()
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		log.Fatalf("无效的密钥: %v", err)
		return err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return err
	}

	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	if _, err := outFile.Write(iv); err != nil {
		return err
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	buf := make([]byte, blockSize)
	buffer := bytes.NewBuffer(nil)

	for {
		n, err := inFile.Read(buf)
		if n > 0 {
			buffer.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// 处理最后一部分数据并进行填充
	paddedData := PKCS7Padding(buffer.Bytes(), blockSize)

	// 加密整个数据
	ciphertext := make([]byte, len(paddedData))
	mode.CryptBlocks(ciphertext, paddedData)

	// 写入加密后的数据
	if _, err := outFile.Write(ciphertext); err != nil {
		return err
	}

	fmt.Println("文件加密完成")
	return nil
}

// DecryptCBCFile 使用AES CBC模式解密文件
func DecryptCBCFile(key, inputFile, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("打开输入文件错误: %v", err)
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("创建输出文件错误: %v", err)
		return err
	}
	defer outFile.Close()
	keyBytes, err := hex.DecodeString(key)
	if err != nil {
		log.Fatalf("无效的密钥: %v", err)
		return err
	}
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return err
	}

	iv := make([]byte, blockSize)
	if _, err := io.ReadFull(inFile, iv); err != nil {
		return err
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	buf := make([]byte, blockSize)
	buffer := bytes.NewBuffer(nil)

	for {
		n, err := inFile.Read(buf)
		if n > 0 {
			buffer.Write(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	// 解密整个数据
	ciphertext := buffer.Bytes()
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// 移除填充
	unpaddedBuf, err := PKCS7UnPadding(plaintext)
	if err != nil {
		return err
	}

	// 写入解密后的数据
	if _, err := outFile.Write(unpaddedBuf); err != nil {
		return err
	}

	fmt.Println("文件解密完成")
	return nil
}
