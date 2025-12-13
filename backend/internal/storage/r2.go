package storage

import (
	"context"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/etreasure/backend/internal/config"
	"github.com/google/uuid"
)

type R2Client struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

func NewR2Client(cfg config.Config) (*R2Client, error) {
	// Create AWS SDK config with custom endpoint
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     cfg.R2AccessKeyID,
				SecretAccessKey: cfg.R2SecretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS config: %w", err)
	}

	// Create S3 client with custom endpoint for R2
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.R2S3Endpoint)
		o.UsePathStyle = true // R2 requires path-style addressing
	})

	return &R2Client{
		client:     client,
		bucketName: cfg.R2BucketName,
		publicURL:  cfg.R2PublicBaseURL,
	}, nil
}

// UploadObject uploads a file to R2 and returns the object key
func (r *R2Client) UploadObject(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	// Detect content type if not provided
	if contentType == "" {
		// Try to detect from file extension
		ext := strings.ToLower(filepath.Ext(key))
		contentType = mime.TypeByExtension(ext)
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload object to R2: %w", err)
	}

	return key, nil
}

// DeleteObject deletes an object from R2
func (r *R2Client) DeleteObject(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete object from R2: %w", err)
	}

	return nil
}

// GetObject retrieves an object from R2
func (r *R2Client) GetObject(ctx context.Context, key string) (*s3.GetObjectOutput, error) {
	result, err := r.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(r.bucketName),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get object from R2: %w", err)
	}

	return result, nil
}

// PublicURL returns the public URL for an object key with proper URL encoding
func (r *R2Client) PublicURL(key string) string {
	if key == "" {
		return ""
	}

	// Ensure base URL doesn't end with slash and key doesn't start with slash
	baseURL := strings.TrimSuffix(r.publicURL, "/")
	cleanKey := strings.TrimPrefix(key, "/")

	// Don't encode the entire path, just encode special characters in the key
	// This preserves the directory structure while making the URL safe
	encodedKey := strings.ReplaceAll(cleanKey, " ", "%20")
	encodedKey = strings.ReplaceAll(encodedKey, "&", "%26")
	encodedKey = strings.ReplaceAll(encodedKey, "?", "%3F")
	encodedKey = strings.ReplaceAll(encodedKey, "#", "%23")

	finalURL := fmt.Sprintf("%s/%s", baseURL, encodedKey)
	return finalURL
}

// GenerateKey generates a unique key for a given media type
func (r *R2Client) GenerateKey(mediaType string, filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		// Default to .webp if no extension
		ext = ".webp"
	}

	uuid := uuid.New().String()
	return fmt.Sprintf("%s/%s%s", mediaType, uuid, ext)
}

// PresignURL generates a presigned URL for direct upload (alternative flow)
func (r *R2Client) PresignURL(ctx context.Context, key string, contentType string, expiresIn time.Duration) (string, error) {
	presigner := s3.NewPresignClient(r.client)

	req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiresIn
	})

	if err != nil {
		return "", fmt.Errorf("failed to presign URL: %w", err)
	}

	return req.URL, nil
}
