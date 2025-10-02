package response

import (
	"database/sql"
	"time"

	"github.com/jianlu8023/go-tools/v2/pkg/json"
	humantime "github.com/jianlu8023/go-tools/v2/pkg/time"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"github.com/jinzhu/copier"
)

// InitUploadResponse 初始化文件上传响应
// @description 初始化文件上传后的响应
// @struct
// @property FileID int64 文件ID
// @property UploadID string 上传ID
// @property TotalChunks int 总分片数
// @property ChunkSize int64 分片大小
type InitUploadResponse struct {
	UploadID    string  `json:"upload_id" yaml:"upload_id"`       // 上传ID
	TotalChunks float64 `json:"total_chunks" yaml:"total_chunks"` // 总分片数
	ChunkSize   float64 `json:"chunk_size" yaml:"chunk_size"`     // 分片大小
}

func (resp InitUploadResponse) MarshalJSON() ([]byte, error) {
	type Alias InitUploadResponse
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&resp),
	}
	return json.Marshal(aux)
}

// String 将InitUploadResponse转换为字符串
// @description 将InitUploadResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp InitUploadResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewInitUploadResponse 创建初始化上传响应对象
// @description 创建一个新的初始化文件上传响应对象
// @param fileID int64 文件ID
// @param uploadID string 上传ID
// @param totalChunks int 总分片数
// @param chunkSize int64 分片大小
// @return *InitUploadResponse 初始化上传响应对象
func NewInitUploadResponse(uploadID string, totalChunks float64, chunkSize float64) InitUploadResponse {
	return InitUploadResponse{
		UploadID:    uploadID,
		TotalChunks: totalChunks,
		ChunkSize:   chunkSize,
	}
}

// UploadChunkResponse 上传文件分片响应
// @description 上传文件分片后的响应
// @struct
// @property UploadID string 文件ID
// @property ChunkIndex int 分片索引
// @property UploadedChunks int 已上传分片数
// @property TotalChunks int 总分片数
// @property Progress float64 上传进度(0-100)
type UploadChunkResponse struct {
	UploadID       string  `json:"upload_id" yaml:"upload_id"`             // 上传ID
	ChunkIndex     float64 `json:"chunk_index" yaml:"chunk_index"`         // 分片索引
	UploadedChunks float64 `json:"uploaded_chunks" yaml:"uploaded_chunks"` // 已上传分片数
	TotalChunks    float64 `json:"total_chunks" yaml:"total_chunks"`       // 总分片数
	Progress       float64 `json:"progress" yaml:"progress"`               // 上传进度(0-100)
}

// String 将UploadChunkResponse转换为字符串
// @description 将UploadChunkResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp UploadChunkResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewUploadChunkResponse 创建上传分片响应对象
// @description 创建一个新的文件分片上传响应对象
// @param fileID int64 文件ID
// @param chunkIndex int 分片索引
// @param uploadedChunks int 已上传分片数
// @param totalChunks int 总分片数
// @param progress float64 上传进度
// @return *UploadChunkResponse 上传分片响应对象
func NewUploadChunkResponse(uploadId string, chunkIndex, uploadedChunks, totalChunks, progress float64) *UploadChunkResponse {
	return &UploadChunkResponse{
		UploadID:       uploadId,
		ChunkIndex:     chunkIndex,
		UploadedChunks: uploadedChunks,
		TotalChunks:    totalChunks,
		Progress:       progress,
	}
}

// CompleteUploadResponse 完成文件上传响应
// @description 完成文件上传后的响应
// @struct
// @property FileID int64 文件ID
// @property FileName string 文件名
// @property FileSize int64 文件大小
// @property FileHash string 文件哈希值
// @property FilePath string 文件路径
// @property DownloadUrl string 下载URL
// @property UploadTime time.Time 上传完成时间
type CompleteUploadResponse struct {
	UploadID    string    `json:"upload_id" yaml:"upload_id"`       // 上传ID
	FileName    string    `json:"file_name" yaml:"file_name"`       // 文件名
	FileSize    float64   `json:"file_size" yaml:"file_size"`       // 文件大小
	FileHash    string    `json:"file_hash" yaml:"file_hash"`       // 文件哈希值
	FilePath    string    `json:"file_path" yaml:"file_path"`       // 文件路径
	DownloadUrl string    `json:"download_url" yaml:"download_url"` // 下载URL
	UploadTime  time.Time `json:"upload_time" yaml:"upload_time"`   // 上传完成时间
}

// String 将CompleteUploadResponse转换为字符串
// @description 将CompleteUploadResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp CompleteUploadResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewCompleteUploadResponse 创建完成上传响应对象
// @description 创建一个新的完成文件上传响应对象
// @param fileID int64 文件ID
// @param fileName string 文件名
// @param fileSize int64 文件大小
// @param fileHash string 文件哈希值
// @param filePath string 文件路径
// @param downloadUrl string 下载URL
// @param uploadTime time.Time 上传时间
// @return *CompleteUploadResponse 完成上传响应对象
func NewCompleteUploadResponse(uploadId string, fileName string, fileSize float64, fileHash, filePath, downloadUrl string, uploadTime time.Time) *CompleteUploadResponse {
	return &CompleteUploadResponse{
		UploadID:    uploadId,
		FileName:    fileName,
		FileSize:    fileSize,
		FileHash:    fileHash,
		FilePath:    filePath,
		DownloadUrl: downloadUrl,
		UploadTime:  uploadTime,
	}
}

// GetUploadStatusResponse 查询文件上传状态响应
// @description 查询文件上传状态的响应
// @struct
// @property FileID int64 文件ID
// @property FileName string 文件名
// @property FileSize int64 文件大小
// @property ChunkSize int64 分片大小
// @property TotalChunks int 总分片数
// @property UploadedChunks int 已上传分片数
// @property Progress float64 上传进度(0-100)
// @property Status string 上传状态
// @property UploadedChunkIndexes []int 已上传分片索引列表
// @property LastChunkTime time.Time 最后分片上传时间
type GetUploadStatusResponse struct {
	UploadID             string    `json:"upload_id" yaml:"upload_id"`                           // 上传ID
	FileName             string    `json:"file_name" yaml:"file_name"`                           // 文件名
	FileSize             float64   `json:"file_size" yaml:"file_size"`                           // 文件大小
	ChunkSize            float64   `json:"chunk_size" yaml:"chunk_size"`                         // 分片大小
	TotalChunks          float64   `json:"total_chunks" yaml:"total_chunks"`                     // 总分片数
	UploadedChunks       float64   `json:"uploaded_chunks" yaml:"uploaded_chunks"`               // 已上传分片数
	Progress             float64   `json:"progress" yaml:"progress"`                             // 上传进度(0-100)
	Status               string    `json:"status" yaml:"status"`                                 // 上传状态
	UploadedChunkIndexes []int     `json:"uploaded_chunk_indexes" yaml:"uploaded_chunk_indexes"` // 已上传分片索引列表
	LastChunkTime        time.Time `json:"last_chunk_time" yaml:"last_chunk_time"`               // 最后分片上传时间
}

// String 将GetUploadStatusResponse转换为字符串
// @description 将GetUploadStatusResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp GetUploadStatusResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewGetUploadStatusResponse 创建上传状态响应对象
// @description 创建一个新的文件上传状态查询响应对象
// @param fileID int64 文件ID
// @param fileName string 文件名
// @param fileSize int64 文件大小
// @param chunkSize int64 分片大小
// @param totalChunks int 总分片数
// @param uploadedChunks int 已上传分片数
// @param progress float64 上传进度
// @param status string 上传状态
// @param chunkIndexes []int 已上传分片索引
// @param lastChunkTime time.Time 最后分片上传时间
// @return *GetUploadStatusResponse 上传状态响应对象
func NewGetUploadStatusResponse(uploadId, fileName string, fileSize, chunkSize float64, totalChunks float64, uploadedChunks float64, progress float64, status string, chunkIndexes []int, lastChunkTime time.Time) GetUploadStatusResponse {
	return GetUploadStatusResponse{
		UploadID:             uploadId,
		FileName:             fileName,
		FileSize:             fileSize,
		ChunkSize:            chunkSize,
		TotalChunks:          totalChunks,
		UploadedChunks:       uploadedChunks,
		Progress:             progress,
		Status:               status,
		UploadedChunkIndexes: chunkIndexes,
		LastChunkTime:        lastChunkTime,
	}
}

// ListFilesResponse 文件列表响应
// @description 查询文件列表的响应
// @struct
// @property Total int 总记录数
// @property List []FileInfoResponse 文件信息列表
// @property Page int 当前页码
// @property PageSize int 每页大小
type ListFilesResponse struct {
	UploadID       string           `json:"upload_id,omitempty" yaml:"upload_id,omitempty" gorm:"column:upload_id;type:varchar(255);default:'';not null"`         // 上传ID
	FileName       string           `json:"file_name,omitempty" yaml:"file_name,omitempty" gorm:"column:file_name;type:text;default:'';"`                         // 文件名
	FileSize       float64          `json:"file_size,omitempty" yaml:"file_size,omitempty" gorm:"column:file_size;type:float;default:0;"`                         // 文件大小
	FilePath       string           `json:"file_path,omitempty" yaml:"file_path,omitempty" gorm:"column:file_path;type:text;default:'';"`                         // 文件路径
	FileHash       string           `json:"file_hash,omitempty" yaml:"file_hash,omitempty" gorm:"column:file_hash;type:varchar(255);default:'';"`                 // 文件hash
	FileType       string           `json:"file_type,omitempty" yaml:"file_type,omitempty" gorm:"column:file_type;type:varchar(255);default:'';"`                 // 文件类型
	UploaderId     string           `json:"uploader_id,omitempty" yaml:"uploader_id,omitempty" gorm:"column:uploader_id;type:varchar(255);default:'';"`           // 上传者ID
	UploadTime     time.Time        `json:"upload_time,omitempty" yaml:"upload_time,omitempty" gorm:"column:upload_time;type:datetime;default:CURRENT_TIMESTAMP"` // 上传时间
	UpdateTime     time.Time        `json:"update_time,omitempty" yaml:"update_time,omitempty" gorm:"column:update_time;type:datetime;default:CURRENT_TIMESTAMP"` // 更新时间
	Status         model.FileStatus `json:"status,omitempty" yaml:"status,omitempty" gorm:"column:status;type:varchar(255);default:'';"`                          // 使用FileStatus类型
	ChunkSize      float64          `json:"chunk_size,omitempty" yaml:"chunk_size,omitempty" gorm:"column:chunk_size;type:float;default:0;"`                      // 分片大小
	TotalChunks    float64          `json:"total_chunks,omitempty" yaml:"total_chunks,omitempty" gorm:"column:total_chunks;type:float;default:0;"`                // 总分片数
	UploadedChunks float64          `json:"uploaded_chunks,omitempty" yaml:"uploaded_chunks,omitempty" gorm:"column:uploaded_chunks;type:float;default:0;"`       // 已上传分片数
	LastChunkTime  time.Time        `json:"last_chunk_time,omitempty" yaml:"last_chunk_time,omitempty" gorm:"column:last_chunk_time;type:datetime;"`              // 最后上传分片时间
	ExpireTime     time.Time        `json:"expire_time,omitempty" yaml:"expire_time,omitempty" gorm:"column:expire_time;type:datetime;"`                          // 过期时间
	DownloadCount  float64          `json:"download_count,omitempty" yaml:"download_count,omitempty" gorm:"column:download_count;type:float;default:0;"`          // 下载次数
	Description    string           `json:"description,omitempty" yaml:"description,omitempty" gorm:"column:description;type:text;default:'';"`                   // 文件描述
	StorageType    string           `json:"storage_type,omitempty" yaml:"storage_type,omitempty" gorm:"column:storage_type;type:varchar(255);default:'';"`        // 存储类型
	IpfsCid        string           `json:"ipfs_cid,omitempty" yaml:"ipfs_cid,omitempty" gorm:"column:ipfs_cid;type:varchar(255);default:'';"`                    // ipfs cid
	IsDelete       sql.NullBool     `json:"is_delete,omitempty" yaml:"is_delete,omitempty" gorm:"column:is_delete;type:tinyint(1);default:0;"`                    // 是否删除
	IsRemove       sql.NullBool     `json:"is_remove,omitempty" yaml:"is_remove,omitempty" gorm:"column:is_remove;type:tinyint(1);default:0;"`                    // 是否移除
}

func (resp ListFilesResponse) MarshalJSON() ([]byte, error) {
	type Alias ListFilesResponse
	aux := struct {
		*Alias
		Status        string `json:"status" yaml:"status"`                   // 使用FileStatus类型
		IsDelete      bool   `json:"is_delete" yaml:"is_delete"`             // 是否删除
		IsRemove      bool   `json:"is_remove" yaml:"is_remove"`             // 是否移除
		LastChunkTime string `json:"last_chunk_time" yaml:"last_chunk_time"` // 最后上传分片时间
	}{
		Alias:         (*Alias)(&resp),
		Status:        string(resp.Status),
		IsDelete:      resp.IsDelete.Bool,
		IsRemove:      resp.IsRemove.Bool,
		LastChunkTime: humantime.HumanTime(resp.LastChunkTime, "unknown"),
	}
	return json.Marshal(aux)
}

// String 将ListFilesResponse转换为字符串
// @description 将ListFilesResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp ListFilesResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewListFilesResponse 创建文件列表响应对象
// @description 创建一个新的文件列表响应对象
// @param total int 总记录数
// @param list []FileInfoResponse 文件列表
// @param page int 当前页码
// @param pageSize int 每页大小
// @return *ListFilesResponse 文件列表响应对象
func NewListFilesResponse(pageInfo dbpage.Info[model.FileInfo]) (dbpage.Info[ListFilesResponse], error) {
	var resp dbpage.Info[ListFilesResponse]
	var convert []ListFilesResponse
	if err := copier.CopyWithOption(&convert, pageInfo.GetRecords(), copier.Option{
		IgnoreEmpty: true,
		DeepCopy:    true,
	}); err != nil {
		return resp, err
	}

	resp.SetPageNo(pageInfo.PageNo)
	resp.SetPageSize(pageInfo.PageSize)
	resp.SetMaxPage(pageInfo.MaxPage)
	resp.SetCount(pageInfo.Count)
	resp.SetRecords(convert)
	return resp, nil
}

// FileInfoResponse 文件信息响应
// @description 文件详细信息响应
// @struct
// @property ID int64 文件ID
// @property FileName string 文件名
// @property FileSize int64 文件大小
// @property FilePath string 文件路径
// @property FileHash string 文件哈希值
// @property FileType string 文件类型
// @property Uploader string 上传者
// @property UploadTime time.Time 上传时间
// @property Status string 文件状态
// @property DownloadCount int 下载次数
// @property Description string 文件描述
// @property StorageType string 存储类型
// @property ExpireTime time.Time 过期时间
// @property DownloadUrl string 下载URL
// @property Extra map[string]interface{} 额外信息
type FileInfoResponse struct {
	ID            int64                  `json:"id" yaml:"id"`                         // 文件ID
	FileName      string                 `json:"file_name" yaml:"file_name"`           // 文件名
	FileSize      int64                  `json:"file_size" yaml:"file_size"`           // 文件大小
	FilePath      string                 `json:"file_path" yaml:"file_path"`           // 文件路径
	FileHash      string                 `json:"file_hash" yaml:"file_hash"`           // 文件哈希值
	FileType      string                 `json:"file_type" yaml:"file_type"`           // 文件类型
	Uploader      string                 `json:"uploader" yaml:"uploader"`             // 上传者
	UploadTime    time.Time              `json:"upload_time" yaml:"upload_time"`       // 上传时间
	Status        string                 `json:"status" yaml:"status"`                 // 文件状态
	DownloadCount int                    `json:"download_count" yaml:"download_count"` // 下载次数
	Description   string                 `json:"description" yaml:"description"`       // 文件描述
	StorageType   string                 `json:"storage_type" yaml:"storage_type"`     // 存储类型
	ExpireTime    time.Time              `json:"expire_time" yaml:"expire_time"`       // 过期时间
	DownloadUrl   string                 `json:"download_url" yaml:"download_url"`     // 下载URL
	Extra         map[string]interface{} `json:"extra" yaml:"extra"`                   // 额外信息
}

// String 将FileInfoResponse转换为字符串
// @description 将FileInfoResponse结构体转换为JSON格式的字符串
// @return string 格式化的JSON字符串
func (resp FileInfoResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

// NewFileInfoResponse 创建文件信息响应对象
// @description 创建一个新的文件详细信息响应对象
// @param id int64 文件ID
// @param fileName string 文件名
// @param fileSize int64 文件大小
// @param filePath string 文件路径
// @param fileHash string 文件哈希值
// @param fileType string 文件类型
// @param uploader string 上传者
// @param uploadTime time.Time 上传时间
// @param status string 文件状态
// @param downloadCount int 下载次数
// @param description string 文件描述
// @param storageType string 存储类型
// @param expireTime time.Time 过期时间
// @param downloadUrl string 下载URL
// @param extra map[string]interface{} 额外信息
// @return *FileInfoResponse 文件信息响应对象
func NewFileInfoResponse(id int64, fileName string, fileSize int64, filePath, fileHash, fileType, uploader string, uploadTime time.Time, status string, downloadCount int, description, storageType string, expireTime time.Time, downloadUrl string, extra map[string]interface{}) *FileInfoResponse {
	return &FileInfoResponse{
		ID:            id,
		FileName:      fileName,
		FileSize:      fileSize,
		FilePath:      filePath,
		FileHash:      fileHash,
		FileType:      fileType,
		Uploader:      uploader,
		UploadTime:    uploadTime,
		Status:        status,
		DownloadCount: downloadCount,
		Description:   description,
		StorageType:   storageType,
		ExpireTime:    expireTime,
		DownloadUrl:   downloadUrl,
		Extra:         extra,
	}
}

type ResumeUploadResponse struct {
	FileName       string    `json:"file_name" yaml:"file_name"`
	FileSize       float64   `json:"file_size" yaml:"file_size"`
	TotalChunks    float64   `json:"total_chunks" yaml:"total_chunks"`
	ChunkSize      float64   `json:"chunk_size" yaml:"chunk_size"`
	UploadedChunks int       `json:"uploaded_chunks" yaml:"uploaded_chunks"`
	ChunkIndices   []int     `json:"chunk_indices" yaml:"chunk_indices"`
	Progress       float64   `json:"progress" yaml:"progress"`
	Status         string    `json:"status" yaml:"status"`
	UploadId       string    `json:"upload_id" yaml:"upload_id"`
	CreateTime     time.Time `json:"create_time" yaml:"create_time"`
}

func (resp ResumeUploadResponse) String() string {
	str, _ := json.MarshalString(resp)
	return str
}

func NewResumeUploadResponse(fileName string, fileSize, totalChunks, chunkSize float64, uploadedChunks int, chunkIndices []int, progress float64, status string, uploadId string, createTime time.Time) ResumeUploadResponse {
	return ResumeUploadResponse{
		FileName:       fileName,
		FileSize:       fileSize,
		TotalChunks:    totalChunks,
		ChunkSize:      chunkSize,
		UploadedChunks: uploadedChunks,
		ChunkIndices:   chunkIndices,
		Progress:       progress,
		Status:         status,
		UploadId:       uploadId,
		CreateTime:     createTime,
	}
}
