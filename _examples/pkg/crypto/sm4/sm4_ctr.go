package sm4

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tjfoc/gmsm/sm4"
)

// EncryptFile 使用SM4 CTR模式加密文件
func EncryptFile(key, inputFile, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("open input file error: %v", err)
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("create output file error: %v", err)
		return err
	}
	defer outFile.Close()

	block, err := sm4.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	if _, err := outFile.Write(iv); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, iv)
	writer := &cipher.StreamWriter{S: stream, W: outFile}

	if _, err := io.Copy(writer, inFile); err != nil {
		return err
	}

	fmt.Println("文件加密完成")
	return nil
}

// DecryptFile 使用SM4 CTR模式解密文件
func DecryptFile(key, inputFile, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("open input file error: %v", err)
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("create output file error: %v", err)
		return err
	}
	defer outFile.Close()

	block, err := sm4.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(inFile, iv); err != nil {
		return err
	}

	stream := cipher.NewCTR(block, iv)
	reader := &cipher.StreamReader{S: stream, R: inFile}

	if _, err := io.Copy(outFile, reader); err != nil {
		return err
	}

	fmt.Println("文件解密完成")
	return nil
}
