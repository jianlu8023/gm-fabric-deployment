package http

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/golang-example/pkg/json"
)

// BaseResponse HTTP基础响应结构体
//
// @description 定义所有HTTP响应的基础结构
// @struct
type BaseResponse struct {
	// Code 响应状态码
	Code int `json:"code" yaml:"code"`
	// Data 响应数据，根据不同接口返回不同类型的数据
	Data interface{} `json:"data" yaml:"data"`
	// Success 操作是否成功
	Success bool `json:"success" yaml:"success"`
	// Message 响应消息，通常用于描述操作结果或错误信息
	Message string `json:"message" yaml:"message"`
}

// String 将BaseResponse转换为字符串
//
// @description 将BaseResponse结构体转换为JSON字符串
// @return string 格式化的JSON字符串
func (b BaseResponse) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

// SuccessResponse 发送成功响应
//
// @description 发送HTTP成功响应，默认消息为"业务处理成功"
// @param ctx *gin.Context Gin上下文
// @param data interface{} 响应数据
func SuccessResponse(ctx *gin.Context, data interface{}) {
	SuccessResponseWithMessage(ctx, data, "业务处理成功")
}

// SuccessResponseWithMessage 发送带自定义消息的成功响应
//
// @description 发送HTTP成功响应，可以自定义响应消息
// @param ctx *gin.Context Gin上下文
// @param data interface{} 响应数据
// @param message string 自定义响应消息
func SuccessResponseWithMessage(ctx *gin.Context, data interface{}, message string) {
	ctx.JSON(http.StatusOK, BaseResponse{
		Code:    http.StatusOK,
		Data:    data,
		Success: true,
		Message: message,
	})
}

// FailedResponse 发送错误响应
//
// @description 发送HTTP错误响应，使用Error结构体中的错误信息
// @param ctx *gin.Context Gin上下文
// @param err Error 错误结构体
func FailedResponse(ctx *gin.Context, err Error) {
	ctx.JSON(http.StatusOK, BaseResponse{
		Code:    int(err.Code),
		Data:    err.Message,
		Success: false,
		Message: "业务处理失败",
	})
}

// FailedResponseWithMessage 发送带自定义消息的错误响应
//
// @description 发送HTTP错误响应，可以自定义错误码和错误消息
// @param ctx *gin.Context Gin上下文
// @param errCode ErrCode 错误码
// @param message string 自定义错误消息
func FailedResponseWithMessage(ctx *gin.Context, errCode ErrCode, message string) {
	FailedResponse(ctx, NewError(errCode, message))
}

// isSafePath 检查文件路径是否安全
//
// @description 验证文件路径是否安全，防止目录遍历攻击
// @param filePath string 需要检查的文件路径
// @return bool 文件路径是否安全
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

// fileDownload 下载文件（内部实现）
//
// @description 实现文件下载功能，包括路径安全检查和分块传输
// @param ctx *gin.Context Gin上下文
// @param filepath string 文件路径
// @param fileName string 下载时显示的文件名
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

// DownloadFile 下载文件（简单版）
//
// @description 直接发送文件，不使用BaseResponse结构体封装，强制文件下载
// @param ctx *gin.Context Gin上下文
// @param filePath string 文件路径
// @param filename string 下载时显示的文件名
func DownloadFile(ctx *gin.Context, filePath, filename string) {
	// 检查文件是否存在（可选）
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		FailedResponseWithMessage(ctx, FileNotFound, "文件不存在")
		return
	}

	ctx.FileAttachment(filePath, filename) // 强制下载
	// ctx.File(filePath) // 直接在浏览器中显示，如果浏览器支持
}

// DownloadFileWithDisposition 带自定义头的文件下载
//
// @description 发送文件并设置详细的Content-Disposition和其他HTTP头
// @param ctx *gin.Context Gin上下文
// @param filePath string 文件路径
// @param filename string 下载时显示的文件名
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
