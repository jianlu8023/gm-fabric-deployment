package main

import (
	"crypto/tls"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// corsMiddleware 跨域中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

const (
	uploadDir = "./uploads" // 上传文件存储目录
)

func upload(c *gin.Context) {
	file, err := c.FormFile("chunk")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filename := c.PostForm("filename")
	chunkIndexStr := c.PostForm("chunkIndex")
	totalChunksStr := c.PostForm("totalChunks")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunkIndex"})
		return
	}

	totalChunks, err := strconv.Atoi(totalChunksStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid totalChunks"})
		return
	}

	// 创建分片保存路径
	chunkDir := filepath.Join(uploadDir, filename+"_chunks")
	if _, err := os.Stat(chunkDir); os.IsNotExist(err) {
		os.MkdirAll(chunkDir, 0755)
	}

	// 保存分片
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", chunkIndex))
	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Uploaded chunk %d of %d for file %s\n", chunkIndex+1, totalChunks, filename)

	c.JSON(http.StatusOK, gin.H{"message": "Chunk uploaded successfully"})
}

// checkChunkHandler 检查分片是否存在
func check(c *gin.Context) {
	filename := c.Query("filename")
	chunkIndexStr := c.Query("chunkIndex")

	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid chunkIndex"})
		return
	}

	chunkDir := filepath.Join(uploadDir, filename+"_chunks")
	chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", chunkIndex))

	if _, err := os.Stat(chunkPath); os.IsNotExist(err) {
		c.JSON(http.StatusOK, gin.H{"exists": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"exists": true})
	}
}

// mergeChunksHandler 合并分片
func merge(c *gin.Context) {
	filename := c.PostForm("filename")

	chunkDir := filepath.Join(uploadDir, filename+"_chunks")
	finalPath := filepath.Join(uploadDir, filename)

	// 创建最终文件
	outFile, err := os.Create(finalPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer outFile.Close()

	// 循环读取分片并写入最终文件
	chunkIndex := 0
	for {
		chunkPath := filepath.Join(chunkDir, fmt.Sprintf("chunk_%d", chunkIndex))
		chunkFile, err := os.Open(chunkPath)

		if os.IsNotExist(err) {
			// 分片不存在，说明合并完成
			break
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = io.Copy(outFile, chunkFile)
		chunkFile.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// 删除已合并的分片
		os.Remove(chunkPath)

		chunkIndex++
	}

	// 删除分片目录
	os.RemoveAll(chunkDir)

	c.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully"})
}

func main() {
	// 创建上传目录，如果不存在
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.MkdirAll(uploadDir, 0755)
	}

	router := gin.Default()

	// 设置跨域，允许所有来源
	router.Use(corsMiddleware())

	router.POST("/upload", upload)
	// 检查分片是否存在
	router.GET("/check", check)
	router.POST("/merge", merge)

	// 配置 TLS
	srv := http.Server{
		Addr:    ":8080",
		Handler: router,
		TLSConfig: &tls.Config{
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,   // 推荐
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,   // 推荐
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256, // 推荐
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, // 推荐
				tls.TLS_RSA_WITH_AES_128_GCM_SHA256,         // 兼容性
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,         // 兼容性
			},
			MinVersion: tls.VersionTLS12, // 强制使用 TLS 1.2 或更高版本
		},
	}

	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
