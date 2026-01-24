package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	appconfig "github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

// CloudStorage 云端存储
type CloudStorage struct {
	cfg    *appconfig.Config
	log    *logger.Logger
	bucket string
	region string
	client *s3.Client
}

// NewCloudStorage 创建云端存储实例
func NewCloudStorage(cfg *appconfig.Config, log *logger.Logger) (*CloudStorage, error) {
	bucket := cfg.CloudBucket
	region := cfg.CloudRegion
	endpoint := cfg.CloudEndpoint
	accessKey := cfg.CloudAccessKey
	secretKey := cfg.CloudSecretKey
	useSSL := cfg.CloudUseSSL
	if bucket == "" {
		return nil, fmt.Errorf("cloud bucket configuration is required")
	}
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("cloud access key and secret key are required")
	}

	if region == "" {
		region = "auto"
	}
	if endpoint == "" {
		endpoint = "https://r2.cloudflarestorage.com"
	}

	log.Infof("Initializing cloud storage: bucket=%s, region=%s, endpoint=%s",
		bucket, region, endpoint)

	// 创建AWS配置
	cfgOpts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			accessKey,
			secretKey,
			"", // session token
		)),
	}

	// 如果是CloudFlare R1/R2
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			scheme := "https"
			if !useSSL {
				scheme = "http"
			}
			return aws.Endpoint{
				URL:               fmt.Sprintf("%s://%s", scheme, endpoint),
				HostnameImmutable: true,
				SigningRegion:     region,
			}, nil
		}
		// 回退
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	cfgOpts = append(cfgOpts, awsconfig.WithEndpointResolverWithOptions(resolver))
	cfgOpts = append(cfgOpts, awsconfig.WithRegion(region))

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), cfgOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// 创建S3客户端
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Warnf("Failed to connect to bucket (might not exist): %v", err)
	}

	return &CloudStorage{
		cfg:    cfg,
		log:    log,
		bucket: bucket,
		region: region,
		client: client,
	}, nil
}

// Save 保存图片到云端存储
func (s *CloudStorage) Save(ctx context.Context, hash string, data []byte, mimeType string) (string, int64, error) {
	// 生成云端存储路径（按hash前两位分目录）
	key := fmt.Sprintf("%s/%s", hash[:2], hash)

	s.log.Infof("Uploading to cloud storage: bucket=%s, key=%s, size=%d",
		s.bucket, key, len(data))

	// 上传到S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(mimeType),
		ContentLength: aws.Int64(int64(len(data))),
		// 设置缓存控制（1年）
		CacheControl: aws.String("public, max-age=31536000"),
	})

	if err != nil {
		return "", 0, fmt.Errorf("failed to upload to cloud storage: %w", err)
	}

	s.log.Infof("Successfully uploaded to cloud storage: bucket=%s, key=%s", s.bucket, key)
	return key, int64(len(data)), nil
}

// GetAbsPath 根据相对路径获取访问URL
func (s *CloudStorage) GetAbsPath(ctx context.Context, relPath string) (string, error) {
	if s.cfg.CloudEndpoint != "" && strings.Contains(s.cfg.CloudEndpoint, ".") {
		scheme := "https"
		if !s.cfg.CloudUseSSL {
			scheme = "http"
		}
		return fmt.Sprintf("%s://%s/%s", scheme, s.cfg.CloudEndpoint, relPath), nil
	}

	return fmt.Sprintf("https://%s.r2.dev/%s", s.bucket, relPath), nil
}

// Delete 删除云端文件
func (s *CloudStorage) Delete(ctx context.Context, relPath string) error {
	s.log.Infof("Deleting from cloud storage: bucket=%s, key=%s",
		s.bucket, relPath)

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(relPath),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from cloud storage: %w", err)
	}

	return nil
}

// Exists 检查文件是否存在于云端
func (s *CloudStorage) Exists(ctx context.Context, relPath string) (bool, error) {
	s.log.Infof("Checking existence in cloud storage: bucket=%s, key=%s",
		s.bucket, relPath)

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(relPath),
	})

	if err != nil {
		// 检查是否是"NotFound"错误
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "NoSuchKey") {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// Type 返回存储类型
func (s *CloudStorage) Type() string {
	return "cloud"
}

// CloudStorageConfig 云端存储配置
type CloudStorageConfig struct {
	Endpoint  string
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

// parseCloudStorageConfig 从环境变量解析云端存储配置
func parseCloudStorageConfig(cfg *appconfig.Config) *CloudStorageConfig {
	return &CloudStorageConfig{
		Endpoint:  cfg.CloudEndpoint,
		Bucket:    cfg.CloudBucket,
		Region:    cfg.CloudRegion,
		AccessKey: cfg.CloudAccessKey,
		SecretKey: cfg.CloudSecretKey,
		UseSSL:    cfg.CloudUseSSL,
	}
}

// getCloudStorageKey 生成云端存储的key
func getCloudStorageKey(hash string) string {
	key := strings.TrimPrefix(hash[:2], "/")
	key = strings.TrimSuffix(key, "/")
	return fmt.Sprintf("%s/%s", key, hash)
}

// EnsureBucketExists 确保存储桶存在
func (s *CloudStorage) EnsureBucketExists(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})

	if err != nil {
		s.log.Infof("Bucket %s does not exist, creating...", s.bucket)

		// 创建桶
		_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(s.bucket),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(s.region),
			},
		})

		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}

		s.log.Infof("Bucket %s created successfully", s.bucket)

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": "s3:GetObject",
					"Resource": "arn:aws:s3:::%s/*"
				}
			]
		}`, s.bucket)

		_, err = s.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
			Bucket: aws.String(s.bucket),
			Policy: aws.String(policy),
		})

		if err != nil {
			s.log.Warnf("Failed to set bucket policy: %v", err)
		}
	}

	return nil
}
