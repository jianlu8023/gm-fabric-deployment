package des

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// EncryptCBCFile 使用DES CBC模式加密文件
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

	block, err := des.NewCipher(keyBytes)
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
	paddedData := PKCS5Padding(buffer.Bytes(), blockSize)

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

// DecryptCBCFile 使用DES CBC模式解密文件
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

	block, err := des.NewCipher(keyBytes)
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
	unpaddedBuf, err := PKCS5UnPadding(plaintext)
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
