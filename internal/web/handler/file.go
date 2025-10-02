package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/service"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/common/http/binding"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// FileHandler 文件处理器结构体
//
// @description 处理文件相关的HTTP请求
// @struct
// @property *Handler 基础处理器，提供日志功能
// @property service *service.FileService 文件服务，处理文件相关的业务逻辑
type FileHandler struct {
	*Handler
	service *service.FileService
}

// NewFileHandler 创建文件处理器
//
// @param baseHandler *Handler 基础处理器
// @param service *service.FileService 文件服务
// @return *FileHandler 文件处理器实例
func NewFileHandler(baseHandler *Handler,
	fileService *service.FileService) *FileHandler {
	return &FileHandler{
		Handler: baseHandler,
		service: fileService,
	}
}

// InitUpload 初始化文件上传处理函数
//
// @description 初始化文件上传，为文件生成唯一ID并创建必要的存储结构
// @method POST
// @url /api/v1/files/init
// @param filename string 文件名 (必需)
// @param file_size int64 文件大小 (必需)
// @param chunk_size int64 分片大小 (必需)
// @return JSON 初始化结果，包含文件ID和分片信息
func (h *FileHandler) InitUpload(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "initUpload",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received file init upload handler...")

	// 解析请求参数
	req := new(request.FileInitUploadRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("file init upload request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.InitUpload(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// UploadChunk 上传文件分片处理函数
//
// @description 上传文件分片，支持大文件的分片上传
// @method POST
// @url /api/v1/files/chunk
// @param file_id int64 文件ID (必需)
// @param chunk_index int 分片索引 (必需)
// @param total_chunks int 总分片数 (必需)
// @param chunk_hash string 分片哈希值 (可选)
// @param file file 分片文件数据 (必需)
// @return JSON 上传结果
func (h *FileHandler) UploadChunk(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "uploadChunk",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received file upload chunk handler...")

	// 解析请求参数
	req := new(request.FileUploadChunkRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("file upload chunk request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.UploadChunk(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// CompleteUpload 完成文件上传处理函数
//
// @description 完成文件上传，合并所有已上传的分片生成最终文件
// @method POST
// @url /api/v1/files/complete
// @param file_id int64 文件ID (必需)
// @return JSON 完成上传结果
func (h *FileHandler) CompleteUpload(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "completeUpload",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received file complete upload handler...")

	// 解析请求参数
	req := new(request.CompleteUploadRequest)
	if err := binding.BindMultiPartForm(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("file complete upload request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.CompleteUpload(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// GetUploadStatus 获取上传状态处理函数
//
// @description 获取文件上传进度和已上传分片信息
// @method GET
// @url /api/v1/files/status
// @param file_id int64 文件ID (必需)
// @return JSON 上传状态信息
func (h *FileHandler) GetUploadStatus(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "getUploadStatus",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received file get upload status handler...")

	// 解析请求参数
	req := new(request.GetUploadStatusRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("file get upload status request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.GetUploadStatus(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// GetFileMetadata 获取文件元数据处理函数
//
// @description 获取文件详细信息
// @method GET
// @url /api/v1/files/metadata
// @param file_id int64 文件ID (必需)
// @return JSON 文件元数据信息
func (h *FileHandler) GetFileMetadata(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "getFileMetadata",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received file get file metadata handler...")

	// 解析请求参数
	req := new(request.GetFileMetadataRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("file get file metadata request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.GetFileMetadata(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// ResumeUpload 断点续传处理函数
//
// @description 获取已上传分片信息，支持断点续传
// @method GET
// @url /api/v1/files/resume
// @param file_id int64 文件ID (必需)
// @return JSON 断点续传信息
func (h *FileHandler) ResumeUpload(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "resumeUpload",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received file resume upload handler...")

	// 解析请求参数
	req := new(request.ResumeUploadRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	if !req.IsLegal() {
		// 验证失败
		h.logger.Errorf("file resume upload request is legal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.ResumeUpload(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// CheckExistingUpload 检查现有上传记录
func (h *FileHandler) CheckExistingUpload(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "checkExistingUpload",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received check existing upload handler...")

	// 获取查询参数
	fileName := ctx.Query("fileName")
	fileSizeStr := ctx.Query("fileSize")

	if fileName == "" || fileSizeStr == "" {
		h.logger.Errorf("fileName or fileSize is empty")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "文件名和文件大小不能为空")
		span.SetStatus(codes.Error, "文件名和文件大小不能为空")
		return
	}

	// 转换文件大小
	fileSize, err := strconv.ParseFloat(fileSizeStr, 64)
	if err != nil {
		h.logger.Errorf("parse fileSize failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "文件大小格式错误")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 调用服务层方法
	h.service.CheckExistingUpload(ctx, fileName, fileSize)
	span.SetStatus(codes.Ok, "success")
}

// DownloadFile 下载文件处理函数
//
// @description 提供文件下载功能
// @method GET
// @url /api/v1/files/:id/download
// @param id int64 文件ID (路径参数)
// @return 文件下载流
func (h *FileHandler) DownloadFile(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "downloadFile",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received file download file handler...")

	// 从路径参数中获取文件ID
	fileIDStr := ctx.Param("id")

	// 创建请求对象
	req := &request.DownloadFileRequest{
		UploadID: fileIDStr,
	}

	// 验证参数合法性
	if !req.IsLegal() {
		h.logger.Errorf("file download request is illegal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.DownloadFile(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// ListFiles 列出文件处理函数
//
// @description 列出所有文件
// @method GET
// @url /api/v1/files
// @param page int 页码 (可选)
// @param page_size int 每页数量 (可选)
// @param file_name string 文件名关键词搜索 (可选)
// @return JSON 文件列表
func (h *FileHandler) ListFiles(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "listFiles",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received list files handler...")

	// 解析请求参数
	req := new(request.ListFilesRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	// 验证参数合法性
	if !req.IsLegal() {
		h.logger.Errorf("list files request is illegal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.ListFiles(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// DeleteFile 删除文件处理函数
//
// @description 删除指定文件
// @method DELETE
// @url /api/v1/files/:id
// @param id int64 文件ID (路径参数)
// @return JSON 删除结果
func (h *FileHandler) DeleteFile(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "deleteFile",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received delete file handler...")

	// 从路径参数中获取文件ID
	uploadIdStr := ctx.Param("id")

	// 创建请求对象
	req := &request.DeleteFileRequest{
		UploadID: uploadIdStr,
	}

	// 验证参数合法性
	if !req.IsLegal() {
		h.logger.Errorf("delete file request is illegal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.DeleteFile(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// CleanupTempDir 清理临时文件夹处理函数
//
// @description 清理上传过程中产生的临时文件，释放存储空间
// @method DELETE
// @url /api/v1/files/temp
// @return JSON 清理结果
func (h *FileHandler) CleanupTempDir(ctx *gin.Context) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileHandler", "cleanupTempDir",
		attribute.String("requestId", requestid.Get(ctx)),
	)
	defer span.End()
	h.logger.Debugf("received cleanup temp dir handler...")

	req := new(request.CleanUpTempDirRequest)
	if err := binding.BindQuery(ctx, req); err != nil {
		messages := binding.GetValidationErrorMessages(err)
		h.logger.Errorf("binding request params failed: %v message: %v",
			err, messages)
		msg := make([]string, 0, len(messages))
		for _, message := range messages {
			msg = append(msg, message)
		}
		if len(msg) == 0 {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, err.Error())
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, stringer.Join(msg, ","))
			span.RecordError(err)
			span.SetStatus(codes.Error, stringer.Join(msg, ","))
		}
		return
	}

	// 验证参数合法性
	if !req.IsLegal() {
		h.logger.Errorf("cleanup temp dir request is illegal...")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, commonhttp.ErrMsgInvalidParameter)
		span.SetStatus(codes.Error, commonhttp.ErrMsgInvalidParameter)
		return
	}

	// 调用服务层方法
	h.service.CleanupTempDir(ctx, req)
	span.SetStatus(codes.Ok, "success")
}

// Routers 获取文件相关路由列表
//
// @return []commonhttp.RouterHandler 文件路由处理器列表
func (h *FileHandler) Routers() []commonhttp.RouterHandler {
	return []commonhttp.RouterHandler{
		// 初始化文件上传
		&commonhttp.MyRouter{
			Name:            "initUpload",
			Uri:             "files/init",
			Method:          http.MethodPost,
			HandlerFunc:     h.InitUpload,
			Enabled:         true,
			Desc:            "初始化文件上传",
			EnableJWtVerify: false,
		},
		// 上传文件分片
		&commonhttp.MyRouter{
			Name:            "uploadChunk",
			Uri:             "files/chunk",
			Method:          http.MethodPost,
			HandlerFunc:     h.UploadChunk,
			Enabled:         true,
			Desc:            "上传文件分片",
			EnableJWtVerify: false,
		},
		// 获取上传状态
		&commonhttp.MyRouter{
			Name:            "getUploadStatus",
			Uri:             "files/status",
			Method:          http.MethodGet,
			HandlerFunc:     h.GetUploadStatus,
			Enabled:         true,
			Desc:            "获取上传状态",
			EnableJWtVerify: false,
		},
		// 完成文件上传
		&commonhttp.MyRouter{
			Name:            "completeUpload",
			Uri:             "files/complete",
			Method:          http.MethodPost,
			HandlerFunc:     h.CompleteUpload,
			Enabled:         true,
			Desc:            "完成文件上传",
			EnableJWtVerify: false,
		},
		// 断点续传
		&commonhttp.MyRouter{
			Name:            "resumeUpload",
			Uri:             "files/resume",
			Method:          http.MethodGet,
			HandlerFunc:     h.ResumeUpload,
			Enabled:         true,
			Desc:            "断点续传",
			EnableJWtVerify: false,
		},
		// 列出文件
		&commonhttp.MyRouter{
			Name:            "listFiles",
			Uri:             "files",
			Method:          http.MethodGet,
			HandlerFunc:     h.ListFiles,
			Enabled:         true,
			Desc:            "列出文件",
			EnableJWtVerify: false,
		},
		// 获取文件元数据
		&commonhttp.MyRouter{
			Name:            "getFileMetadata",
			Uri:             "files/metadata",
			Method:          http.MethodGet,
			HandlerFunc:     h.GetFileMetadata,
			Enabled:         true,
			Desc:            "获取文件元数据",
			EnableJWtVerify: false,
		},
		// 下载文件
		&commonhttp.MyRouter{
			Name:            "downloadFile",
			Uri:             "files/:id/download",
			Method:          http.MethodGet,
			HandlerFunc:     h.DownloadFile,
			Enabled:         true,
			Desc:            "下载文件",
			EnableJWtVerify: false,
		},
		// 删除文件
		&commonhttp.MyRouter{
			Name:            "deleteFile",
			Uri:             "files/:id",
			Method:          http.MethodDelete,
			HandlerFunc:     h.DeleteFile,
			Enabled:         true,
			Desc:            "删除文件",
			EnableJWtVerify: false,
		},
		// 清理临时文件夹
		&commonhttp.MyRouter{
			Name:            "cleanupTempDir",
			Uri:             "files/temp",
			Method:          http.MethodDelete,
			HandlerFunc:     h.CleanupTempDir,
			Enabled:         true,
			Desc:            "清理临时文件夹",
			EnableJWtVerify: false,
		},
		// 检查现有上传记录
		&commonhttp.MyRouter{
			Name:            "checkExistingUpload",
			Uri:             "files/check",
			Method:          http.MethodGet,
			HandlerFunc:     h.CheckExistingUpload,
			Enabled:         true,
			Desc:            "检查现有上传记录",
			EnableJWtVerify: false,
		},
	}
}
