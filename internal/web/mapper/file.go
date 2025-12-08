package mapper

import (
	"context"
	"errors"
	"math"
	"time"
	
	"github.com/jianlu8023/go-tools/v2/pkg/sqlnull"
	"github.com/jianlu8023/go-tools/v2/pkg/stringer"
	"github.com/jianlu8023/golang-example/internal/web/model"
	"github.com/jianlu8023/golang-example/pkg/control/datasource"
	"github.com/jianlu8023/golang-example/pkg/control/tracer"
	"github.com/jianlu8023/golang-example/pkg/dbpage"
	"go.opentelemetry.io/otel/codes"
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

// QueryFileInfoExist 查询文件是否存在
// @description 根据文件查询条件查询文件是否存在
// @param query model.FileInfo 文件查询条件，非空字段将作为查询条件
// @return bool 文件是否存在
// @return error 错误信息
func (m *FileMapper) QueryFileInfoExist(ctx context.Context, query model.FileInfo) (bool, error) {
	_, span := tracer.StartSpan(ctx, "fileMapper", "queryFileExist")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return false, datasource.ErrNoDataSourceConn
	}

	var count int64
	if err := m.db.WithContext(ctx).Model(&model.FileInfo{}).Where(&query).Count(&count).Error; err != nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return false, err
	}
	span.SetStatus(codes.Ok, "success")
	return count > 0, nil
}

// CreateFileInfo 创建文件信息记录
// @description 直接插入文件信息，不包含存在性检查，应由service层调用QueryFileExist进行检查
func (m *FileMapper) CreateFileInfo(ctx context.Context, fileInfo *model.FileInfo) error {
	_, span := tracer.StartSpan(ctx, "fileMapper", "createFileInfo")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return datasource.ErrNoDataSourceConn
	}
	// 直接插入文件信息
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.FileInfo{}).Create(fileInfo).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		span.SetStatus(codes.Ok, "success")
		return nil
	})
}

func (m *FileMapper) GetFileInfoOneByQuery(ctx context.Context, query model.FileInfo) (*model.FileInfo, error) {
	_, span := tracer.StartSpan(ctx, "fileMapper", "getOneByQuery")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return nil, datasource.ErrNoDataSourceConn
	}

	var record model.FileInfo
	if err := m.db.WithContext(ctx).Model(&model.FileInfo{}).Where(&query).First(&record).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	span.SetStatus(codes.Ok, "success")
	return &record, nil
}

// GetFileInfoByID 根据ID获取文件信息
func (m *FileMapper) GetFileInfoByID(autoUid int) (*model.FileInfo, error) {
	if m.db == nil {
		return nil, datasource.ErrNoDataSourceConn
	}

	fileInfo := &model.FileInfo{}
	if err := m.db.Where(&model.FileInfo{
		AutoUid: autoUid,
	}).First(fileInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	if err := m.db.Where(&model.FileInfo{
		UploadID: uploadID,
	}).First(fileInfo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, datasource.ErrNotExists
		}
		return nil, err
	}

	return fileInfo, nil
}

// UpdateFileInfo 更新文件信息
// @description 直接更新文件信息，不包含存在性检查，应由service层调用QueryFileExist进行检查
func (m *FileMapper) UpdateFileInfo(ctx context.Context, fileInfo *model.FileInfo) error {
	_, span := tracer.StartSpan(ctx, "fileMapper", "updateFileInfo")
	defer span.End()
	if m.db == nil {
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		span.RecordError(datasource.ErrNoDataSourceConn)
		return datasource.ErrNoDataSourceConn
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.FileInfo
		if err := tx.WithContext(ctx).Model(&model.FileInfo{}).Where(&model.FileInfo{
			UploadID: fileInfo.UploadID,
		}).First(&exist).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		if stringer.IsBlank(exist.UploadID) {
			span.SetStatus(codes.Error, datasource.ErrNotExists.Error())
			span.RecordError(datasource.ErrNotExists)
			return datasource.ErrNotExists
		}

		fileInfo.UpdateTime = time.Now()
		fileInfo.AutoUid = exist.AutoUid
		if err := tx.WithContext(ctx).Model(&model.FileInfo{}).Where(&model.FileInfo{
			AutoUid: fileInfo.AutoUid,
		}).Updates(fileInfo).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		span.SetStatus(codes.Ok, "success")
		return nil
	})
}

// CreateFileChunk 创建文件分片记录
// @description 直接插入分片信息，不包含存在性检查，应由service层调用CheckChunkExists进行检查
func (m *FileMapper) CreateFileChunk(ctx context.Context, fileChunk *model.FileChunk) error {
	_, span := tracer.StartSpan(ctx, "fileMapper", "createFileChunk")
	defer span.End()

	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return datasource.ErrNoDataSourceConn
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Model(&model.FileChunk{}).Create(fileChunk).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		span.SetStatus(codes.Ok, "success")
		return nil
	})
}

// CheckFileChunkExists 检查分片是否已存在
func (m *FileMapper) CheckFileChunkExists(ctx context.Context, query model.FileChunk) (bool, error) {
	_, span := tracer.StartSpan(ctx, "fileMapper", "checkChunkExists")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return false, datasource.ErrNoDataSourceConn
	}

	var count int64
	if err := m.db.WithContext(ctx).Model(&model.FileChunk{}).Where(&query).Count(&count).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return false, err
	}
	span.SetStatus(codes.Ok, "success")
	return count > 0, nil
}

// UpdateFileChunk 更新分片信息
// @description 直接更新分片信息，不包含存在性检查，应由service层调用CheckChunkExists进行检查
func (m *FileMapper) UpdateFileChunk(ctx context.Context, fileChunk *model.FileChunk) error {
	_, span := tracer.StartSpan(ctx, "fileMapper", "updateFileChunk")
	defer span.End()
	if m.db == nil {
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		span.RecordError(datasource.ErrNoDataSourceConn)
		return datasource.ErrNoDataSourceConn
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exist model.FileChunk
		if err := tx.WithContext(ctx).Model(&model.FileChunk{}).Where(&model.FileChunk{
			UploadID:   fileChunk.UploadID,
			ChunkIndex: fileChunk.ChunkIndex,
		}).First(&exist).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		if stringer.IsBlank(exist.UploadID) {
			span.SetStatus(codes.Error, datasource.ErrNotExists.Error())
			span.RecordError(datasource.ErrNotExists)
			return datasource.ErrNotExists
		}
		fileChunk.AutoUid = exist.AutoUid
		fileChunk.UploadTime = time.Now()
		if err := tx.WithContext(ctx).Model(&model.FileChunk{}).Where(&model.FileChunk{
			AutoUid: fileChunk.AutoUid,
		}).Updates(fileChunk).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		span.SetStatus(codes.Ok, "success")
		return nil
	})
}

// GetUploadedChunkCount 获取已上传分片数
func (m *FileMapper) GetUploadedChunkCount(ctx context.Context, uploadId string) (float64, error) {
	_, span := tracer.StartSpan(ctx, "fileMapper", "getUploadedChunkCount")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return 0, datasource.ErrNoDataSourceConn
	}

	var count int64
	if err := m.db.WithContext(ctx).Model(&model.FileChunk{}).Where(&model.FileChunk{
		UploadID: uploadId,
		IsDelete: sqlnull.FalseToNull(), // 只查询未删除的记录
	}).Count(&count).Error; err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return -1, err
	}
	span.SetStatus(codes.Ok, "success")
	return float64(count), nil
}

// GetFileChunks 获取所有分片信息
func (m *FileMapper) GetFileChunks(ctx context.Context, query model.FileChunk) ([]model.FileChunk, error) {
	_, span := tracer.StartSpan(ctx, "fileMapper", "getFileChunks")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return nil, datasource.ErrNoDataSourceConn
	}

	// 添加IsDelete条件，确保只查询未删除的记录
	query.IsDelete = sqlnull.FalseToNull()
	var chunks []model.FileChunk
	if err := m.db.WithContext(ctx).
		Model(&model.FileChunk{}).
		Where(&query).
		Order("chunk_index ASC"). // 按索引排序
		Find(&chunks).
		Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	span.SetStatus(codes.Ok, "success")
	return chunks, nil
}

// GetUploadedFileChunkIndexes 获取已上传分片索引
func (m *FileMapper) GetUploadedFileChunkIndexes(ctx context.Context, query model.FileChunk) ([]int, error) {
	_, span := tracer.StartSpan(ctx, "fileMapper", "getUploadedFileChunkIndexes")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return nil, datasource.ErrNoDataSourceConn
	}

	// 修复：使用float64数组接收数据，然后转换为int
	var chunkIndexes []float64
	// 添加IsDelete条件，确保只查询未删除的记录
	query.IsDelete = sqlnull.FalseToNull()
	if err := m.db.Model(model.NewFileChunk()).
		Where(&query).
		Order("chunk_index ASC").
		Pluck("chunk_index", &chunkIndexes).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// 转换float64为int，并验证索引的合法性
	indexes := make([]int, 0, len(chunkIndexes))
	for _, idx := range chunkIndexes {
		intIdx := int(idx)
		// 验证索引的合法性（非负数且为整数）
		if intIdx >= 0 && float64(intIdx) == idx {
			indexes = append(indexes, intIdx)
		} else {
			// 记录不合法的索引但不中断查询
			m.logger.Warnf("invalid chunk index found: %.2f, uploadId: %s", idx, query.UploadID)
		}
	}

	span.SetStatus(codes.Ok, "success")
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
func (m *FileMapper) ListFiles(ctx context.Context, query model.FileInfo, pageInfo dbpage.Info[model.FileInfo]) (dbpage.Info[model.FileInfo], error) {
	_, span := tracer.StartSpan(ctx, "fileMapper", "listFiles")
	defer span.End()
	if m.db == nil {
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		span.RecordError(datasource.ErrNoDataSourceConn)
		return pageInfo, datasource.ErrNoDataSourceConn
	}

	var count int64
	if err := m.db.WithContext(ctx).Model(model.NewFileInfo()).Where(&query).Count(&count).Error; err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return pageInfo, err
	}
	pageInfo.SetCount(count)
	maxPage := math.Ceil(float64(count) / float64(pageInfo.PageSize))
	pageInfo.SetMaxPage(int64(maxPage))

	var files []model.FileInfo
	if err := m.db.WithContext(ctx).Model(model.NewFileInfo()).Where(&query).
		Offset(int((pageInfo.PageNo - 1) * pageInfo.PageSize)).
		Limit(int(pageInfo.PageSize)).Find(&files).Error; err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
		return pageInfo, err
	}
	pageInfo.SetRecords(files)
	span.SetStatus(codes.Ok, "success")
	return pageInfo, nil
}

// DeleteFileInfo 删除文件信息
// @description 直接删除文件信息，不包含存在性检查，应由service层调用QueryFileExist进行检查
func (m *FileMapper) DeleteFileInfo(ctx context.Context, query model.FileInfo) error {
	_, span := tracer.StartSpan(ctx, "fileMapper", "deleteFileInfo")
	defer span.End()
	if m.db == nil {
		span.RecordError(datasource.ErrNoDataSourceConn)
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		return datasource.ErrNoDataSourceConn
	}
	return m.db.WithContext(ctx).Model(model.NewFileInfo()).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.WithContext(ctx).Model(model.NewFileInfo()).Where(&query).Count(&count).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				span.SetStatus(codes.Ok, "success")
				return nil
			}
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		if count <= 0 {
			span.SetStatus(codes.Ok, "success")
			return nil
		}
		if err := tx.WithContext(ctx).Model(model.NewFileInfo()).Where(&query).Where(
			&model.FileInfo{
				IsDelete: sqlnull.FalseToNull(),
			}).Updates(&model.FileChunk{
			IsDelete: sqlnull.TrueToNull(),
		}).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		span.SetStatus(codes.Ok, "success")
		return nil
	})
}

// DeleteFileChunks 删除分片信息
func (m *FileMapper) DeleteFileChunks(ctx context.Context, query model.FileChunk) error {
	_, span := tracer.StartSpan(ctx, "fileMapper", "deleteFileChunks")
	defer span.End()
	if m.db == nil {
		span.SetStatus(codes.Error, datasource.ErrNoDataSourceConn.Error())
		span.RecordError(datasource.ErrNoDataSourceConn)
		return datasource.ErrNoDataSourceConn
	}

	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.WithContext(ctx).Model(model.NewFileChunk()).Where(&query).
			Where(&model.FileChunk{
				IsDelete: sqlnull.FalseToNull(),
			}).
			Count(&count).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				span.SetStatus(codes.Ok, "success")
				return nil
			}
			span.SetStatus(codes.Error, err.Error())
			span.RecordError(err)
			return err
		}

		if count <= 0 {
			span.SetStatus(codes.Ok, "success")
			return nil
		}

		if err := tx.WithContext(ctx).Model(model.NewFileChunk()).Where(&model.FileChunk{
			IsDelete: sqlnull.FalseToNull(),
		}).Where(&query).Updates(&model.FileChunk{
			IsDelete: sqlnull.TrueToNull(),
		}).Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
		span.SetStatus(codes.Ok, "success")
		return nil
	})
}
