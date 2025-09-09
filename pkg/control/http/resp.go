package http

import (
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/gm-fabric-deployment/pkg/json"
	"io"
	"net/http"
	"os"
	"strings"
)

type BaseResponse struct {
	Code    int         `json:"code" yaml:"code"`
	Data    interface{} `json:"data" yaml:"data"`
	Success bool        `json:"success" yaml:"success"`
	Message string      `json:"message" yaml:"message"`
}

func (b BaseResponse) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

func SuccessResponse(ctx *gin.Context, data interface{}) {
	SuccessResponseWithMessage(ctx, data, "业务处理成功")
}

func SuccessResponseWithMessage(ctx *gin.Context, data interface{}, message string) {
	ctx.JSON(http.StatusOK, BaseResponse{
		Code:    http.StatusOK,
		Data:    data,
		Success: true,
		Message: message,
	})
}

func FailedResponse(ctx *gin.Context, err Error) {
	ctx.JSON(http.StatusOK, BaseResponse{
		Code:    int(err.Code),
		Data:    err.Message,
		Success: false,
		Message: "业务处理失败",
	})
}

func FailedResponseWithMessage(ctx *gin.Context, errCode ErrCode, message string) {
	FailedResponse(ctx, NewError(errCode, message))
}

// 检查文件路径是否安全 (示例)
func isSafePath(filePath string) bool {
	// filePath, err := filepath.Abs(filePath)
	// if err != nil {
	//	return false
	// }
	// basePath, err := filepath.Abs("./") // 项目根目录
	// if err != nil {
	//	return false
	// }
	//
	// return strings.HasPrefix(filePath, basePath)

	// 这里只做演示，实际情况请根据你的项目来做更严格的校验
	return !strings.Contains(filePath, "..")
}

// fileDownload 下载文件
func fileDownload(ctx *gin.Context, filepath, fileName string) {

	// 检查文件路径的安全性 (非常重要!)
	if !isSafePath(filepath) {
		FailedResponseWithMessage(ctx, NormalFailed, "非法文件路径")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		ctx.Status(http.StatusNotFound)
		ctx.Abort()
		return
	}

	// 设置 HTTP 响应头
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Type", "application/octet-stream")
	// ctx.Header("Accept-Ranges", "bytes") // 支持断点续传

	// 设置响应为 chunked，即分块传输
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Writer.Flush() // 发送响应头

	// 打开文件

	file, err := os.Open(filepath)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
		}
	}(file)

	// 将文件内容传输到响应体
	_, err = io.Copy(ctx.Writer, file)
	if err != nil && err != io.EOF {
		FailedResponseWithMessage(ctx, NormalFailed, "文件下载失败")
		return
	}
}

// DownloadFile 直接发送文件，不使用 BaseResponse 结构体
func DownloadFile(ctx *gin.Context, filePath, filename string) {
	// 检查文件是否存在（可选）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		FailedResponseWithMessage(ctx, FileNotFound, "文件不存在")
		return
	}

	ctx.FileAttachment(filePath, filename) // 强制下载
	// ctx.File(filePath) // 直接在浏览器中显示，如果浏览器支持
}

// DownloadFileWithDisposition 发送文件并设置 Content-Disposition
func DownloadFileWithDisposition(ctx *gin.Context, filePath, filename string) {
	// 检查文件是否存在（可选）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		FailedResponseWithMessage(ctx, FileNotFound, "文件不存在")
		return
	}

	ctx.Header("Content-Description", "File Transfer")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Disposition", "attachment; filename="+filename)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(filePath)
}
