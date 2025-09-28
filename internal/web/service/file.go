package service

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jianlu8023/go-tools/v2/pkg/path"
	"github.com/jianlu8023/go-tools/v2/pkg/random/uuid"
	"github.com/jianlu8023/golang-example/internal/web/mapper"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/internal/web/request"
	"github.com/jianlu8023/golang-example/internal/web/response"
	commonhttp "github.com/jianlu8023/golang-example/pkg/common/http"
)

// FileService 文件服务
// @description 提供文件相关的服务功能，如文件上传、下载、管理等
// @struct
// @property *Service 基础服务
// @property mapper *mapper.FileMapper 文件映射器
// @property uploadDir string 上传文件保存目录
// @property tempDir string 临时文件保存目录
type FileService struct {
	*Service
	mapper    *mapper.FileMapper
	uploadDir string // 上传文件保存目录
	tempDir   string // 临时文件保存目录
}

// NewFileService 创建文件服务实例
// @description 创建并返回一个新的文件服务实例
// @param baseService *Service 基础服务
// @param fileMapper *mapper.FileMapper 文件映射器
// @return *FileService 文件服务实例
func NewFileService(baseService *Service, fileMapper *mapper.FileMapper) *FileService {
	// 设置上传目录，实际项目中应该从配置中读取
	uploadDir := "uploads"
	tempDir := filepath.Join(uploadDir, "temp")

	// 确保目录存在
	ensureDir(uploadDir)
	ensureDir(tempDir)

	return &FileService{
		Service:   baseService,
		mapper:    fileMapper,
		uploadDir: uploadDir,
		tempDir:   tempDir,
	}
}

// ensureDir 确保目录存在，如果不存在则创建
// @param dir string 目录路径
func ensureDir(dir string) {
	_, _ = path.CreateDir(dir)
}

// InitUpload 初始化文件上传服务
// @description 初始化文件上传，为文件生成唯一ID并创建必要的存储结构
// @param ctx *gin.Context Gin上下文
// @param req *request.FileInitUploadRequest 初始化上传请求参数
func (s *FileService) InitUpload(ctx *gin.Context, req *request.FileInitUploadRequest) {
	s.logger.Debugf("received file init upload request with params: %v", req)

	// 计算总分片数
	totalChunks := math.Ceil(req.FileSize / req.ChunkSize)

	// 生成上传ID
	uploadID := uuid.GetUUID()

	// 先检查文件是否已存在
	fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{UploadID: uploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		return
	}
	if fileExist {
		s.logger.Errorf("file already exists: upload_id=%s", uploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件已存在")
		return
	}

	// 创建文件信息记录
	fileInfo := &model.FileInfo{
		UploadID:       uploadID,
		FileName:       req.FileName,
		FileSize:       req.FileSize,
		FilePath:       "",
		FileHash:       req.FileHash,
		ChunkSize:      req.ChunkSize,
		TotalChunks:    totalChunks,
		UploadedChunks: 0,
		FileType:       req.FileType,
		Description:    req.Description,
		StorageType:    req.StorageType,
		Status:         model.FileStatusUploading,
		UploadTime:     time.Now(),
		UpdateTime:     time.Now(),
	}

	// 保存文件信息
	err = s.mapper.CreateFileInfo(fileInfo)
	if err != nil {
		s.logger.Errorf("save file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "保存文件信息失败")
		return
	}

	// 创建临时目录用于存储分片
	tempChunkDir := filepath.Join(s.tempDir, uploadID)
	ensureDir(tempChunkDir)

	s.logger.Infof("init file upload success, file_id: %d, upload_id: %s", fileInfo.UploadID, uploadID)
	commonhttp.SuccessResponse(ctx, response.NewInitUploadResponse(uploadID, totalChunks, req.ChunkSize, tempChunkDir))
}

// UploadChunk 上传文件分片服务
// @description 上传文件分片，支持大文件的分片上传
// @param ctx *gin.Context Gin上下文
// @param req *request.FileUploadChunkRequest 上传分片请求参数
func (s *FileService) UploadChunk(ctx *gin.Context, req *request.FileUploadChunkRequest) {
	s.logger.Debugf("received upload chunk request with params: %v", req)

	// 先检查文件是否存在
	// fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{AutoUid: req.FileID})
	// if err != nil {
	// 	s.logger.Errorf("check file exist failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
	// 	return
	// }
	// if !fileExist {
	// 	s.logger.Errorf("file not found: file_id=%d", req.FileID)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
	// 	return
	// }

	// 获取文件信息
	// fileInfo, err := s.mapper.GetFileInfoByID(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
	// 	return
	// }

	// 检查文件状态
	// if fileInfo.Status != model.FileStatusUploading {
	// 	s.logger.Errorf("file status incorrect: %v", fileInfo.Status)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件状态不正确，无法上传分片")
	// 	return
	// }

	// 从上下文中获取上传的文件
	file, err := ctx.FormFile("file")
	if err != nil {
		s.logger.Errorf("get uploaded file failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.InvalidParameter, "获取上传文件失败")
		return
	}

	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		s.logger.Errorf("open uploaded file failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "打开上传文件失败")
		return
	}
	defer src.Close()

	// 保存分片文件
	// tempChunkDir := filepath.Join(s.tempDir, fileInfo.UploadID)
	// chunkFilePath := filepath.Join(tempChunkDir, fmt.Sprintf("chunk_%d", req.ChunkIndex))

	// 打开或创建分片文件
	// chunkOutput, err := os.Create(chunkFilePath)
	// if err != nil {
	// 	s.logger.Errorf("create chunk file failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "创建分片文件失败")
	// 	return
	// }
	// defer chunkOutput.Close()

	// 写入分片数据
	// written, err := io.Copy(chunkOutput, src)
	// if err != nil {
	// 	s.logger.Errorf("write chunk data failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "写入分片数据失败")
	// 	return
	// }

	// 获取文件的实际大小
	// chunkSize := written

	// if written != req.ChunkSize && req.ChunkIndex != req.TotalChunks-1 {
	// 	s.logger.Warnf("chunk data incomplete: expected %d bytes, wrote %d bytes", req.ChunkSize, written)
	// }

	// 记录已上传分片
	fileChunk := &model.FileChunk{
		// FileID:     req.FileID,
		// ChunkIndex: req.ChunkIndex,
		// ChunkSize:  chunkSize,
		// FilePath:   chunkFilePath,
		// UploadTime: time.Now(),
	}

	// 检查分片是否已存在
	exists, err := s.mapper.CheckChunkExists(req.FileID, req.ChunkIndex)
	if err != nil {
		s.logger.Errorf("check chunk exists failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查分片是否存在失败")
		return
	}

	if exists {
		// 更新分片信息
		if err := s.mapper.UpdateFileChunk(fileChunk); err != nil {
			s.logger.Errorf("update file chunk failed: %v", err)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "更新分片信息失败")
			return
		}
	} else {
		// 插入新分片信息
		if _, err := s.mapper.CreateFileChunk(fileChunk); err != nil {
			s.logger.Errorf("save file chunk failed: %v", err)
			commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "保存分片信息失败")
			return
		}
	}

	// 更新已上传分片数
	// uploadedChunks, err := s.mapper.GetUploadedChunkCount(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get uploaded chunk count failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片数失败")
	// 	return
	// }

	// 更新文件信息
	// fileInfo.UploadedChunks = uploadedChunks
	// fileInfo.UpdateTime = time.Now()

	// if uploadedChunks >= fileInfo.TotalChunks {
	// 	fileInfo.Status = model.FileStatusPendingMerge // 等待合并
	// }

	// if err := s.mapper.UpdateFileInfo(fileInfo); err != nil {
	// 	s.logger.Errorf("update file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "更新文件信息失败")
	// 	return
	// }

	// 计算上传进度
	// progress := 0.0
	// if fileInfo.TotalChunks > 0 {
	// 	progress = float64(uploadedChunks) / float64(fileInfo.TotalChunks) * 100
	// }

	// s.logger.Infof("upload chunk success, file_id: %d, chunk_index: %d, progress: %.2f%%",
	// 	req.FileID, req.ChunkIndex, progress)

	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_id":         req.FileID,
	// 	"chunk_index":     req.ChunkIndex,
	// 	"uploaded_chunks": uploadedChunks,
	// 	"total_chunks":    fileInfo.TotalChunks,
	// 	"progress":        progress,
	// })
}

// CompleteUpload 完成文件上传服务
// @description 合并所有分片文件，完成文件上传过程
// @param ctx *gin.Context Gin上下文
// @param req *request.CompleteUploadRequest 完成上传请求参数
func (s *FileService) CompleteUpload(ctx *gin.Context, req *request.CompleteUploadRequest) {
	s.logger.Debugf("received complete upload request: file_id=%d", req.FileID)

	// 先检查文件是否存在
	// fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{AutoUid: req.FileID})
	// if err != nil {
	// 	s.logger.Errorf("check file exist failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
	// 	return
	// }
	// if !fileExist {
	// 	s.logger.Errorf("file not found: file_id=%d", req.FileID)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
	// 	return
	// }

	// 获取文件信息
	// fileInfo, err := s.mapper.GetFileInfoByID(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
	// 	return
	// }

	// 检查文件状态
	// if fileInfo.Status != model.FileStatusUploading && fileInfo.Status != model.FileStatusPendingMerge {
	// 	s.logger.Errorf("file status incorrect: %v", fileInfo.Status)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件状态不正确，无法完成上传")
	// 	return
	// }

	// 检查所有分片是否都已上传
	// uploadedChunks, err := s.mapper.GetUploadedChunkCount(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get uploaded chunk count failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片数失败")
	// 	return
	// }

	// if uploadedChunks < fileInfo.TotalChunks {
	// 	s.logger.Errorf("not all chunks uploaded: %d/%d", uploadedChunks, fileInfo.TotalChunks)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError,
	// 		fmt.Sprintf("分片未全部上传：已上传 %d 个，共 %d 个", uploadedChunks, fileInfo.TotalChunks))
	// 	return
	// }

	// 获取所有分片信息
	// chunks, err := s.mapper.GetFileChunks(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get file chunks failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取分片信息失败")
	// 	return
	// }

	// 按分片索引排序
	// sort.Slice(chunks, func(i, j int) bool {
	// 	return chunks[i].ChunkIndex < chunks[j].ChunkIndex
	// })

	// 创建最终文件
	// fileExt := filepath.Ext(fileInfo.FileName)
	// fileNameWithoutExt := fileInfo.FileName[:len(fileInfo.FileName)-len(fileExt)]
	// destFileName := fmt.Sprintf("%s_%s%s", fileNameWithoutExt, time.Now().Format("20060102_150405"), fileExt)
	// destFilePath := filepath.Join(s.uploadDir, destFileName)

	// 打开目标文件
	// destFile, err := os.Create(destFilePath)
	// if err != nil {
	// 	s.logger.Errorf("create destination file failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "创建目标文件失败")
	// 	return
	// }
	// defer destFile.Close()

	// 合并分片
	// for _, chunk := range chunks {
	// 	// 打开分片文件
	// 	chunkFile, err := os.Open(chunk.FilePath)
	// 	if err != nil {
	// 		s.logger.Errorf("open chunk file failed: %v", err)
	// 		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "打开分片文件失败")
	// 		return
	// 	}
	//
	// 	// 写入目标文件
	// 	size, err := io.Copy(destFile, chunkFile)
	// 	chunkFile.Close()
	//
	// 	if err != nil {
	// 		s.logger.Errorf("merge chunk data failed: %v", err)
	// 		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed, "合并分片数据失败")
	// 		return
	// 	}
	//
	// 	if size != chunk.ChunkSize && chunk.ChunkIndex != chunks[len(chunks)-1].ChunkIndex {
	// 		s.logger.Errorf("chunk data incomplete: expected %d bytes, got %d bytes for chunk %d", chunk.ChunkSize, size, chunk.ChunkIndex)
	// 		commonhttp.FailedResponseWithMessage(ctx, commonhttp.NormalFailed,
	// 			fmt.Sprintf("合并分片数据不完整：分片 %d 期望 %d 字节，实际读取 %d 字节",
	// 				chunk.ChunkIndex, chunk.ChunkSize, size))
	// 		return
	// 	}
	// }

	// 计算文件哈希值
	// fileHash, err := calculateFileHash(destFilePath)
	// if err != nil {
	// 	s.logger.Warnf("calculate file hash failed: %v", err)
	// 	// 不中断流程，继续处理
	// }
	//
	// // 更新文件信息
	// fileInfo.Status = model.FileStatusCompleted
	// fileInfo.FilePath = destFilePath
	// fileInfo.FileHash = fileHash
	// fileInfo.UpdateTime = time.Now()
	//
	// if err := s.mapper.UpdateFileInfo(fileInfo); err != nil {
	// 	s.logger.Errorf("update file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "更新文件信息失败")
	// 	return
	// }
	//
	// // 清理临时文件
	// go func() {
	// 	tempChunkDir := filepath.Join(s.tempDir, fileInfo.UploadID)
	// 	os.RemoveAll(tempChunkDir)
	// }()
	//
	// // 生成下载URL
	// downloadURL := fmt.Sprintf("/api/v1/files/%d/download", req.FileID)
	//
	// s.logger.Infof("complete file upload success: file_id=%d, file_name=%s",
	// 	req.FileID, fileInfo.FileName)
	//
	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_id":       req.FileID,
	// 	"file_name":     fileInfo.FileName,
	// 	"file_size":     fileInfo.FileSize,
	// 	"file_hash":     fileHash,
	// 	"file_path":     destFilePath,
	// 	"download_url":  downloadURL,
	// 	"upload_time":   fileInfo.UploadTime,
	// 	"complete_time": fileInfo.UpdateTime,
	// })
}

// GetUploadStatus 获取文件上传状态服务
// @description 获取指定文件的上传进度、已上传分片信息和当前状态
// @param ctx *gin.Context Gin上下文
// @param req *request.GetUploadStatusRequest 获取上传状态请求参数
func (s *FileService) GetUploadStatus(ctx *gin.Context, req *request.GetUploadStatusRequest) {
	s.logger.Debugf("received get upload status request: file_id=%d", req.FileID)

	// 先检查文件是否存在
	// fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{AutoUid: req.FileID})
	// if err != nil {
	// 	s.logger.Errorf("check file exist failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
	// 	return
	// }
	// if !fileExist {
	// 	s.logger.Errorf("file not found: file_id=%d", req.FileID)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
	// 	return
	// }

	// 获取文件信息
	// fileInfo, err := s.mapper.GetFileInfoByID(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
	// 	return
	// }

	// 获取已上传分片数
	// uploadedChunks, err := s.mapper.GetUploadedChunkCount(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get uploaded chunk count failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片数失败")
	// 	return
	// }

	// 获取已上传分片索引
	// chunkIndices, err := s.mapper.GetUploadedChunkIndexes(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get uploaded chunk indices failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片索引失败")
	// 	return
	// }

	// 计算上传进度
	// progress := 0.0
	// if fileInfo.TotalChunks > 0 {
	// 	progress = float64(uploadedChunks) / float64(fileInfo.TotalChunks) * 100
	// }

	// s.logger.Infof("get upload status success: file_id=%d, progress=%.2f%%", req.FileID, progress)
	//
	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_id":         req.FileID,
	// 	"file_name":       fileInfo.FileName,
	// 	"file_size":       fileInfo.FileSize,
	// 	"total_chunks":    fileInfo.TotalChunks,
	// 	"uploaded_chunks": uploadedChunks,
	// 	"chunk_indices":   chunkIndices,
	// 	"progress":        progress,
	// 	"status":          string(fileInfo.Status),
	// 	"upload_id":       fileInfo.UploadID,
	// 	"create_time":     fileInfo.UploadTime,
	// })
}

// calculateFileHash 计算文件的MD5哈希值
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
// @description 获取文件的下载路径和文件名，用于客户端下载文件
// @param ctx *gin.Context Gin上下文
// @param req *request.DownloadFileRequest 下载文件请求参数
func (s *FileService) DownloadFile(ctx *gin.Context, req *request.DownloadFileRequest) {
	s.logger.Debugf("received download file request: file_id=%d", req.FileID)

	// 先检查文件是否存在
	// fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{AutoUid: req.FileID})
	// if err != nil {
	// 	s.logger.Errorf("check file exist failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
	// 	return
	// }
	// if !fileExist {
	// 	s.logger.Errorf("file not found: file_id=%d", req.FileID)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
	// 	return
	// }

	// 获取文件信息
	// fileInfo, err := s.mapper.GetFileInfoByID(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
	// 	return
	// }

	// 检查文件状态
	// if fileInfo.Status != model.FileStatusCompleted {
	// 	s.logger.Errorf("file not completed: file_id=%d, status=%v", req.FileID, fileInfo.Status)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件未完成上传，无法下载")
	// 	return
	// }

	// 检查文件是否存在
	// if _, err := os.Stat(fileInfo.FilePath); os.IsNotExist(err) {
	// 	s.logger.Errorf("file not found: path=%s", fileInfo.FilePath)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.FileNotFound, "文件不存在")
	// 	return
	// }

	// 更新下载次数
	// fileInfo.DownloadCount++
	// if err := s.mapper.UpdateFileInfo(fileInfo); err != nil {
	// 	s.logger.Errorf("update download count failed: %v", err)
	// 	// 不中断流程，继续处理
	// }

	// s.logger.Infof("download file success: file_id=%d, file_name=%s", req.FileID, fileInfo.FileName)

	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_id":      req.FileID,
	// 	"file_name":    fileInfo.FileName,
	// 	"file_size":    fileInfo.FileSize,
	// 	"file_path":    fileInfo.FilePath,
	// 	"file_hash":    fileInfo.FileHash,
	// 	"upload_time":  fileInfo.UploadTime,
	// 	"download_url": fmt.Sprintf("/api/v1/files/%d/download", req.FileID),
	// })
}

// ListFiles 列出文件服务
// @description 分页获取文件列表，支持关键词搜索
// @param ctx *gin.Context Gin上下文
// @param req *request.ListFilesRequest 列出文件请求参数
func (s *FileService) ListFiles(ctx *gin.Context, req *request.ListFilesRequest) {
	s.logger.Debugf("received list files request: page=%d, page_size=%d, file_name=%s",
		req.Page, req.PageSize, req.FileName)

	// 计算偏移量
	// offset := (req.Page - 1) * req.PageSize

	// 查询文件列表
	// files, total, err := s.mapper.ListFiles(req.FileName, req.Status, req.Uploader, offset, req.PageSize, req.OrderBy, req.OrderType)
	// if err != nil {
	// 	s.logger.Errorf("get file list failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件列表失败")
	// 	return
	// }

	// 转换为响应格式
	// fileList := make([]map[string]interface{}, 0, len(files))
	// for _, file := range files {
	// 	fileItem := map[string]interface{}{
	// 		"file_id":        file.AutoUid,
	// 		"file_name":      file.FileName,
	// 		"file_size":      file.FileSize,
	// 		"file_path":      file.FilePath,
	// 		"file_hash":      file.FileHash,
	// 		"file_type":      file.FileType,
	// 		"uploader":       file.Uploader,
	// 		"upload_time":    file.UploadTime,
	// 		"status":         string(file.Status),
	// 		"download_count": file.DownloadCount,
	// 		"description":    file.Description,
	// 		"storage_type":   file.StorageType,
	// 		"expire_time":    file.ExpireTime,
	// 		"download_url":   fmt.Sprintf("/api/v1/files/%d/download", file.AutoUid),
	// 		"extra":          file.Extra,
	// 	}
	// 	fileList = append(fileList, fileItem)
	// }

	// 计算总页数
	// totalPages := 0
	// if total > 0 {
	// 	totalPages = int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	// }

	// s.logger.Infof("list files success: page=%d, page_size=%d, total=%d",
	// 	req.Page, req.PageSize, total)
	//
	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"files":       fileList,
	// 	"total":       total,
	// 	"page":        req.Page,
	// 	"page_size":   req.PageSize,
	// 	"total_pages": totalPages,
	// })
}

// DeleteFile 删除文件服务
// @description 删除指定的文件，包括文件实体和数据库记录
// @param ctx *gin.Context Gin上下文
// @param req *request.DeleteFileRequest 删除文件请求参数
func (s *FileService) DeleteFile(ctx *gin.Context, req *request.DeleteFileRequest) {
	s.logger.Debugf("received delete file request: file_id=%d", req.FileID)

	// 先检查文件是否存在
	// fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{AutoUid: req.FileID})
	// if err != nil {
	// 	s.logger.Errorf("check file exist failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
	// 	return
	// }
	// if !fileExist {
	// 	s.logger.Errorf("file not found: file_id=%d", req.FileID)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
	// 	return
	// }

	// 获取文件信息
	// fileInfo, err := s.mapper.GetFileInfoByID(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
	// 	return
	// }

	// 检查文件状态
	// if fileInfo.Status == model.FileStatusUploading || fileInfo.Status == model.FileStatusPendingMerge {
	// 	s.logger.Errorf("file is uploading or pending merge: file_id=%d, status=%v", req.FileID, fileInfo.Status)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件正在上传中，无法删除")
	// 	return
	// }

	// 检查文件是否存在
	// filePath := fileInfo.FilePath
	// if _, err := os.Stat(filePath); err != nil {
	// 	s.logger.Warnf("file not found: path=%s", filePath)
	// 	// 文件不存在不影响删除数据库记录
	// }

	// 删除文件
	// if filePath != "" {
	// 	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
	// 		s.logger.Warnf("delete file failed: %v", err)
	// 		// 文件删除失败不影响删除数据库记录
	// 	} else if err == nil {
	// 		s.logger.Infof("delete file success: path=%s", filePath)
	// 	}
	// }

	// 删除分片信息
	// if err := s.mapper.DeleteFileChunks(req.FileID); err != nil {
	// 	s.logger.Errorf("delete file chunks failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "删除分片信息失败")
	// 	return
	// }

	// 删除文件信息
	// if err := s.mapper.DeleteFileInfo(req.FileID); err != nil {
	// 	s.logger.Errorf("delete file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "删除文件信息失败")
	// 	return
	// }

	// s.logger.Infof("delete file success: file_id=%d, file_name=%s", req.FileID, fileInfo.FileName)
	//
	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_id":   req.FileID,
	// 	"deleted":   true,
	// 	"file_name": fileInfo.FileName,
	// })
}

// ResumeUpload 恢复上传服务
// @description 根据uploadID恢复上传，获取已上传的分片信息
// @param ctx *gin.Context Gin上下文
// @param req *request.ResumeUploadRequest 恢复上传请求参数
func (s *FileService) ResumeUpload(ctx *gin.Context, req *request.ResumeUploadRequest) {
	s.logger.Debugf("received resume upload request: upload_id=%s", req.UploadID)

	// 先检查文件是否存在
	fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{UploadID: req.UploadID})
	if err != nil {
		s.logger.Errorf("check file exist failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
		return
	}
	if !fileExist {
		s.logger.Errorf("file not found: upload_id=%s", req.UploadID)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
		return
	}

	// 根据uploadID查询文件信息
	fileInfo, err := s.mapper.GetFileInfoByUploadID(req.UploadID)
	if err != nil {
		s.logger.Errorf("get file info failed: %v", err)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
		return
	}

	// 检查文件状态
	if fileInfo.Status != model.FileStatusUploading && fileInfo.Status != model.FileStatusPendingMerge {
		s.logger.Errorf("file status incorrect for resume: status=%v", fileInfo.Status)
		commonhttp.FailedResponseWithMessage(ctx, commonhttp.BusinessLogicError, "文件状态不正确，无法恢复上传")
		return
	}

	// 获取已上传分片索引
	// chunkIndices, err := s.mapper.GetUploadedChunkIndexes(fileInfo.ID)
	// if err != nil {
	// 	s.logger.Errorf("get uploaded chunk indices failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取已上传分片索引失败")
	// 	return
	// }

	// 计算已上传分片数和进度
	// uploadedChunks := len(chunkIndices)
	// progress := 0.0
	// if fileInfo.TotalChunks > 0 {
	// 	progress = float64(uploadedChunks) / float64(fileInfo.TotalChunks) * 100
	// }
	//
	// s.logger.Infof("resume upload success: upload_id=%s, file_id=%d, progress=%.2f%%",
	// 	req.UploadID, fileInfo.ID, progress)

	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_id":         fileInfo.AutoUid,
	// 	"file_name":       fileInfo.FileName,
	// 	"file_size":       fileInfo.FileSize,
	// 	"total_chunks":    fileInfo.TotalChunks,
	// 	"chunk_size":      fileInfo.ChunkSize,
	// 	"uploaded_chunks": uploadedChunks,
	// 	"chunk_indices":   chunkIndices,
	// 	"progress":        progress,
	// 	"status":          string(fileInfo.Status),
	// 	"upload_id":       req.UploadID,
	// 	"create_time":     fileInfo.UploadTime,
	// })
}

// GetFileMetadata 获取文件元数据服务
// @description 获取指定文件的详细元数据信息
// @param ctx *gin.Context Gin上下文
// @param req *request.GetFileMetadataRequest 获取文件元数据请求参数
func (s *FileService) GetFileMetadata(ctx *gin.Context, req *request.GetFileMetadataRequest) {
	s.logger.Debugf("received get file metadata request: file_id=%d", req.FileID)

	// 先检查文件是否存在
	// fileExist, err := s.mapper.QueryFileExist(&model.FileInfo{AutoUid: req.FileID})
	// if err != nil {
	// 	s.logger.Errorf("check file exist failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "检查文件是否存在失败")
	// 	return
	// }
	// if !fileExist {
	// 	s.logger.Errorf("file not found: file_id=%d", req.FileID)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.RecordNotFound, "文件不存在")
	// 	return
	// }

	// 获取文件信息
	// fileInfo, err := s.mapper.GetFileInfoByID(req.FileID)
	// if err != nil {
	// 	s.logger.Errorf("get file info failed: %v", err)
	// 	commonhttp.FailedResponseWithMessage(ctx, commonhttp.DatabaseError, "获取文件信息失败")
	// 	return
	// }

	// 获取文件统计信息
	// fileStats := map[string]interface{}{}
	// if fileInfo.FilePath != "" {
	// 	if stat, err := os.Stat(fileInfo.FilePath); err == nil {
	// 		fileStats["actual_size"] = stat.Size()
	// 		fileStats["mod_time"] = stat.ModTime()
	// 	}
	// }

	// s.logger.Infof("get file metadata success: file_id=%d, file_name=%s", req.FileID, fileInfo.FileName)

	// commonhttp.SuccessResponse(ctx, map[string]interface{}{
	// 	"file_id":         fileInfo.AutoUid,
	// 	"upload_id":       fileInfo.UploadID,
	// 	"file_name":       fileInfo.FileName,
	// 	"file_size":       fileInfo.FileSize,
	// 	"file_path":       fileInfo.FilePath,
	// 	"file_hash":       fileInfo.FileHash,
	// 	"file_type":       fileInfo.FileType,
	// 	"uploader":        fileInfo.Uploader,
	// 	"upload_time":     fileInfo.UploadTime,
	// 	"update_time":     fileInfo.UpdateTime,
	// 	"status":          string(fileInfo.Status),
	// 	"chunk_size":      fileInfo.ChunkSize,
	// 	"total_chunks":    fileInfo.TotalChunks,
	// 	"uploaded_chunks": fileInfo.UploadedChunks,
	// 	"last_chunk_time": fileInfo.LastChunkTime,
	// 	"expire_time":     fileInfo.ExpireTime,
	// 	"download_count":  fileInfo.DownloadCount,
	// 	"description":     fileInfo.Description,
	// 	"storage_type":    fileInfo.StorageType,
	// 	"download_url":    fmt.Sprintf("/api/v1/files/%d/download", fileInfo.AutoUid),
	// 	"extra":           fileInfo.Extra,
	// 	"file_stats":      fileStats,
	// })
}
