package des

import (
	"crypto/cipher"
	"crypto/des"
	"io"
	"os"
)

func EncryptFile(key, inputFile, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	iv := make([]byte, des.BlockSize)
	stream := cipher.NewCTR(block, iv)

	writer := &cipher.StreamWriter{S: stream, W: outFile}

	_, err = io.Copy(writer, inFile)
	return err
}

func DecryptFile(key, inputFile string, outputFile string) error {
	inFile, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	block, err := des.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	iv := make([]byte, des.BlockSize)
	stream := cipher.NewCTR(block, iv)

	reader := &cipher.StreamReader{S: stream, R: inFile}

	_, err = io.Copy(outFile, reader)
	return err
}
