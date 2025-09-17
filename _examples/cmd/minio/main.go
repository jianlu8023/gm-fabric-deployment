package main

import (
	"context"

	"github.com/jianlu8023/go-tools/pkg/format/json"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"_examples/internal/logger"
)

var appLogger = logger.GetAppLogger()

func main() {
	appLogger.Infof("starting test minio ...")
	ctx := context.Background()
	endpoint := "192.168.209.128:9000"
	accessKeyID := "5sqWsok3bJMlMQTbTfjm"
	secretAccessKey := "XDqQKH4GroEg55V9oASOI3kR6j5wgaeOuvhxxsSb"
	useSSL := false

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		appLogger.Errorf("failed to create minio client: %v", err)
		return
	}

	// Upload the test file
	// Change the value of filePath if the file is in another location
	// Upload the test file with FPutObject
	// info, err := minioClient.FPutObject(
	//	ctx,
	//	"go-demo",
	//	"demo.txt",
	//	"./testdata/demo.txt",
	//	minio.PutObjectOptions{ContentType: "application/octet-stream"})
	// if err != nil {
	//	appLogger.Errorf("failed to upload file: %v", err)
	//	return
	// }
	//
	// appLogger.Infof("Successfully uploaded %s of size %d", objectName, info.Size)

	buckets, err := minioClient.ListBuckets(ctx)
	if err != nil {

		appLogger.Errorf("failed to list buckets: %v", err)
		return
	}
	for _, bucket := range buckets {
		toJSON, _ := json.ToJSON(bucket)
		appLogger.Infof("bucket name %v", toJSON)
		objects := minioClient.ListObjects(ctx, bucket.Name, minio.ListObjectsOptions{})
		for obj := range objects {
			toJSON, _ := json.PrettyJSON(obj)
			appLogger.Infof("object name\n %v", toJSON)
		}
	}
}
