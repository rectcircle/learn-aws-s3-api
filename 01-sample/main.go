package main

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	// 配置参数
	ctx := context.Background()
	endpoint := "127.0.0.1:9000"
	accessKeyID := "1qJ4sGlF6HzTWIHsakYK"
	secretAccessKey := "UwkpCLEMX2ODx5Cg9FfsxGGokIWXRofFwO8Chiq0"
	useSSL := false

	// 初始化 minio 客户端
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		panic(err)
	}

	// bucket 配置
	bucketName := "bucket-test"
	// location := "us-east-1"  // minio 部署的时候可以配置，默认测试的服务的默认值是 "us-east-1"

	// 测试数据
	// "dir1/file1" 为 object name，格式一般写法为类似于文件系统的路径，
	//              以 / 分割，但最前面不需要有 /，即使有也会被忽略 TrimPrefix 掉。
	testData := [][2]string{
		{"dir1/file1", "abcdef"},
		{"dir1/file2", "abcdef"},
		{"dir1/dir2/file3", "abcdef"},
		{"file2", "abcdef"},
	}

	// 上传文件
	for _, item := range testData {
		filePath := item[0]
		fileContent := item[1]

		info, err := minioClient.PutObject(ctx, bucketName, filePath, bytes.NewBufferString(fileContent), int64(len(fileContent)),
			minio.PutObjectOptions{
				ContentType: "text/plain", // 给对象设置 MIME 媒体类型，会影响匿名开发下载链接返回的媒体类型。
			})
		if err != nil {
			panic(err)
		}
		fmt.Printf("upload %s success\n", info.Key)
	}

	// 读取某个文件
	obj, err := minioClient.GetObject(ctx, bucketName, "file2", minio.GetObjectOptions{})
	if err != nil {
		panic(err)
	}
	content, err := io.ReadAll(obj)
	if err != nil {
		panic(err)
	}
	fmt.Printf("GetObject file2, content is %s\n", string(content))

	// 遍历读取文件
	// 特别说明的是，object name 仅仅是个字符串，没有目录的那种层级关系。因此：
	// 1. 要想类似于文件系统目录方式检索对象，其实现是基于字符串前缀的方式。
	// 2. 没有空目录的概念，如果想实现需要通过将目录元信息编码为一个对象。
	dir1RecursiveItemsChan := minioClient.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    "dir1",
		Recursive: true,
	})
	for item := range dir1RecursiveItemsChan {
		obj, err := minioClient.GetObject(ctx, bucketName, item.Key, minio.GetObjectOptions{})
		if err != nil {
			panic(err)
		}
		content, err := io.ReadAll(obj)
		if err != nil {
			panic(err)
		}
		fmt.Printf("ListObjects dir1, objectName is %s, content is %s\n", item.Key, string(content))
	}

	// 由于此 bucket 这设置了匿名访问，所以可以通过如下链接直接下载内容
	// curl http://127.0.0.1:9000/bucket-test/dir1/dir2/file3
}
