package mapper

import (
	"time"

	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"gorm.io/gorm"
)

type FileMapper struct {
	*Mapper
}

func NewFileMapper(baseMapper *Mapper) *FileMapper {
	return &FileMapper{
		Mapper: baseMapper,
	}
}

// QueryFileExist 查询文件是否存在
// @description 根据文件查询条件查询文件是否存在
// @param query *model.FileInfo 文件查询条件，非空字段将作为查询条件
// @return bool 文件是否存在
// @return error 错误信息
func (m *FileMapper) QueryFileExist(query *model.FileInfo) (bool, error) {
	if m.db == nil {
		return false, datasource.ErrNoDataSourceConn
	}

	var count int64
	if err := m.db.Model(&model.FileInfo{}).Where(query).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateFileInfo 创建文件信息记录
// @description 直接插入文件信息，不包含存在性检查，应由service层调用QueryFileExist进行检查
func (m *FileMapper) CreateFileInfo(fileInfo *model.FileInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}

	// 直接插入文件信息
	return m.db.Create(fileInfo).Error
}

// GetFileInfoByID 根据ID获取文件信息
func (m *FileMapper) GetFileInfoByID(autoUid int) (*model.FileInfo, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}

	fileInfo := &model.FileInfo{}
	if err := m.db.Where("auto_uid = ?", autoUid).First(fileInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, datasource.ErrNotExists
		}
		return nil, err
	}

	return fileInfo, nil
}

// GetFileInfoByUploadID 根据上传ID获取文件信息
func (m *FileMapper) GetFileInfoByUploadID(uploadID string) (*model.FileInfo, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}

	fileInfo := &model.FileInfo{}
	if err := m.db.Where("upload_id = ?", uploadID).First(fileInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, datasource.ErrNotExists
		}
		return nil, err
	}

	return fileInfo, nil
}

// UpdateFileInfo 更新文件信息
// @description 直接更新文件信息，不包含存在性检查，应由service层调用QueryFileExist进行检查
func (m *FileMapper) UpdateFileInfo(fileInfo *model.FileInfo) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}

	fileInfo.UpdateTime = time.Now()
	return m.db.Model(fileInfo).Updates(fileInfo).Error

}

// CreateFileChunk 创建文件分片记录
// @description 直接插入分片信息，不包含存在性检查，应由service层调用CheckChunkExists进行检查
func (m *FileMapper) CreateFileChunk(fileChunk *model.FileChunk) (int64, error) {
	if m.db == nil {
		return 0, datasource.ErrNoDataSourceConn
	}

	if err := m.db.Create(fileChunk).Error; err != nil {
		return 0, err
	}

	return 1, nil
}

// CheckChunkExists 检查分片是否已存在
func (m *FileMapper) CheckChunkExists(fileID int64, chunkIndex int) (bool, error) {
	if m.db == nil {
		return false, datasource.ErrNoDataSourceConn
	}

	var count int64
	if err := m.db.Model(&model.FileChunk{}).
		Where("file_id = ? AND chunk_index = ?", fileID, chunkIndex).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

// UpdateFileChunk 更新分片信息
// @description 直接更新分片信息，不包含存在性检查，应由service层调用CheckChunkExists进行检查
func (m *FileMapper) UpdateFileChunk(fileChunk *model.FileChunk) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}

	fileChunk.UploadTime = time.Now()
	return m.db.Model(fileChunk).
		Where("file_id = ? AND chunk_index = ?", fileChunk.UploadID, fileChunk.ChunkIndex).
		Updates(fileChunk).Error
}

// GetUploadedChunkCount 获取已上传分片数
func (m *FileMapper) GetUploadedChunkCount(fileID int64) (int, error) {
	if m.db == nil {
		return 0, datasource.ErrNoDataSourceConn
	}

	var count int64
	if err := m.db.Model(&model.FileChunk{}).
		Where("file_id = ?", fileID).
		Count(&count).Error; err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetFileChunks 获取所有分片信息
func (m *FileMapper) GetFileChunks(fileID int) ([]*model.FileChunk, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}

	var chunks []*model.FileChunk
	if err := m.db.Where("file_id = ?", fileID).
		Order("chunk_index ASC").
		Find(&chunks).Error; err != nil {
		return nil, err
	}

	return chunks, nil
}

// GetUploadedChunkIndexes 获取已上传分片索引
func (m *FileMapper) GetUploadedChunkIndexes(fileID int) ([]int, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}

	var indexes []int
	if err := m.db.Model(&model.FileChunk{}).
		Where("file_id = ?", fileID).
		Order("chunk_index ASC").
		Pluck("chunk_index", &indexes).Error; err != nil {
		return nil, err
	}

	return indexes, nil
}

// GetLastChunkUploadTime 获取最后分片上传时间
func (m *FileMapper) GetLastChunkUploadTime(fileID int) (time.Time, error) {
	if m.db == nil {
		return time.Time{}, datasource.ErrNoDataSourceConn
	}

	var maxUploadTime time.Time
	if err := m.db.Model(&model.FileChunk{}).
		Where("file_id = ?", fileID).
		Select("MAX(upload_time)").
		Scan(&maxUploadTime).Error; err != nil {
		return time.Time{}, err
	}

	return maxUploadTime, nil
}

// ListFiles 列出文件
func (m *FileMapper) ListFiles(fileName, status, uploaderId string, offset, limit int, orderBy, orderType string) ([]*model.FileInfo, int, error) {
	if m.db == nil {
		return nil, 0, datasource.ErrNoDataSourceConn
	}

	var files []*model.FileInfo
	query := m.db.Model(&model.FileInfo{})

	// 构建查询条件
	if fileName != "" {
		query = query.Where("file_name LIKE ?", "%"+fileName+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if uploaderId != "" {
		query = query.Where("uploader_id = ?", uploaderId)
	}

	// 获取总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 构建排序
	if orderBy != "" {
		orderDirection := "ASC"
		if orderType == "desc" {
			orderDirection = "DESC"
		}
		query = query.Order(orderBy + " " + orderDirection)
	} else {
		// 默认按更新时间倒序
		query = query.Order("update_time DESC")
	}

	// 分页查询
	if err := query.Offset(offset).Limit(limit).Find(&files).Error; err != nil {
		return nil, 0, err
	}

	return files, int(total), nil
}

// DeleteFileInfo 删除文件信息
// @description 直接删除文件信息，不包含存在性检查，应由service层调用QueryFileExist进行检查
func (m *FileMapper) DeleteFileInfo(fileID int) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}

	return m.db.Where("auto_uid = ?", fileID).Delete(&model.FileInfo{}).Error

}

// DeleteFileChunks 删除分片信息
func (m *FileMapper) DeleteFileChunks(fileID int) error {
	if m.db == nil {
		return datasource.ErrNoDataSourceConn
	}

	return m.db.Where("file_id = ?", fileID).Delete(&model.FileChunk{}).Error
}
