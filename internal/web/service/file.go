package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/go-tools/v2/pkg/sqlnull"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type FileService interface {
	InitUpload(ctx *gin.Context, req *request.FileInitUploadRequest)
	UploadChunk(ctx *gin.Context, req *request.FileUploadChunkRequest)
	CompleteUpload(ctx *gin.Context, req *request.CompleteUploadRequest)
	GetUploadStatus(ctx *gin.Context, req *request.GetUploadStatusRequest)
	GetFileMetadata(ctx *gin.Context, req *request.GetFileMetadataRequest)
	ResumeUpload(ctx *gin.Context, req *request.ResumeUploadRequest)
	CheckExistingUpload(ctx *gin.Context, fileName string, fileSize float64)
	DownloadFile(ctx *gin.Context, req *request.DownloadFileRequest)
	ListFiles(ctx *gin.Context, req *request.ListFilesRequest)
	DeleteFile(ctx *gin.Context, req *request.DeleteFileRequest)
	CleanupTempDir(ctx *gin.Context, req *request.CleanUpTempDirRequest)
}

// fileService 文件服务结构体
//
// @description 提供文件相关的服务功能，如文件上传、下载、管理等
// @struct
type fileService struct {
	*Service                    // Service 基础服务，提供日志功能
	mapper    mapper.FileMapper // mapper 文件映射器，用于数据访问
	uploadDir string            // uploadDir 上传文件保存目录
	tempDir   string            // tempDir 临时文件保存目录
}

// NewFileService 创建文件服务实例
//
// @description 创建并返回一个新的文件服务实例
// @param baseService *Service 基础服务
// @param fileMapper mapper.FileMapper 文件映射器
// @param uploadDir string 上传目录
// @param uploadCacheDir string 上传缓存目录
// @return FileService 文件服务实例
func NewFileService(baseService *Service, fileMapper mapper.FileMapper, uploadDir, uploadCacheDir string) FileService {
	if stringer.IsBlank(uploadDir) {
		uploadDir = "uploads"
	}
	if stringer.IsBlank(uploadCacheDir) {
		uploadCacheDir = filepath.Clean(filepath.Join(uploadDir, "temp"))
	}

	// 确保目录存在
	ensureDir(uploadDir)
	ensureDir(uploadCacheDir)

	return &fileService{
		Service:   baseService,
		mapper:    fileMapper,
		uploadDir: uploadDir,
		tempDir:   uploadCacheDir,
	}
}

// ensureDir 确保目录存在，如果不存在则创建
//
// @description 确保指定的目录路径存在，如果不存在则创建该目录
// @param dir string 目录路径
func ensureDir(dir string) {
	_, _ = path.CreateDir(dir)
}

// CleanupTempDir 清理临时文件夹
//
// @description 删除临时文件夹下的所有内容，用于清理上传过程中产生的临时文件
// @param ctx *gin.Context Gin上下文
// @param req *request.CleanUpTempDirRequest 清理临时文件夹请求参数
func (s *fileService) CleanupTempDir(ctx *gin.Context, req *request.CleanUpTempDirRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "cleanupTempDir",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received file cleanup temp dir request with params: %v", req)

	tempPath := s.tempDir
	if !stringer.IsBlank(req.UploadID) {
		tempPath = filepath.Clean(filepath.Join(s.tempDir, req.UploadID))
	}

	// 检查临时文件夹是否存在
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		s.logger.Warnf("临时文件夹不存在: %s", s.tempDir)
		commonhttp.SuccessResponse(ctx, map[string]interface{}{
			"message":  "临时文件夹不存在",
			"temp_dir": tempPath,
			"cleaned":  false,
		})
		return
	}

	// 清理临时文件夹内容
	if err := path.ClearDir(tempPath); err != nil {
		s.logger.Errorf("清理临时文件夹失败: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, fmt.Sprintf("清理临时文件夹失败: %v", err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	s.logger.Infof("临时文件夹清理成功: %s", tempPath)
	commonhttp.SuccessResponse(ctx, map[string]interface{}{
		"message":  "临时文件夹清理成功",
		"temp_dir": tempPath,
		"cleaned":  true,
	})

	span.SetStatus(codes.Ok, "success")
}

// InitUpload 初始化文件上传服务
//
// @description 初始化文件上传，为文件生成唯一ID并创建必要的存储结构
// @param ctx *gin.Context Gin上下文
// @param req *request.FileInitUploadRequest 初始化上传请求参数
func (s *fileService) InitUpload(ctx *gin.Context, req *request.FileInitUploadRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "initUpload",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received file init upload request with params: %v", req)

	// 计算总分片数 - 对小文件的特殊处理
	var totalChunks float64
	var actualChunkSize float64

	// 如果文件大小小于或等于分片大小，则只需要1个分片
	if req.FileSize <= req.ChunkSize {
		totalChunks = 1
		actualChunkSize = req.FileSize // 对于小文件，实际分片大小就是文件大小
		s.logger.Debugf("小文件处理: 文件大小=%.0f, 分片大小=%.0f, 总分片数=%.0f", req.FileSize, actualChunkSize, totalChunks)
	} else {
		// 正常大文件处理
		totalChunks = math.Ceil(req.FileSize / req.ChunkSize)
		actualChunkSize = req.ChunkSize
		s.logger.Debugf("大文件处理: 文件大小=%.0f, 分片大小=%.0f, 总分片数=%.0f", req.FileSize, actualChunkSize, totalChunks)
	}

	// 防御性检查：确保至少有1个分片
	if totalChunks < 1 {
		totalChunks = 1
	}

	// 生成上传ID
	uploadID := uuid.GetUUID()

	// 先检查文件是否已存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{
		FileName:   req.FileName,
		UploaderId: req.UploaderId,
	})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if fileExist {
		s.logger.Errorf("file not found: fileName: %v uploaderId: %v", req.FileName, req.UploaderId)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件已存在")
		span.SetStatus(codes.Error, "文件已存在")
		return
	}

	// 创建文件信息记录
	fileInfo := &model.FileInfo{
		UploadID:       uploadID,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		FilePath:       filepath.Clean(filepath.Join(s.uploadDir, req.FileName)),
		ChunkSize:      actualChunkSize, // 使用实际计算的分片大小
		TotalChunks:    totalChunks,
		UploadedChunks: 0,
		FileType:       req.FileType,
		Description:    req.Description,
		StorageType:    req.StorageType,
		Status:         model.FileStatusUploading,
		UploadTime:     time.Now(),
		UpdateTime:     time.Now(),
	}
	if stringer.IsBlank(fileInfo.FileType) {
		fileInfo.FileType = filepath.Ext(req.FileName)
	}

	// 保存文件信息
	if err = s.mapper.CreateFileInfo(ctx.Request.Context(), fileInfo); err != nil {
		s.logger.Errorf("save file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "保存文件信息失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	// 创建临时目录用于存储分片
	ensureDir(filepath.Join(s.tempDir, uploadID))

	commonhttp.SuccessResponse(ctx, response.NewInitUploadResponse(uploadID, totalChunks, actualChunkSize))
	span.SetStatus(codes.Ok, "success")
}

// UploadChunk 上传文件分片服务
//
// @description 上传文件分片，支持大文件的分片上传
// @param ctx *gin.Context Gin上下文
// @param req *request.FileUploadChunkRequest 上传分片请求参数
func (s *fileService) UploadChunk(ctx *gin.Context, req *request.FileUploadChunkRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "uploadChunk",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: uploadId=%v", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		span.SetStatus(codes.Error, "文件记录不存在")
		return
	}

	// 获取文件信息
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 检查文件状态
	if fileInfo.Status != model.FileStatusUploading {
		s.logger.Errorf("file status incorrect: %v", fileInfo.Status)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件状态不正确，无法上传分片")
		span.SetStatus(codes.Error, "文件状态不正确，无法上传分片")
		return
	}

	// 检查分片索引的合法性 - 增强验证
	if req.ChunkIndex < 0 || float64(req.ChunkIndex) >= fileInfo.TotalChunks {
		s.logger.Errorf("invalid chunk index: %d, total chunks: %.0f", req.ChunkIndex, fileInfo.TotalChunks)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter,
			fmt.Sprintf("分片索引不合法: %d，总分片数: %.0f", req.ChunkIndex, fileInfo.TotalChunks))
		span.SetStatus(codes.Error, "分片索引不合法")
		return
	}

	// 检查总分片数一致性
	if req.TotalChunks != int(fileInfo.TotalChunks) {
		s.logger.Errorf("total chunks mismatch: request=%d, stored=%.0f", req.TotalChunks, fileInfo.TotalChunks)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter,
			fmt.Sprintf("总分片数不匹配: 请求=%d，存储=%.0f", req.TotalChunks, fileInfo.TotalChunks))
		span.SetStatus(codes.Error, "总分片数不匹配")
		return
	}

	// 打开上传的文件
	src, err := req.File.Open()
	if err != nil {
		s.logger.Errorf("open uploaded file failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "打开上传文件失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	defer func(src multipart.File) {
		if err := src.Close(); err != nil {
			s.logger.Errorf("close uploaded file failed: %v", err)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	}(src)

	// 保存分片文件
	tempChunkDir := filepath.Join(s.tempDir, fileInfo.UploadID)
	chunkFilePath := filepath.Join(tempChunkDir, fmt.Sprintf("chunk_%d", req.ChunkIndex))

	// 打开或创建分片文件
	chunkOutput, err := os.Create(chunkFilePath)
	if err != nil {
		s.logger.Errorf("create chunk file failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "创建分片文件失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}
	defer func(chunkOutput *os.File) {
		if err := chunkOutput.Close(); err != nil {
			s.logger.Errorf("close chunk file failed: %v", err)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	}(chunkOutput)

	// if err = ctx.SaveUploadedFile(req.File, chunkFilePath); err != nil {
	// 	s.logger.Errorf("save file failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, err.Error())
	// 	span.SetStatus(codes.Error, err.Error())
	// 	span.RecordError(err)
	// 	return
	// }

	// 写入分片数据
	written, err := io.Copy(chunkOutput, src)
	if err != nil {
		s.logger.Errorf("write chunk data failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "写入分片数据失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	// 获取文件的实际大小
	chunkSize := written

	// 增强验证：检查分片大小的合理性
	if written <= 0 {
		s.logger.Errorf("chunk data is empty: chunk_index=%d", req.ChunkIndex)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "分片数据为空")
		span.RecordError(fmt.Errorf("分片数据为空"))
		span.SetStatus(codes.Error, "分片数据为空")
		return
	}

	// 对于非最后一个分片，检查大小是否合理
	// 修复：索引从0开始，所以最后一个分片的索引是totalChunks-1
	isLastChunk := req.ChunkIndex == (req.TotalChunks - 1)
	expectedSize := fileInfo.ChunkSize
	if isLastChunk {
		// 最后一个分片的期望大小
		remaining := fileInfo.FileSize - float64(req.ChunkIndex)*fileInfo.ChunkSize
		if remaining > 0 {
			expectedSize = remaining
		}
	}

	// 验证分片大小（允许一定的误差）
	if !isLastChunk && math.Abs(float64(written)-expectedSize) > 1 {
		s.logger.Warnf("chunk size mismatch: chunk_index=%d, expected=%.0f, actual=%d",
			req.ChunkIndex, expectedSize, written)
	}

	s.logger.Debugf("chunk upload: index=%d, size=%d, expected=%.0f, is_last=%t, total_chunks=%d",
		req.ChunkIndex, written, expectedSize, isLastChunk, req.TotalChunks)

	calcHash, err := calculateFileHash(chunkFilePath)
	if err != nil {
		s.logger.Errorf("calculate chunk hash failed: %v", err)
	}

	// 记录已上传分片 - 增加调试信息
	fileChunk := &model.FileChunk{
		UploadID:      fileInfo.UploadID,
		ChunkIndex:    float64(req.ChunkIndex),
		ChunkSize:     float64(chunkSize),
		ChunkHash:     req.ChunkHash,
		CalcChunkSize: calcHash,
		FilePath:      chunkFilePath,
		UploadTime:    time.Now(),
		Status:        sqlnull.TrueToNull(),
	}

	s.logger.Debugf("分片上传详情: uploadId=%s, chunkIndex=%d, chunkSize=%d, 期望大小=%.0f, 是否最后分片=%t",
		fileInfo.UploadID, req.ChunkIndex, chunkSize, expectedSize, isLastChunk)

	// 检查分片是否已存在
	exists, err := s.mapper.CheckFileChunkExists(ctx.Request.Context(), model.FileChunk{
		UploadID:   fileInfo.UploadID,
		ChunkIndex: float64(req.ChunkIndex),
	})
	if err != nil {
		s.logger.Errorf("check chunk exists failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查分片是否存在失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	if exists {
		// 更新分片信息
		if err = s.mapper.UpdateFileChunk(ctx.Request.Context(), fileChunk); err != nil {
			s.logger.Errorf("update file chunk failed: %v", err)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "更新分片信息失败")
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return
		}
	} else {
		// 插入新分片信息
		if err = s.mapper.CreateFileChunk(ctx.Request.Context(), fileChunk); err != nil {
			s.logger.Errorf("save file chunk failed: %v", err)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "保存分片信息失败")
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return
		}
	}

	// 更新已上传分片数
	uploadedChunks, err := s.mapper.GetUploadedChunkCount(ctx.Request.Context(), req.UploadID)
	if err != nil {
		s.logger.Errorf("get uploaded chunk count failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片数失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 更新文件信息 - 增强错误处理
	fileInfo.UploadedChunks = uploadedChunks
	fileInfo.UpdateTime = time.Now()
	fileInfo.LastChunkTime = time.Now() // 更新最后分片上传时间

	if uploadedChunks >= fileInfo.TotalChunks {
		fileInfo.Status = model.FileStatusPendingMerge // 等待合并
		s.logger.Debugf("文件 %s 所有分片上传完成，状态更新为等待合并", fileInfo.FileName)
	}

	// 尝试更新文件信息，如果失败不影响分片上传成功的响应
	if err = s.mapper.UpdateFileInfo(ctx.Request.Context(), fileInfo); err != nil {
		s.logger.Errorf("update file info failed: %v, but chunk upload success", err)
		// 不返回错误，只记录日志，因为分片已经成功保存
		// 可以在后续的合并操作中重新计算和更新
		// 注意：这里不应该返回错误，因为分片已经成功保存
		s.logger.Warnf("分片 %d 上传成功，但更新文件信息失败，将在后续操作中重新计算", req.ChunkIndex)
	}

	// 计算上传进度（即使更新文件信息失败也要返回成功响应）
	uploadedChunksForProgress := uploadedChunks
	if err != nil {
		// 如果更新失败，重新获取一次分片数用于进度计算
		if retryCount, retryErr := s.mapper.GetUploadedChunkCount(ctx.Request.Context(), req.UploadID); retryErr == nil {
			uploadedChunksForProgress = retryCount
		}
	}

	progress := 0.0
	if fileInfo.TotalChunks > 0 {
		progress = uploadedChunksForProgress / fileInfo.TotalChunks * 100
	}

	s.logger.Infof("upload chunk success, fileAutoUid: %d, chunkIndex: %d, progress: %.2f%%", fileInfo.AutoUid, req.ChunkIndex, progress)

	commonhttp.SuccessResponse(ctx, response.NewUploadChunkResponse(req.UploadID, fileChunk.ChunkIndex,
		uploadedChunksForProgress, fileInfo.TotalChunks, progress,
	))
	span.SetStatus(codes.Ok, "success")
}

// CompleteUpload 完成文件上传服务
//
// @description 合并所有分片文件，完成文件上传过程
// @param ctx *gin.Context Gin上下文
// @param req *request.CompleteUploadRequest 完成上传请求参数
func (s *fileService) CompleteUpload(ctx *gin.Context, req *request.CompleteUploadRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "completeUpload",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: uploadId=%v", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		span.SetStatus(codes.Error, "文件记录不存在")
		return
	}

	// 获取文件信息
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 检查文件状态
	if fileInfo.Status != model.FileStatusUploading && fileInfo.Status != model.FileStatusPendingMerge {
		s.logger.Errorf("file status incorrect: %v", fileInfo.Status)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件状态不正确，无法完成上传")
		span.SetStatus(codes.Error, "文件状态不正确，无法完成上传")
		return
	}

	// 检查所有分片是否都已上传
	uploadedChunks, err := s.mapper.GetUploadedChunkCount(ctx.Request.Context(), req.UploadID)
	if err != nil {
		s.logger.Errorf("get uploaded chunk count failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片数失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	if uploadedChunks < fileInfo.TotalChunks {
		s.logger.Errorf("not all chunks uploaded: %f/%f", uploadedChunks, fileInfo.TotalChunks)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError,
			fmt.Sprintf("分片未全部上传：已上传 %f 个，共 %f 个", uploadedChunks, fileInfo.TotalChunks))
		span.SetStatus(codes.Error,
			fmt.Sprintf("分片未全部上传：已上传 %f 个，共 %f 个", uploadedChunks, fileInfo.TotalChunks))
		return
	}

	// 获取所有分片信息 - 增强调试
	chunks, err := s.mapper.GetFileChunks(ctx.Request.Context(), model.FileChunk{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file chunks failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取分片信息失败")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return
	}

	// 数据一致性检查和调试信息
	s.logger.Debugf("合并前检查: uploadId=%s, 预期分片数=%.0f, 数据库分片数=%d, 统计分片数=%.0f",
		req.UploadID, fileInfo.TotalChunks, len(chunks), uploadedChunks)

	// 检查分片索引的完整性
	expectedIndexes := make(map[int]bool)
	actualIndexes := make(map[int]bool)

	// 生成预期的索引集合（0 到 totalChunks-1）
	for i := 0; i < int(fileInfo.TotalChunks); i++ {
		expectedIndexes[i] = true
	}

	// 记录实际的索引
	for _, chunk := range chunks {
		intIndex := int(chunk.ChunkIndex)
		actualIndexes[intIndex] = true
		s.logger.Debugf("分片信息: 索引=%.0f, 大小=%.0f, 路径=%s",
			chunk.ChunkIndex, chunk.ChunkSize, chunk.FilePath)
	}

	// 找出缺失的分片
	missingChunks := make([]int, 0)
	for expectedIndex := range expectedIndexes {
		if !actualIndexes[expectedIndex] {
			missingChunks = append(missingChunks, expectedIndex)
		}
	}

	if len(missingChunks) > 0 {
		s.logger.Errorf("发现缺失分片: %v, uploadId: %s", missingChunks, req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError,
			fmt.Sprintf("分片缺失：缺少分片 %v，已上传 %d 个，共 %.0f 个",
				missingChunks, len(chunks), fileInfo.TotalChunks))
		span.SetStatus(codes.Error,
			fmt.Sprintf("分片缺失：缺少分片 %v", missingChunks))
		return
	}

	// 按分片索引排序
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkIndex < chunks[j].ChunkIndex
	})

	// 创建最终文件
	fileExt := filepath.Ext(fileInfo.FileName)
	fileNameWithoutExt := fileInfo.FileName[:len(fileInfo.FileName)-len(fileExt)]
	destFileName := fmt.Sprintf("%s_%s%s", fileNameWithoutExt, time.Now().Format("20060102_150405"), fileExt)
	destFilePath := filepath.Join(s.uploadDir, destFileName)

	// 打开目标文件
	destFile, err := os.Create(destFilePath)
	if err != nil {
		s.logger.Errorf("create destination file failed: %v", err)
		// 合并失败时逻辑删除相关记录
		s.cleanupFailedUpload(ctx.Request.Context(), req.UploadID, "创建目标文件失败")
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "创建目标文件失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	defer func(destFile *os.File) {
		if err := destFile.Close(); err != nil {
			s.logger.Errorf("close file failed: %v", err)
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}
	}(destFile)

	// 合并分片 - 修复索引问题
	s.logger.Debugf("开始合并 %d 个分片", len(chunks))
	for i, chunk := range chunks {
		s.logger.Debugf("合并分片 %d: index=%.0f, path=%s", i, chunk.ChunkIndex, chunk.FilePath)

		// 打开分片文件
		chunkFile, err := os.Open(chunk.FilePath)
		if err != nil {
			s.logger.Errorf("open chunk file failed: %v", err)
			// 合并失败时逻辑删除相关记录
			s.cleanupFailedUpload(ctx.Request.Context(), req.UploadID, "打开分片文件失败")
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "打开分片文件失败")
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return
		}

		// 写入目标文件
		size, err := io.Copy(destFile, chunkFile)

		if err != nil {
			s.logger.Errorf("merge chunk data failed: %v", err)
			// 合并失败时逻辑删除相关记录
			s.cleanupFailedUpload(ctx.Request.Context(), req.UploadID, "合并分片数据失败")
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "合并分片数据失败")
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return
		}

		if err = chunkFile.Close(); err != nil {
			s.logger.Errorf("close file failed: %v", err)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, err.Error())
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
		}

		// 修复索引判断逻辑：
		// - chunks数组按ChunkIndex排序，最后一个元素是最大索引
		// - 最大索引应该是totalChunks-1（因为索引从0开始）
		// - 这里判断是否为最后一个分片
		isLastChunk := chunk.ChunkIndex == chunks[len(chunks)-1].ChunkIndex

		// 验证分片大小：非最后分片应该等于标准分片大小，最后分片可能小于标准大小
		if !isLastChunk && float64(size) != chunk.ChunkSize {
			s.logger.Warnf("chunk size mismatch: expected %.0f bytes, got %d bytes for chunk %.0f (not last chunk)",
				chunk.ChunkSize, size, chunk.ChunkIndex)
		} else if isLastChunk {
			s.logger.Debugf("last chunk size: %.0f bytes for chunk %.0f", chunk.ChunkSize, chunk.ChunkIndex)
		}

		s.logger.Debugf("merged chunk %.0f successfully, size: %d bytes", chunk.ChunkIndex, size)
	}

	// 计算文件哈希值
	fileHash, err := calculateFileHash(destFilePath)
	if err != nil {
		s.logger.Warnf("calculate file hash failed: %v", err)
		// 不中断流程，继续处理
	}

	// // 更新文件信息
	fileInfo.Status = model.FileStatusCompleted
	fileInfo.FilePath = destFilePath
	fileInfo.FileHash = fileHash
	fileInfo.UpdateTime = time.Now()

	if err = s.mapper.UpdateFileInfo(ctx.Request.Context(), fileInfo); err != nil {
		s.logger.Errorf("update file info failed: %v", err)
		// 更新文件信息失败时也进行清理，但不删除已生成的文件
		s.logger.Warnf("文件合并成功但更新数据库失败，保留已生成的文件: %s", destFilePath)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "更新文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 清理临时文件
	go func() {
		tempChunkDir := filepath.Join(s.tempDir, fileInfo.UploadID)
		if err := path.ClearDir(tempChunkDir); err != nil {
			s.logger.Errorf("clear temp chunk dir failed: %v", err)
			span.RecordError(err)
		}
	}()

	// 生成下载URL
	downloadURL := fmt.Sprintf("/files/%s/download", req.UploadID)

	commonhttp.SuccessResponse(ctx, response.NewCompleteUploadResponse(
		req.UploadID,
		fileInfo.FileName,
		fileInfo.FileSize,
		fileInfo.FileHash,
		destFilePath,
		downloadURL,
		fileInfo.UpdateTime,
	))
	span.SetStatus(codes.Ok, "success")
}

// CheckExistingUpload 检查现有上传记录
//
// @description 检查指定文件名和文件大小的现有上传记录
// @param ctx *gin.Context HTTP上下文
// @param fileName string 文件名
// @param fileSize float64 文件大小
func (s *fileService) CheckExistingUpload(ctx *gin.Context, fileName string, fileSize float64) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "checkExistingUpload")
	defer span.End()
	s.logger.Debugf("checking existing upload for file: %s, size: %.0f", fileName, fileSize)

	// 构造查询条件，查找未完成的上传记录
	query := model.FileInfo{
		FileName: fileName,
		FileSize: fileSize,
		Status:   model.FileStatusUploading, // 只查找正在上传的记录
	}

	// 查询现有的上传记录
	existingFile, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), query)
	if err != nil {
		// 如果没找到记录，尝试查找等待合并的记录
		query.Status = model.FileStatusPendingMerge
		existingFile, err = s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), query)
		if err != nil {
			// 没有找到任何未完成的上传记录
			s.logger.Debugf("no existing upload record found for file: %s", fileName)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "没有找到现有的上传记录")
			span.SetStatus(codes.Ok, "no existing upload found")
			return
		}
	}

	// 找到了未完成的上传记录
	s.logger.Infof("found existing upload record: uploadId=%s, status=%s", existingFile.UploadID, existingFile.Status)

	// 获取已上传的分片索引
	uploadedIndexes, err := s.mapper.GetUploadedFileChunkIndexes(ctx.Request.Context(), model.FileChunk{
		UploadID: existingFile.UploadID,
	})
	if err != nil {
		s.logger.Warnf("get uploaded chunk indexes failed: %v", err)
		uploadedIndexes = []int{} // 默认为空
	}

	commonhttp.SuccessResponse(ctx, map[string]interface{}{
		"upload_id":              existingFile.UploadID,
		"file_name":              existingFile.FileName,
		"file_size":              existingFile.FileSize,
		"chunk_size":             existingFile.ChunkSize,
		"total_chunks":           existingFile.TotalChunks,
		"uploaded_chunks":        existingFile.UploadedChunks,
		"status":                 existingFile.Status,
		"upload_time":            existingFile.UploadTime,
		"update_time":            existingFile.UpdateTime,
		"uploaded_chunk_indexes": uploadedIndexes,
	})
	span.SetStatus(codes.Ok, "success")
}

// GetUploadStatus 获取文件上传状态服务
//
// @description 获取指定文件的上传进度、已上传分片信息和当前状态
// @param ctx *gin.Context Gin上下文
// @param req *request.GetUploadStatusRequest 获取上传状态请求参数
func (s *fileService) GetUploadStatus(ctx *gin.Context, req *request.GetUploadStatusRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "getUploadStatus",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: uploadId=%v", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		span.SetStatus(codes.Error, "文件记录不存在")
		return
	}

	// 获取文件信息
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 获取已上传分片数
	uploadedChunks, err := s.mapper.GetUploadedChunkCount(ctx.Request.Context(), req.UploadID)
	if err != nil {
		s.logger.Errorf("get uploaded chunk count failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片数失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 获取已上传分片索引 - 增强调试信息
	chunkIndices, err := s.mapper.GetUploadedFileChunkIndexes(ctx.Request.Context(), model.FileChunk{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get uploaded chunk indices failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片索引失败")
		return
	}

	// 调试日志：详细记录已上传分片信息
	s.logger.Debugf("upload status for %s: total_chunks=%.0f, db_count=%.0f, chunk_indexes=%v",
		req.UploadID, fileInfo.TotalChunks, uploadedChunks, chunkIndices)

	// 数据一致性检查
	if float64(len(chunkIndices)) != uploadedChunks {
		s.logger.Warnf("数据不一致: 索引数组长度=%d, 数据库计数=%.0f, uploadId=%s",
			len(chunkIndices), uploadedChunks, req.UploadID)
	}

	// 计算上传进度
	progress := 0.0
	if fileInfo.TotalChunks > 0 {
		progress = uploadedChunks / fileInfo.TotalChunks * 100
	}

	commonhttp.SuccessResponse(ctx, response.NewGetUploadStatusResponse(
		req.UploadID,
		fileInfo.FileName,
		fileInfo.FileSize,
		fileInfo.ChunkSize,
		fileInfo.TotalChunks,
		fileInfo.UploadedChunks,
		progress,
		string(fileInfo.Status),
		chunkIndices,
		fileInfo.LastChunkTime,
	))
	span.SetStatus(codes.Ok, "success")
}

// cleanupFailedUpload 清理失败的上传记录
//
// @description 当文件合并失败时，逻辑删除相关的文件信息和分片记录
// @param ctx context.Context 上下文
// @param uploadID string 上传ID
// @param reason string 失败原因
func (s *fileService) cleanupFailedUpload(ctx context.Context, uploadID string, reason string) {
	_, span := tracer.StartSpan(ctx, "fileService", "cleanupFailedUpload")
	defer span.End()

	s.logger.Warnf("开始清理失败的上传记录: uploadID=%s, 原因=%s", uploadID, reason)

	// 1. 逻辑删除所有分片记录
	if err := s.mapper.DeleteFileChunks(ctx, model.FileChunk{
		UploadID: uploadID,
	}); err != nil {
		s.logger.Errorf("逻辑删除分片记录失败: %v", err)
		span.RecordError(err)
	} else {
		s.logger.Infof("成功逻辑删除分片记录: uploadID=%s", uploadID)
	}

	// 2. 更新文件信息状态为失败并逻辑删除
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx, model.FileInfo{
		UploadID: uploadID,
	})
	if err != nil {
		s.logger.Errorf("获取文件信息失败: %v", err)
		span.RecordError(err)
		return
	}

	// 更新文件状态为失败
	fileInfo.Status = model.FileStatusFailed
	fileInfo.UpdateTime = time.Now()
	fileInfo.Description = fmt.Sprintf("合并失败: %s", reason)

	if err = s.mapper.UpdateFileInfo(ctx, fileInfo); err != nil {
		s.logger.Errorf("更新文件状态为失败状态失败: %v", err)
		span.RecordError(err)
	} else {
		s.logger.Infof("成功更新文件状态为失败: uploadID=%s", uploadID)
	}

	// 3. 逻辑删除文件信息记录
	if err = s.mapper.DeleteFileInfo(ctx, model.FileInfo{
		UploadID: uploadID,
	}); err != nil {
		s.logger.Errorf("逻辑删除文件信息记录失败: %v", err)
		span.RecordError(err)
	} else {
		s.logger.Infof("成功逻辑删除文件信息记录: uploadID=%s", uploadID)
	}

	// 4. 清理临时文件
	tempChunkDir := filepath.Join(s.tempDir, uploadID)
	if err := path.ClearDir(tempChunkDir); err != nil {
		s.logger.Errorf("清理临时文件失败: %v", err)
	} else {
		s.logger.Infof("成功清理临时文件: %s", tempChunkDir)
	}

	s.logger.Warnf("完成失败上传记录清理: uploadID=%s", uploadID)
	span.SetStatus(codes.Ok, "cleanup completed")
}

// calculateFileHash 计算文件的MD5哈希值
//
// @description 计算指定文件的MD5哈希值
// @param filePath string 文件路径
// @return string 文件的MD5哈希值
// @return error 错误信息
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// DownloadFile 下载文件服务
//
// @description 获取文件的下载路径和文件名，用于客户端下载文件
// @param ctx *gin.Context Gin上下文
// @param req *request.DownloadFileRequest 下载文件请求参数
func (s *fileService) DownloadFile(ctx *gin.Context, req *request.DownloadFileRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "downloadFile",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: uploadId=%v", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		span.SetStatus(codes.Error, "文件记录不存在")
		return
	}

	// 获取文件信息
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 检查文件状态
	if fileInfo.Status != model.FileStatusCompleted {
		s.logger.Errorf("file not completed: fileName=%s, status=%v", fileInfo.FileName, fileInfo.Status)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件未完成上传，无法下载")
		span.SetStatus(codes.Error, "文件未完成上传，无法下载")
		return
	}

	// 检查文件是否存在
	if _, err = os.Stat(fileInfo.FilePath); os.IsNotExist(err) {
		s.logger.Errorf("file not found: path=%s", fileInfo.FilePath)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.FileNotFound, "文件不存在")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 更新下载次数
	fileInfo.DownloadCount++
	if err = s.mapper.UpdateFileInfo(ctx, fileInfo); err != nil {
		s.logger.Errorf("update download count failed: %v", err)
		// 不中断流程，继续处理
	}

	commonhttp.DownloadFile(ctx, fileInfo.FilePath, fileInfo.FileName)

	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_name":   fileInfo.FileName,
	// 	"file_size":   fileInfo.FileSize,
	// 	"file_path":   fileInfo.FilePath,
	// 	"file_hash":   fileInfo.FileHash,
	// 	"upload_time": fileInfo.UploadTime,
	// })
	span.SetStatus(codes.Ok, "success")
}

// ListFiles 列出文件服务
//
// @description 分页获取文件列表，支持关键词搜索
// @param ctx *gin.Context Gin上下文
// @param req *request.ListFilesRequest 列出文件请求参数
func (s *fileService) ListFiles(ctx *gin.Context, req *request.ListFilesRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "listFiles",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)
	// 查询文件列表
	pageInfo, err := s.mapper.ListFiles(ctx.Request.Context(), model.FileInfo{
		FileName:   req.FileName,
		UploaderId: req.Uploader,
	}, dbpage.Info[model.FileInfo]{
		PageNo:   int64(req.PageNo),
		PageSize: int64(req.PageSize),
	})

	if err != nil {
		s.logger.Errorf("list files failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	filesResponse, err := response.NewListFilesResponse(pageInfo)
	if err != nil {
		s.logger.Errorf("list files failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, err.Error())
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}

	commonhttp.SuccessResponse(ctx, filesResponse)
	span.SetStatus(codes.Ok, "success")
}

// DeleteFile 删除文件服务
//
// @description 删除指定的文件，包括文件实体和数据库记录
// @param ctx *gin.Context Gin上下文
// @param req *request.DeleteFileRequest 删除文件请求参数
func (s *fileService) DeleteFile(ctx *gin.Context, req *request.DeleteFileRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "deleteFile",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: uploadId=%v", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		span.SetStatus(codes.Error, "文件记录不存在")
		return
	}

	// 获取文件信息
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 检查文件状态
	if fileInfo.Status == model.FileStatusUploading || fileInfo.Status == model.FileStatusPendingMerge {
		s.logger.Errorf("file is uploading or pending merge: uploadId=%v, status=%v", req.UploadID, fileInfo.Status)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件正在上传中，无法删除")
		span.SetStatus(codes.Error, "文件正在上传中，无法删除")
		return
	}

	// 检查文件是否存在
	filePath := fileInfo.FilePath
	if _, err := os.Stat(filePath); err != nil {
		s.logger.Warnf("file not found: path=%s", filePath)
		// 文件不存在不影响删除数据库记录
	}

	// 删除文件
	if !stringer.IsBlank(filePath) {
		if err = path.RemoveFile(filePath); err != nil {
			s.logger.Warnf("delete file failed: %v", err)
		} else {
			s.logger.Infof("delete file success: path=%s", filePath)
		}
	}

	// 删除分片信息
	if err = s.mapper.DeleteFileChunks(ctx.Request.Context(), model.FileChunk{
		UploadID: req.UploadID,
	}); err != nil {
		s.logger.Errorf("delete file chunks failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "删除分片信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 删除文件信息
	if err = s.mapper.DeleteFileInfo(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	}); err != nil {
		s.logger.Errorf("delete file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "删除文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	commonhttp.SuccessResponse(ctx, map[string]interface{}{
		"upload_id": req.UploadID,
		"deleted":   true,
		"file_name": fileInfo.FileName,
	})
	span.SetStatus(codes.Ok, "success")
}

// ResumeUpload 恢复上传服务
//
// @description 根据uploadID恢复上传，获取已上传的分片信息
// @param ctx *gin.Context Gin上下文
// @param req *request.ResumeUploadRequest 恢复上传请求参数
func (s *fileService) ResumeUpload(ctx *gin.Context, req *request.ResumeUploadRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "resumeUpload",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: uploadId=%v", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		span.SetStatus(codes.Error, "文件记录不存在")
		return
	}

	// 根据uploadID查询文件信息
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 检查文件状态
	if fileInfo.Status != model.FileStatusUploading && fileInfo.Status != model.FileStatusPendingMerge {
		s.logger.Errorf("file status incorrect for resume: status=%v", fileInfo.Status)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件状态不正确，无法恢复上传")
		span.SetStatus(codes.Error, "文件状态不正确，无法恢复上传")
		return
	}

	// 获取已上传分片索引
	chunkIndices, err := s.mapper.GetUploadedFileChunkIndexes(ctx.Request.Context(), model.FileChunk{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get uploaded chunk indices failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片索引失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 计算已上传分片数和进度
	uploadedChunks := len(chunkIndices)
	progress := 0.0
	if fileInfo.TotalChunks > 0 {
		progress = float64(uploadedChunks) / fileInfo.TotalChunks * 100
	}

	commonhttp.SuccessResponse(ctx, response.NewResumeUploadResponse(
		fileInfo.FileName,
		fileInfo.FileSize,
		fileInfo.TotalChunks,
		fileInfo.ChunkSize,
		uploadedChunks,
		chunkIndices,
		progress,
		string(fileInfo.Status),
		fileInfo.UploadID,
		fileInfo.UploadTime,
	))
	span.SetStatus(codes.Ok, "success")
}

// GetFileMetadata 获取文件元数据服务
//
// @description 获取指定文件的详细元数据信息
// @param ctx *gin.Context Gin上下文
// @param req *request.GetFileMetadataRequest 获取文件元数据请求参数
func (s *fileService) GetFileMetadata(ctx *gin.Context, req *request.GetFileMetadataRequest) {
	_, span := tracer.StartSpan(ctx.Request.Context(), "fileService", "getFileMetadata",
		attribute.String("requestParam", req.String()),
	)
	defer span.End()
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileInfoExist(ctx.Request.Context(), model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: uploadId=%v", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		span.SetStatus(codes.Error, "文件记录不存在")
		return
	}

	// 获取文件信息
	fileInfo, err := s.mapper.GetFileInfoOneByQuery(ctx.Request.Context(), model.FileInfo{
		UploadID: req.UploadID,
	})
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return
	}

	// 获取文件统计信息
	fileStats := map[string]interface{}{}
	if !stringer.IsBlank(fileInfo.FilePath) {
		if stat, err := os.Stat(fileInfo.FilePath); err == nil {
			fileStats["actual_size"] = stat.Size()
			fileStats["mod_time"] = stat.ModTime()
		}
	}

	commonhttp.SuccessResponse(ctx, map[string]interface{}{
		"upload_id":       fileInfo.UploadID,
		"file_name":       fileInfo.FileName,
		"file_size":       fileInfo.FileSize,
		"file_path":       fileInfo.FilePath,
		"file_hash":       fileInfo.FileHash,
		"file_type":       fileInfo.FileType,
		"uploader":        fileInfo.UploaderId,
		"upload_time":     fileInfo.UploadTime,
		"update_time":     fileInfo.UpdateTime,
		"status":          string(fileInfo.Status),
		"chunk_size":      fileInfo.ChunkSize,
		"total_chunks":    fileInfo.TotalChunks,
		"uploaded_chunks": fileInfo.UploadedChunks,
		"last_chunk_time": fileInfo.LastChunkTime,
		"expire_time":     fileInfo.ExpireTime,
		"download_count":  fileInfo.DownloadCount,
		"description":     fileInfo.Description,
		"storage_type":    fileInfo.StorageType,
		"download_url":    fmt.Sprintf("/files/%s/download", fileInfo.UploadID),
		"file_stats":      fileStats,
	})
	span.SetStatus(codes.Ok, "success")
}
