package ecc

import (
	"io"
	"log"
	"os"

	ecies "github.com/ecies/go/v2"
)

func EncryptFile(inputFile, outputFile string, publicKey *ecies.PublicKey) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalln("打开待加密文件失败:", err)
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalln("创建加密文件失败:", err)
		return err
	}
	defer outFile.Close()

	blockSize := 4096 - 97 // SM2 加密后的块大小相对固定，可以根据实际情况调整
	buffer := make([]byte, blockSize)

	for {
		n, err := inFile.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatalln("读取待加密文件失败:", err)
			return err
		}
		if n == 0 {
			break
		}

		encryptedBlock, err := ecies.Encrypt(publicKey, buffer[:n])
		if err != nil {
			log.Fatalln("加密文件失败:", err)
			return err
		}

		_, err = outFile.Write(encryptedBlock)
		if err != nil {
			log.Fatalln("写入加密文件失败:", err)
			return err
		}
	}
	return nil
}

func DecryptFile(inputFile, outputFile string, privateKey *ecies.PrivateKey) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalln("打开待解密文件失败:", err)
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalln("创建解密文件失败:", err)
		return err
	}
	defer outFile.Close()

	blockSize := 4096 // SM2 解密时也可以根据实际情况调整块大小
	buffer := make([]byte, blockSize)

	for {
		n, err := inFile.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatalln("读取待解密文件失败:", err)
			return err
		}
		if n == 0 {
			break
		}

		decryptedBlock, err := ecies.Decrypt(privateKey, buffer[:n])
		if err != nil {
			log.Fatalln("解密文件失败:", err)
			return err
		}

		_, err = outFile.Write(decryptedBlock)
		if err != nil {
			log.Fatalln("写入解密文件失败:", err)
			return err
		}
	}
	return nil
}
