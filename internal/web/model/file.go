package model

import (
	"database/sql"
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
//
// @description 文件信息表模型
// @struct
type FileInfo struct {
	AutoUid        int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                      // 自增id
	UploadID       string       `json:"upload_id,omitempty" yaml:"upload_id,omitempty" gorm:"column:upload_id;type:varchar(255);default:'';not null"`         // 上传ID
	FileName       string       `json:"file_name,omitempty" yaml:"file_name,omitempty" gorm:"column:file_name;type:text;default:'';"`                         // 文件名
	FileSize       float64      `json:"file_size,omitempty" yaml:"file_size,omitempty" gorm:"column:file_size;type:float;default:0;"`                         // 文件大小
	FilePath       string       `json:"file_path,omitempty" yaml:"file_path,omitempty" gorm:"column:file_path;type:text;default:'';"`                         // 文件路径
	FileHash       string       `json:"file_hash,omitempty" yaml:"file_hash,omitempty" gorm:"column:file_hash;type:varchar(255);default:'';"`                 // 文件hash
	FileType       string       `json:"file_type,omitempty" yaml:"file_type,omitempty" gorm:"column:file_type;type:varchar(255);default:'';"`                 // 文件类型
	UploaderId     string       `json:"uploader_id,omitempty" yaml:"uploader_id,omitempty" gorm:"column:uploader_id;type:varchar(255);default:'';"`           // 上传者ID
	UploadTime     time.Time    `json:"upload_time,omitempty" yaml:"upload_time,omitempty" gorm:"column:upload_time;type:datetime;default:CURRENT_TIMESTAMP"` // 上传时间
	UpdateTime     time.Time    `json:"update_time,omitempty" yaml:"update_time,omitempty" gorm:"column:update_time;type:datetime;default:CURRENT_TIMESTAMP"` // 更新时间
	Status         FileStatus   `json:"status,omitempty" yaml:"status,omitempty" gorm:"column:status;type:varchar(255);default:'';"`                          // 使用FileStatus类型
	ChunkSize      float64      `json:"chunk_size,omitempty" yaml:"chunk_size,omitempty" gorm:"column:chunk_size;type:float;default:0;"`                      // 分片大小
	TotalChunks    float64      `json:"total_chunks,omitempty" yaml:"total_chunks,omitempty" gorm:"column:total_chunks;type:float;default:0;"`                // 总分片数
	UploadedChunks float64      `json:"uploaded_chunks,omitempty" yaml:"uploaded_chunks,omitempty" gorm:"column:uploaded_chunks;type:float;default:0;"`       // 已上传分片数
	LastChunkTime  time.Time    `json:"last_chunk_time,omitempty" yaml:"last_chunk_time,omitempty" gorm:"column:last_chunk_time;type:datetime;"`              // 最后上传分片时间
	ExpireTime     time.Time    `json:"expire_time,omitempty" yaml:"expire_time,omitempty" gorm:"column:expire_time;type:datetime;"`                          // 过期时间
	DownloadCount  float64      `json:"download_count,omitempty" yaml:"download_count,omitempty" gorm:"column:download_count;type:float;default:0;"`          // 下载次数
	Description    string       `json:"description,omitempty" yaml:"description,omitempty" gorm:"column:description;type:text;default:'';"`                   // 文件描述
	StorageType    string       `json:"storage_type,omitempty" yaml:"storage_type,omitempty" gorm:"column:storage_type;type:varchar(255);default:'';"`        // 存储类型
	IpfsCid        string       `json:"ipfs_cid,omitempty" yaml:"ipfs_cid,omitempty" gorm:"column:ipfs_cid;type:varchar(255);default:'';"`                    // ipfs cid
	IsDelete       sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:tinyint(1);default:0;"`                    // 是否删除
	IsRemove       sql.NullBool `json:"is_remove,omitempty" yaml:"is_remove,omitempty" gorm:"column:is_remove;type:tinyint(1);default:0;"`                    // 是否移除
}

// TableName 返回数据库表名
//
// @description 实现gorm接口，指定FileInfo结构体对应的数据库表名
// @return string 数据库表名
func (model FileInfo) TableName() string {
	return fileInfoTableName
}

// String 将文件信息转换为字符串表示
//
// @description 将FileInfo结构体转换为JSON格式的字符串
// @return string 文件信息的JSON格式字符串
func (model FileInfo) String() string {
	str, _ := json.MarshalString(model)
	return str
}

// NewFileInfo 创建新的文件信息实例
//
// @description 初始化一个空的文件信息结构体指针
// @return *FileInfo 文件信息结构体指针
func NewFileInfo() *FileInfo {
	return &FileInfo{}
}

// FileChunk 文件分片模型
//
// @description 文件分片信息
// @struct
const (
	fileChunkTableName = "t_file_chunk"
)

type FileChunk struct {
	AutoUid       int          `json:"auto_uid,omitempty" yaml:"auto_uid,omitempty" gorm:"column:auto_uid;primary_key;auto_increment;"`                        // 自增id
	UploadID      string       `json:"upload_id,omitempty" yaml:"upload_id,omitempty" gorm:"column:upload_id;type:varchar(255);default:'';not null"`           // 上传ID
	ChunkIndex    float64      `json:"chunk_index,omitempty" yaml:"chunk_index,omitempty" gorm:"column:chunk_index;type:float;default:0;not null;"`            // 分片索引
	ChunkSize     float64      `json:"chunk_size,omitempty" yaml:"chunk_size,omitempty" gorm:"column:chunk_size;type:float;default:0;not null;"`               // 分片大小
	CalcChunkSize string       `json:"calc_chunk_size,omitempty" yaml:"calc_chunk_size,omitempty" gorm:"column:calc_chunk_size;type:varchar(255);default:'';"` // 后台计算的hash
	ChunkHash     string       `json:"chunk_hash,omitempty" yaml:"chunk_hash,omitempty" gorm:"column:chunk_hash;type:varchar(255);default:'';"`                // 分片哈希值
	FilePath      string       `json:"file_path,omitempty" yaml:"file_path,omitempty" gorm:"column:file_path;type:text;default:'';not null;"`                  // 分片文件路径
	UploadTime    time.Time    `json:"upload_time,omitempty" yaml:"upload_time,omitempty" gorm:"column:upload_time;type:datetime;default:CURRENT_TIMESTAMP"`   // 上传时间
	Status        sql.NullBool `json:"status,omitempty" yaml:"status,omitempty" gorm:"column:status;type:tinyint(1);default:1;"`                               // 状态 1: upload 0: missing
	IsDelete      sql.NullBool `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:tinyint(1);default:0;"`
}

// String 将文件分片信息转换为字符串表示
//
// @description 将FileChunk结构体转换为JSON格式的字符串
// @return string 文件分片信息的JSON格式字符串
func (model FileChunk) String() string {
	str, _ := json.MarshalString(model)
	return str
}

// TableName 返回数据库表名
//
// @description 实现gorm接口，指定FileChunk结构体对应的数据库表名
// @return string 数据库表名
func (model FileChunk) TableName() string {
	return fileChunkTableName
}

// NewFileChunk 创建新的文件分片信息实例
//
// @description 初始化一个空的文件分片信息结构体指针
// @return *FileChunk 文件分片信息结构体指针
func NewFileChunk() *FileChunk {
	return &FileChunk{}
}
