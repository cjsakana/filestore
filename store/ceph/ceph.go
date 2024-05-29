package ceph

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"os"
)
import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	endpoint = "http://192.168.0.101:7480/" //endpoint设置，不要动
	client   *s3.Client
)

// GetCephConn 获取连接
func GetCephConn() *s3.Client {
	//Endpoint的配置需要单独写一个Resolver
	customResolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: endpoint,
			}, nil
		})

	//然后使用这个Resolver创建一个S3客户端
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(customResolver),
		// 一定要配置地区.....
		config.WithRegion("us-east-1"))
	if err != nil {
		fmt.Println("ceph connect err:", err)
		return nil
	}

	client = s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.UsePathStyle = true
	})
	return client
}

// CreateBucket 创建桶
func CreateBucket(client *s3.Client, bktName string) {
	bktInput := &s3.CreateBucketInput{
		Bucket: &bktName,
	}

	_, err := client.CreateBucket(context.TODO(), bktInput)
	if err != nil {
		panic(err)
	}
}

// ListBuckets 遍历桶
func ListBuckets(client *s3.Client) {
	out, err := client.ListBuckets(context.TODO(), nil)
	if err != nil {
		panic(err)
	}

	fmt.Print("[")
	for i := range out.Buckets {
		if i > 0 {
			fmt.Print(", ")
		}
		bkt := out.Buckets[i]
		fmt.Print(*bkt.Name)
	}
	fmt.Print("]\n")
}

// UploadFile 上传文件到ceph
func UploadFile(client *s3.Client, bucket, key, filePath string) {
	// 打开本地文件
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("upload file err=", err)
	}
	defer file.Close()
	// 上传文件到S3 bucket
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      &bucket,
		Key:         &key,
		Body:        file,
		ACL:         types.ObjectCannedACLPublicRead, // 设置为只读权限
		ContentType: aws.String("octet-stream"),
	})
	if err != nil {
		fmt.Println("upload file err=", err)
		return
	}
	fmt.Printf("upload file success, bucket=%v, name=%v\n", bucket, key)
}

// ListFile 遍历桶的文件
func ListFile(client *s3.Client, bucket string) {
	bktInput := s3.ListObjectsV2Input{
		Bucket: &bucket,
	}
	out, err := client.ListObjectsV2(context.TODO(), &bktInput)
	if err != nil {
		panic(err)
	}
	for i := range out.Contents {
		obj := out.Contents[i]
		fmt.Printf("out: %v\n", *obj.Key)
	}
}
