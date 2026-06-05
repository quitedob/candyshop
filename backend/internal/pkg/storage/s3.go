package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3StorageService stores files in AWS S3 (or compatible object storage).
type S3StorageService struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
	cdnDomain string
	// key prefix whitelist maps folders to S3 key prefixes
	allowedPrefixes map[string]bool
}

// S3Config holds configuration for the S3 storage service.
type S3Config struct {
	Bucket    string
	Region    string
	AccessKey string
	SecretKey string
	Endpoint  string
	CDNDomain string
}

// NewS3StorageService creates a new S3 storage service.
func NewS3StorageService(ctx context.Context, cfg S3Config) (*S3StorageService, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required for s3 storage driver")
	}

	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(cfg.Region))

	if cfg.Endpoint != "" {
		// For MinIO or other S3-compatible services
		opts = append(opts, config.WithBaseEndpoint(cfg.Endpoint))
	}

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.UsePathStyle = true
		}
	})

	return &S3StorageService{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    cfg.Bucket,
		cdnDomain: strings.TrimRight(cfg.CDNDomain, "/"),
		allowedPrefixes: map[string]bool{
			"images/": true, "documents/": true, "products/": true,
			"certifications/": true, "avatars/": true, "uploads/": true,
			"payment-proofs/": true, "kyb/": true, "order-messages/": true,
		},
	}, nil
}

// Upload stores a file in S3 and returns its URL.
func (s *S3StorageService) Upload(ctx context.Context, reader io.Reader, opts UploadOptions) (string, error) {
	ext := strings.ToLower(filepath.Ext(opts.FileName))
	if !allowedExtensions[ext] {
		return "", fmt.Errorf("file extension %s is not allowed", ext)
	}

	cleanFolder := strings.Trim(filepath.Clean(opts.Folder), "/.")
	if cleanFolder == "" {
		cleanFolder = "uploads"
	}
	keyPrefix := cleanFolder + "/"

	// Validate key prefix
	if !s.allowedPrefixes[keyPrefix] {
		keyPrefix = "uploads/"
	}

	// Read entire content into buffer for magic byte check and S3 upload
	uploadReader := reader
	if opts.MaxFileSize > 0 {
		uploadReader = io.LimitReader(reader, opts.MaxFileSize+1)
	}
	buf, err := io.ReadAll(uploadReader)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	if opts.MaxFileSize > 0 && int64(len(buf)) > opts.MaxFileSize {
		return "", fmt.Errorf("upload exceeds maximum size")
	}

	// Validate magic bytes
	if len(buf) > 0 {
		sample := buf
		if len(sample) > 512 {
			sample = sample[:512]
		}
		detectedType := http.DetectContentType(sample)
		if !contentTypeAllowed(detectedType, opts.FileName) {
			return "", fmt.Errorf("file content type %s is not allowed", detectedType)
		}
	}

	// Detect content type from bytes
	contentType := http.DetectContentType(buf)
	if len(buf) >= 512 {
		contentType = http.DetectContentType(buf[:512])
	}

	key := fmt.Sprintf("%s%s_%d%s", keyPrefix, uuid.New().String()[:8], time.Now().Unix(), ext)

	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(buf),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	if s.cdnDomain != "" {
		return fmt.Sprintf("%s/%s", s.cdnDomain, key), nil
	}
	return fmt.Sprintf("s3://%s/%s", s.bucket, key), nil
}

// Delete removes a file from S3 by its URL or key.
func (s *S3StorageService) Delete(ctx context.Context, url string) error {
	key := s.KeyFromURL(url)
	if key == "" {
		return fmt.Errorf("invalid S3 URL: %s", url)
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// GetPresignedURL returns a presigned GET URL valid for the specified duration.
func (s *S3StorageService) GetPresignedURL(ctx context.Context, url string, expiry time.Duration) (string, error) {
	key := s.KeyFromURL(url)
	if key == "" {
		return "", fmt.Errorf("invalid S3 URL: %s", url)
	}
	req, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("failed to presign URL: %w", err)
	}
	return req.URL, nil
}

// PresignUpload creates a direct PUT URL for browser-to-object-storage uploads.
func (s *S3StorageService) PresignUpload(ctx context.Context, opts PresignUploadOptions, expiry time.Duration) (*PresignedUpload, error) {
	if opts.SizeBytes <= 0 {
		return nil, fmt.Errorf("upload size must be positive")
	}
	if opts.MaxFileSize > 0 && opts.SizeBytes > opts.MaxFileSize {
		return nil, fmt.Errorf("upload exceeds maximum size")
	}
	ext := strings.ToLower(filepath.Ext(opts.FileName))
	if !allowedExtensions[ext] {
		return nil, fmt.Errorf("file extension %s is not allowed", ext)
	}
	contentType := normalizeContentType(opts.ContentType)
	if contentType == "" {
		return nil, fmt.Errorf("content type is required")
	}
	if !contentTypeAllowed(contentType, opts.FileName) {
		return nil, fmt.Errorf("content type %s is not allowed", contentType)
	}
	key := s.newKey(opts.Folder, ext)
	req, err := s.presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(opts.SizeBytes),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return nil, fmt.Errorf("failed to presign upload: %w", err)
	}
	return &PresignedUpload{
		Key:       key,
		URL:       req.URL,
		Method:    http.MethodPut,
		Headers:   map[string]string{"Content-Type": contentType},
		PublicURL: s.objectURL(key),
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

// Stat verifies an object exists in S3 and returns its metadata.
func (s *S3StorageService) Stat(ctx context.Context, url string) (*ObjectInfo, error) {
	key := s.KeyFromURL(url)
	if key == "" {
		return nil, fmt.Errorf("invalid S3 URL: %s", url)
	}
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:         key,
		URL:         s.objectURL(key),
		ContentType: aws.ToString(out.ContentType),
		SizeBytes:   aws.ToInt64(out.ContentLength),
	}, nil
}

func (s *S3StorageService) newKey(folder, ext string) string {
	cleanFolder := strings.Trim(filepath.Clean(folder), "/.")
	if cleanFolder == "" {
		cleanFolder = "uploads"
	}
	keyPrefix := cleanFolder + "/"
	if !s.allowedPrefixes[keyPrefix] {
		keyPrefix = "uploads/"
	}
	return fmt.Sprintf("%s%s_%d%s", keyPrefix, uuid.New().String()[:8], time.Now().Unix(), ext)
}

func (s *S3StorageService) objectURL(key string) string {
	if s.cdnDomain != "" {
		return fmt.Sprintf("%s/%s", s.cdnDomain, key)
	}
	return fmt.Sprintf("s3://%s/%s", s.bucket, key)
}

// KeyFromURL extracts the S3 key from a URL (s3://bucket/key, CDN URL, or /key).
func (s *S3StorageService) KeyFromURL(url string) string {
	// s3://bucket/key format
	if strings.HasPrefix(url, "s3://"+s.bucket+"/") {
		return strings.TrimPrefix(url, "s3://"+s.bucket+"/")
	}
	// CDN URL format
	if s.cdnDomain != "" && strings.HasPrefix(url, s.cdnDomain+"/") {
		return strings.TrimPrefix(url, s.cdnDomain+"/")
	}
	// Plain key
	if !strings.Contains(url, "://") {
		return url
	}
	return ""
}
