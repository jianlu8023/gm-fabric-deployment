package request

import (
	"github.com/jianlu8023/go-tools/v2/pkg/json"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
)

// FileInitUploadRequest 初始化文件上传请求
// @description 用于初始化文件上传，获取上传ID和分片信息
// @struct
// @property FileName string 文件名
// @property FileSize int64 文件大小
// @property FileHash string 文件哈希值
// @property ChunkSize int64 分片大小
// @property FileType string 文件类型
// @property Description string 文件描述
// @property StorageType string 存储类型
// @property ExpireTime int64 过期时间戳(秒)
type FileInitUploadRequest struct {
	FileName    string `form:"file_name" binding:"required,max=255" validate:"max=255"`
	FileSize    int64  `form:"file_size" binding:"required,gte=0" validate:"gte=0"`
	FileHash    string `form:"file_hash" binding:"omitempty,max=255" validate:"max=255"`
	ChunkSize   int64  `form:"chunk_size" binding:"required,gt=0,lte=10485760" validate:"gt=0,lte=10485760"` // 最大10MB
	FileType    string `form:"file_type" binding:"omitempty,max=100" validate:"max=100"`
	Description string `form:"description" binding:"omitempty,max=1000" validate:"max=1000"`
	StorageType string `form:"storage_type" binding:"omitempty,max=50" validate:"max=50"`
	ExpireTime  int64  `form:"expire_time" binding:"omitempty,gte=0" validate:"gte=0"`
}

func (req FileInitUploadRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// IsLegal 验证初始化上传请求参数是否合法
// @description 检查文件名、文件大小、分片大小等参数的合法性
// @return bool 参数是否合法
func (req FileInitUploadRequest) IsLegal() bool {
	// 文件名不能为空
	if stringer.IsBlank(req.FileName) {
		return false
	}
	// 文件大小必须大于0
	if req.FileSize <= 0 {
		return false
	}
	// 分片大小必须在合理范围内 (1KB - 10MB)
	if req.ChunkSize <= 0 || req.ChunkSize > 10*1024*1024 {
		return false
	}
	// 分片大小不能大于文件大小
	if req.ChunkSize > req.FileSize {
		return false
	}
	return true
}

// FileUploadChunkRequest 上传文件分片请求
// @description 用于上传文件分片
// @struct
// @property FileID int64 文件ID
// @property ChunkIndex int 分片索引(从0开始)
// @property ChunkHash string 分片哈希值
// @property TotalChunks int 总分片数
// @property File multipart.FileHeader 分片文件数据
type FileUploadChunkRequest struct {
	FileID      int64  `form:"file_id" binding:"required,gt=0" validate:"gt=0"`
	ChunkIndex  int    `form:"chunk_index" binding:"required,gte=0" validate:"gte=0"`
	ChunkHash   string `form:"chunk_hash" binding:"omitempty,max=255" validate:"max=255"`
	TotalChunks int    `form:"total_chunks" binding:"required,gt=0" validate:"gt=0"`
	// File字段用于接收文件数据，在handler中通过ctx.FormFile("file")获取
}

// IsLegal 验证上传分片请求参数是否合法
// @description 检查文件ID、分片索引、总分片数等参数的合法性
// @return bool 参数是否合法
func (req FileUploadChunkRequest) IsLegal() bool {
	// 文件ID必须大于0
	if req.FileID <= 0 {
		return false
	}
	// 分片索引必须大于等于0
	if req.ChunkIndex < 0 {
		return false
	}
	// 总分片数必须大于0
	if req.TotalChunks <= 0 {
		return false
	}
	// 分片索引不能大于等于总分片数
	if req.ChunkIndex >= req.TotalChunks {
		return false
	}
	return true
}
func (req FileUploadChunkRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// CompleteUploadRequest 完成文件上传请求
// @description 用于通知服务器文件分片上传完成，进行文件合并
// @struct
// @property FileID int64 文件ID
// @property FileHash string 文件哈希值
// @property Uploader string 上传者
type CompleteUploadRequest struct {
	FileID   int64  `form:"file_id" binding:"required,gt=0" validate:"gt=0"`
	FileHash string `form:"file_hash" binding:"omitempty,max=255" validate:"max=255"`
	Uploader string `form:"uploader" binding:"omitempty,max=100" validate:"max=100"`
}

// IsLegal 验证完成上传请求参数是否合法
// @description 检查文件ID等参数的合法性
// @return bool 参数是否合法
func (req CompleteUploadRequest) IsLegal() bool {
	// 文件ID必须大于0
	if req.FileID <= 0 {
		return false
	}
	return true
}
func (req CompleteUploadRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// GetUploadStatusRequest 查询文件上传状态请求
// @description 用于查询文件上传状态，获取已上传分片信息，支持断点续传
// @struct
// @property FileID int64 文件ID
type GetUploadStatusRequest struct {
	FileID int64 `form:"file_id" binding:"required,gt=0" validate:"gt=0"`
}

// IsLegal 验证获取上传状态请求参数是否合法
// @description 检查文件ID等参数的合法性
// @return bool 参数是否合法
func (req GetUploadStatusRequest) IsLegal() bool {
	// 文件ID必须大于0
	if req.FileID <= 0 {
		return false
	}
	return true
}
func (req GetUploadStatusRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// DownloadFileRequest 下载文件请求
// @description 用于下载文件
// @struct
// @property FileID int64 文件ID
type DownloadFileRequest struct {
	FileID int64 `form:"file_id" binding:"required,gt=0" validate:"gt=0"`
}

// IsLegal 验证下载文件请求参数是否合法
// @description 检查文件ID等参数的合法性
// @return bool 参数是否合法
func (req DownloadFileRequest) IsLegal() bool {
	// 文件ID必须大于0
	if req.FileID <= 0 {
		return false
	}
	return true
}
func (req DownloadFileRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// ListFilesRequest 查询文件列表请求
// @description 用于查询文件列表
// @struct
// @property Page int 页码
// @property PageSize int 每页大小
// @property Status string 文件状态
// @property FileName string 文件名(模糊查询)
// @property Uploader string 上传者
// @property OrderBy string 排序字段
// @property OrderType string 排序类型(asc/desc)
type ListFilesRequest struct {
	Page      int    `form:"page" binding:"omitempty,gte=1" validate:"gte=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,gte=1,lte=100" validate:"gte=1,lte=100"`
	Status    string `form:"status" binding:"omitempty,max=20" validate:"max=20"`
	FileName  string `form:"file_name" binding:"omitempty,max=255" validate:"max=255"`
	Uploader  string `form:"uploader" binding:"omitempty,max=100" validate:"max=100"`
	OrderBy   string `form:"order_by" binding:"omitempty,max=50" validate:"max=50"`
	OrderType string `form:"order_type" binding:"omitempty,eq=asc|eq=desc" validate:"eq=asc|eq=desc"`
}

// IsLegal 验证文件列表请求参数是否合法
// @description 检查分页参数、排序参数等的合法性
// @return bool 参数是否合法
func (req ListFilesRequest) IsLegal() bool {
	// 页码必须大于0
	if req.Page <= 0 {
		req.Page = 1
	}
	// 页面大小必须在合理范围内
	if req.PageSize <= 0 {
		req.PageSize = 10
	} else if req.PageSize > 100 {
		req.PageSize = 100
	}
	// 排序类型只能是asc或desc
	if req.OrderType != "" && req.OrderType != "asc" && req.OrderType != "desc" {
		return false
	}
	return true
}
func (req ListFilesRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// DeleteFileRequest 删除文件请求
// @description 用于删除指定文件
// @struct
// @property FileID int64 文件ID
type DeleteFileRequest struct {
	FileID int64 `form:"file_id" binding:"required,gt=0" validate:"gt=0"`
}

// IsLegal 验证删除文件请求参数是否合法
// @description 检查文件ID等参数的合法性
// @return bool 参数是否合法
func (req DeleteFileRequest) IsLegal() bool {
	// 文件ID必须大于0
	if req.FileID <= 0 {
		return false
	}
	return true
}
func (req DeleteFileRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// GetFileMetadataRequest 获取文件元数据请求
// @description 用于获取文件的详细元数据信息
// @struct
// @property FileID int64 文件ID
type GetFileMetadataRequest struct {
	FileID int64 `form:"file_id" binding:"required,gt=0" validate:"gt=0"`
}

// IsLegal 验证获取文件元数据请求参数是否合法
// @description 检查文件ID等参数的合法性
// @return bool 参数是否合法
func (req GetFileMetadataRequest) IsLegal() bool {
	// 文件ID必须大于0
	if req.FileID <= 0 {
		return false
	}
	return true
}

func (req GetFileMetadataRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}

// ResumeUploadRequest 恢复上传请求
// @description 用于恢复中断的文件上传
// @struct
// @property UploadID string 上传ID
type ResumeUploadRequest struct {
	UploadID string `form:"upload_id" binding:"required" validate:"required"`
}

// IsLegal 验证恢复上传请求参数是否合法
// @description 检查上传ID等参数的合法性
// @return bool 参数是否合法
func (req ResumeUploadRequest) IsLegal() bool {
	// 上传ID不能为空
	if stringer.IsBlank(req.UploadID) {
		return false
	}
	return true
}

func (req ResumeUploadRequest) String() string {
	str, _ := json.MarshalString(req)
	return str
}
