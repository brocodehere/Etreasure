package config

import (
	"log"
	"os"
)

type Config struct {
	DBURL            string
	JWTSecret        string
	RefreshSecret    string
	MediaBucketURL   string
	UploadDir        string
	UploadHMACSecret string
	RazorpayKeyID    string
	RazorpaySecret   string
	// Cloudflare R2 Configuration
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2S3Endpoint      string
	R2PublicBaseURL   string
	// Legacy AWS S3 Configuration (deprecated)
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	S3Bucket           string
	S3Endpoint         string // For custom S3 providers
}

func Load() Config {
	cfg := Config{
		DBURL:            os.Getenv("DATABASE_URL"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		RefreshSecret:    os.Getenv("REFRESH_SECRET"),
		MediaBucketURL:   os.Getenv("MEDIA_BUCKET_URL"),
		UploadDir:        os.Getenv("UPLOAD_DIR"),
		UploadHMACSecret: os.Getenv("UPLOAD_HMAC_SECRET"),
		RazorpayKeyID:    os.Getenv("RAZORPAY_KEY_ID"),
		RazorpaySecret:   os.Getenv("RAZORPAY_KEY_SECRET"),
		// Cloudflare R2 Configuration
		R2AccountID:       os.Getenv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     os.Getenv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: os.Getenv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      os.Getenv("R2_BUCKET_NAME"),
		R2S3Endpoint:      os.Getenv("R2_S3_ENDPOINT"),
		R2PublicBaseURL:   os.Getenv("R2_PUBLIC_BASE_URL"),
		// Legacy AWS S3 Configuration (deprecated)
		AWSAccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		AWSSecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		AWSRegion:          os.Getenv("AWS_REGION"),
		S3Bucket:           os.Getenv("S3_BUCKET"),
		S3Endpoint:         os.Getenv("S3_ENDPOINT"),
	}

	if cfg.DBURL == "" {
		log.Println("WARNING: DATABASE_URL is not set")
	}

	if cfg.UploadDir == "" {
		cfg.UploadDir = "uploads"
	}
	if cfg.UploadHMACSecret == "" {
		cfg.UploadHMACSecret = "dev-upload-secret"
	}

	return cfg
}
