package minio

import (
	"context"
	"github.com/minio/minio-go/v7"

	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ctx = context.Background()

type MinioClient struct {
	client *minio.Client
}

func NewMinIOClient(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*MinioClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{client: minioClient}, nil
}

func (m *MinioClient) UploadFile(bucketName,
	objectName,
	filePath string) (*minio.UploadInfo, error) {
	info, err := m.client.FPutObject(
		ctx,
		bucketName,
		objectName,
		filePath,
		minio.PutObjectOptions{ContentType: "application/octet-stream"})
	if err != nil {
		return nil, err
	}
	return &info, nil
}
