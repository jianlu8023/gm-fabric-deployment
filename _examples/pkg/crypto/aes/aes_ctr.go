package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
)

// EncryptFile 使用AES CTR模式加密文件
func EncryptFile(key, inputFile, outputFile string) error {
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

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	if _, err := outFile.Write(iv); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, iv)
	buf := make([]byte, aes.BlockSize)

	for {
		n, err := inFile.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf[:n], buf[:n])
			if _, err := outFile.Write(buf[:n]); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	fmt.Println("文件加密完成")
	return nil
}

// DecryptFile 使用AES CTR模式解密文件
func DecryptFile(key, inputFile, outputFile string) error {
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

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(inFile, iv); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, iv)
	buf := make([]byte, aes.BlockSize)

	for {
		n, err := inFile.Read(buf)
		if n > 0 {
			stream.XORKeyStream(buf[:n], buf[:n])
			if _, err := outFile.Write(buf[:n]); err != nil {
				return err
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	fmt.Println("文件解密完成")
	return nil
}
