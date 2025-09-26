package model

import (
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
)

// FileStatus 文件状态常量
type FileStatus string

const (
	// FileStatusUploading 文件正在上传
	FileStatusUploading FileStatus = "uploading"
	// FileStatusPendingMerge 文件等待合并
	FileStatusPendingMerge FileStatus = "pending_merge"
	// FileStatusCompleted 文件上传完成
	FileStatusCompleted FileStatus = "completed"
	// FileStatusFailed 文件上传失败
	FileStatusFailed FileStatus = "failed"
)

const (
	fileInfoTableName = "t_file_info"
)

// FileInfo 文件信息模型
// @description 文件信息表模型
// @struct
// @property ID int 文件ID
// @property FileName string 文件名
// @property FileSize int64 文件大小
// @property FilePath string 文件路径
// @property FileHash string 文件哈希值
// @property FileType string 文件类型
// @property Uploader string 上传者
// @property UploadTime time.Time 上传时间
// @property Status string 文件状态
// @property ChunkSize int64 分片大小
// @property TotalChunks int 总分片数
// @property UploadedChunks int 已上传分片数
// @property LastChunkTime time.Time 最后分片上传时间
// @property ExpireTime time.Time 过期时间
// @property DownloadCount int 下载次数
// @property Description string 文件描述
// @property StorageType string 存储类型
// @property Extra map[string]interface{} 额外信息
type FileInfo struct {
	ID             int64                  `json:"id" db:"id"`
	UploadID       string                 `json:"upload_id" db:"upload_id"` // 新增：上传ID
	FileName       string                 `json:"file_name" db:"file_name"`
	FileSize       int64                  `json:"file_size" db:"file_size"`
	FilePath       string                 `json:"file_path" db:"file_path"`
	FileHash       string                 `json:"file_hash" db:"file_hash"`
	FileType       string                 `json:"file_type" db:"file_type"`
	Uploader       string                 `json:"uploader" db:"uploader"`
	UploadTime     time.Time              `json:"upload_time" db:"upload_time"`
	UpdateTime     time.Time              `json:"update_time" db:"update_time"` // 新增：更新时间
	Status         FileStatus             `json:"status" db:"status"`           // 使用FileStatus类型
	ChunkSize      int64                  `json:"chunk_size" db:"chunk_size"`
	TotalChunks    int                    `json:"total_chunks" db:"total_chunks"`
	UploadedChunks int                    `json:"uploaded_chunks" db:"uploaded_chunks"`
	LastChunkTime  time.Time              `json:"last_chunk_time" db:"last_chunk_time"`
	ExpireTime     time.Time              `json:"expire_time" db:"expire_time"`
	DownloadCount  int                    `json:"download_count" db:"download_count"`
	Description    string                 `json:"description" db:"description"`
	StorageType    string                 `json:"storage_type" db:"storage_type"`
	Extra          map[string]interface{} `json:"extra" db:"extra"`
}

func (model FileInfo) TableName() string {
	return fileInfoTableName
}

func (model FileInfo) String() string {
	str, _ := json.MarshalString(model)
	return str
}

func NewFileInfo() *FileInfo {
	return &FileInfo{}
}

// FileChunk 文件分片模型
// @description 文件分片信息
// @struct
// @property FileID int64 文件ID
// @property ChunkIndex int 分片索引
// @property ChunkSize int64 分片大小
// @property ChunkHash string 分片哈希值
// @property UploadTime time.Time 上传时间
// @property Status string 分片状态
const (
	fileChunkTableName = "t_file_chunk"
)

type FileChunk struct {
	FileID     int64     `json:"file_id" db:"file_id"`
	ChunkIndex int       `json:"chunk_index" db:"chunk_index"`
	ChunkSize  int64     `json:"chunk_size" db:"chunk_size"`
	ChunkHash  string    `json:"chunk_hash" db:"chunk_hash"`
	FilePath   string    `json:"file_path" db:"file_path"` // 新增：分片文件路径
	UploadTime time.Time `json:"upload_time" db:"upload_time"`
	Status     string    `json:"status" db:"status"` // "uploaded", "missing"
}

func (model FileChunk) TableName() string {
	return fileChunkTableName
}

func NewFileChunk() *FileChunk {
	return &FileChunk{}
}
