package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client
var PublicClient *minio.Client
var BucketName string

type Config struct {
	EndPoint       string
	PublicEndPoint string // externally-reachable host:port used to sign URLs returned to clients; defaults to EndPoint
	AccessKey      string
	SecretKey      string
	Bucket         string
	UseSSL         bool
	PublicUseSSL   bool
}

var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

var allowedContentTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// magic bytes for image type detection
var imageMagicBytes = []struct {
	prefix []byte
	ext    string
	ct     string
}{
	{[]byte{0xFF, 0xD8, 0xFF}, ".jpg", "image/jpeg"},
	{[]byte{0x89, 0x50, 0x4E, 0x47}, ".png", "image/png"},
	{[]byte{0x52, 0x49, 0x46, 0x46, 0x00, 0x00, 0x00, 0x00, 0x57, 0x45, 0x42, 0x50}, ".webp", "image/webp"}, // RIFF....WEBP
	{[]byte{0x47, 0x49, 0x46, 0x38}, ".gif", "image/gif"},                                                   // GIF8
}

func Init(cfg Config) error {
	BucketName = cfg.Bucket

	var err error
	Client, err = minio.New(cfg.EndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return fmt.Errorf("failed to create minio client: %w", err)
	}

	publicEndPoint, publicUseSSL := cfg.PublicEndPoint, cfg.PublicUseSSL
	if publicEndPoint == "" {
		publicEndPoint, publicUseSSL = cfg.EndPoint, cfg.UseSSL
	}
	// Region must be set explicitly: without it, PresignedGetObject makes a live
	// GetBucketLocation call against PublicEndPoint, which is typically unreachable
	// from inside the container (e.g. it points at the host's "localhost").
	PublicClient, err = minio.New(publicEndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: publicUseSSL,
		Region: "us-east-1",
	})
	if err != nil {
		return fmt.Errorf("failed to create public minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := Client.BucketExists(ctx, BucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = Client.MakeBucket(ctx, BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Printf("Created bucket: %s", BucketName)
	}

	log.Printf("Connected to S3 storage at %s", cfg.EndPoint)
	return nil
}

func detectImageType(header *multipart.FileHeader, file multipart.File) (ext string, contentType string, err error) {
	buf := make([]byte, 12)
	n, readErr := io.ReadFull(file, buf)
	buf = buf[:n]
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		return "", "", fmt.Errorf("failed to read file header: %w", readErr)
	}

	for _, m := range imageMagicBytes {
		if bytes.HasPrefix(buf, m.prefix) {
			ext = m.ext
			contentType = m.ct
			break
		}
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", "", fmt.Errorf("failed to seek file: %w", err)
	}

	if ext == "" {
		ext = strings.ToLower(filepath.Ext(header.Filename))
		if !allowedExtensions[ext] {
			return "", "", fmt.Errorf("unsupported file type: %s", ext)
		}
		ct := header.Header.Get("Content-Type")
		if allowedContentTypes[ct] {
			contentType = ct
		} else {
			contentType = "application/octet-stream"
		}
	}

	return ext, contentType, nil
}

func UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader) (string, error) {
	if header.Size <= 0 {
		return "", fmt.Errorf("empty or invalid file")
	}

	if header.Size > 10<<20 {
		return "", fmt.Errorf("file too large: max 10MB")
	}

	ext, contentType, err := detectImageType(header, file)
	if err != nil {
		return "", err
	}

	objectName := fmt.Sprintf("products/%s%s", uuid.New().String(), ext)

	_, err = Client.PutObject(ctx, BucketName, objectName, file, header.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return objectName, nil
}

func GetPresignedURL(ctx context.Context, objectName string) (string, error) {
	presignedURL, err := PublicClient.PresignedGetObject(ctx, BucketName, objectName, 24*time.Hour, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func DeleteImage(ctx context.Context, objectName string) error {
	if !strings.HasPrefix(objectName, "products/") {
		return fmt.Errorf("invalid object path: must start with products/")
	}
	err := Client.RemoveObject(ctx, BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func GetObject(ctx context.Context, objectName string) (io.ReadCloser, error) {
	if !strings.HasPrefix(objectName, "products/") {
		return nil, fmt.Errorf("invalid object path: must start with products/")
	}
	obj, err := Client.GetObject(ctx, BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	return obj, nil
}
