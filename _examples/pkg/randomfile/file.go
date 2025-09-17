package randomfile

import (
	"fmt"
	"math/rand/v2"
	"os"

	"github.com/jianlu8023/go-tools/pkg/path"
)

func RandomFile(filePath string) bool {
	if err := path.Ensure(filePath, false); err != nil {
		fmt.Println("ensure file error:", err)
		return false
	}
	fileSize := 2097152
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open file error:", err)
		return false
	}

	for i := 0; i < fileSize; i++ {
		if _, err := file.Write([]byte{byte(rand.Int64N(256))}); err != nil {
			fmt.Println("write file error:", err)
			return false
		}
	}

	return true
}
