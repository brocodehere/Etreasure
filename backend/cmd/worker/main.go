package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type ImageProcessingJob struct {
	ID       string `json:"id"`
	Key      string `json:"key"`
	Bucket   string `json:"bucket"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Quality  int    `json:"quality"`
	Format   string `json:"format"`
	ResizeOp string `json:"resize_op"` // fit | fill | crop
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	s3Client := s3.NewFromConfig(awsCfg)

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	// Process jobs in a simple loop (in production, use a proper queue)
	for {
		processPendingJobs(s3Client, uploadDir)
		time.Sleep(5 * time.Second)
	}
}

func processPendingJobs(s3Client *s3.Client, uploadDir string) {
	// This is a simplified implementation
	// In production, you'd use SQS, RabbitMQ, or similar
	// For now, we'll process files in a "processing" directory

	processingDir := filepath.Join(uploadDir, "processing")
	if _, err := os.Stat(processingDir); os.IsNotExist(err) {
		os.MkdirAll(processingDir, 0755)
		return
	}

	files, err := os.ReadDir(processingDir)
	if err != nil {
		log.Printf("Error reading processing directory: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		jobPath := filepath.Join(processingDir, file.Name())
		processJob(s3Client, jobPath, uploadDir)
	}
}

func processJob(s3Client *s3.Client, jobPath, uploadDir string) {
	data, err := os.ReadFile(jobPath)
	if err != nil {
		log.Printf("Error reading job file %s: %v", jobPath, err)
		return
	}

	var job ImageProcessingJob
	if err := json.Unmarshal(data, &job); err != nil {
		log.Printf("Error parsing job file %s: %v", jobPath, err)
		return
	}

	log.Printf("Processing image job: %+v", job)

	// Download original image from S3 or local
	var imgData []byte
	if job.Bucket != "" {
		// Download from S3
		resp, err := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
			Bucket: &job.Bucket,
			Key:    &job.Key,
		})
		if err != nil {
			log.Printf("Error downloading from S3: %v", err)
			return
		}
		defer resp.Body.Close()
		imgData, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading S3 object: %v", err)
			return
		}
	} else {
		// Local file
		localPath := filepath.Join(uploadDir, job.Key)
		imgData, err = os.ReadFile(localPath)
		if err != nil {
			log.Printf("Error reading local file %s: %v", localPath, err)
			return
		}
	}

	// Process image
	processedImg, err := processImage(imgData, job)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		return
	}

	// Save processed image
	outputKey := generateOutputKey(job.Key, job)
	outputPath := filepath.Join(uploadDir, outputKey)

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		return
	}

	if err := os.WriteFile(outputPath, processedImg, 0644); err != nil {
		log.Printf("Error saving processed image: %v", err)
		return
	}

	// If using S3, upload processed image
	if job.Bucket != "" {
		_, err = s3Client.PutObject(context.Background(), &s3.PutObjectInput{
			Bucket: &job.Bucket,
			Key:    &outputKey,
			Body:   bytes.NewReader(processedImg),
		})
		if err != nil {
			log.Printf("Error uploading processed image to S3: %v", err)
			return
		}
	}

	// Mark job as complete
	if err := os.Remove(jobPath); err != nil {
		log.Printf("Error removing job file: %v", err)
	}

	log.Printf("Successfully processed image: %s", outputKey)
}

func processImage(data []byte, job ImageProcessingJob) ([]byte, error) {
	img, err := imaging.Decode(strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize based on operation
	var resized image.Image
	switch job.ResizeOp {
	case "fit":
		resized = imaging.Fit(img, job.Width, job.Height, imaging.Lanczos)
	case "fill":
		resized = imaging.Fill(img, job.Width, job.Height, imaging.Center, imaging.Lanczos)
	case "crop":
		resized = imaging.CropCenter(img, job.Width, job.Height)
	default:
		resized = imaging.Resize(img, job.Width, job.Height, imaging.Lanczos)
	}

	// Encode with quality
	var buf bytes.Buffer
	switch job.Format {
	case "png":
		err = imaging.Encode(&buf, resized, imaging.PNG)
	case "webp":
		err = imaging.Encode(&buf, resized, imaging.webp)
	default:
		err = imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(job.Quality))
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

func generateOutputKey(originalKey string, job ImageProcessingJob) string {
	ext := filepath.Ext(originalKey)
	base := strings.TrimSuffix(originalKey, ext)

	suffix := fmt.Sprintf("_%dx%d", job.Width, job.Height)
	if job.Quality != 85 {
		suffix += fmt.Sprintf("_q%d", job.Quality)
	}
	if job.Format != "" && job.Format != "jpg" {
		suffix += "." + job.Format
	} else {
		suffix += ext
	}

	return base + suffix
}
