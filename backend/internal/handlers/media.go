package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/etreasure/backend/internal/config"
	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaHandler struct {
	DB         *pgxpool.Pool
	UploadDir  string
	HMACSecret string
	R2Client   *storage.R2Client
	Config     config.Config
}

type PresignRequest struct {
	Filename string `json:"filename" binding:"required"`
	MimeType string `json:"mime_type"`
	Type     string `json:"type"` // "product", "banner", "category"
}

type PresignResponse struct {
	ID        string `json:"id"`
	UploadURL string `json:"uploadUrl"`
	Path      string `json:"path"`
	ExpiresAt int64  `json:"expiresAt"`
}

type UploadRequest struct {
	File multipart.File `form:"file" binding:"required"`
	Type string         `form:"type" binding:"required"` // "product", "banner", "category"
}

type UploadResponse struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

func (h *MediaHandler) sign(id string) string {
	mac := hmac.New(sha256.New, []byte(h.HMACSecret))
	mac.Write([]byte(id))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *MediaHandler) verify(id, token string) bool {
	return h.sign(id) == token
}

// POST /api/admin/media/presign (legacy - for backward compatibility)
func (h *MediaHandler) Presign(c *gin.Context) {
	var req PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	id := uuid.New().String()
	token := h.sign(id)
	uploadURL := fmt.Sprintf("/api/admin/media/upload/%s?token=%s", id, token)
	path := fmt.Sprintf("/uploads/%s", id)
	c.JSON(http.StatusOK, PresignResponse{ID: id, UploadURL: uploadURL, Path: path, ExpiresAt: time.Now().Add(10 * time.Minute).Unix()})
}

// POST /api/admin/media/upload (new R2 upload endpoint)
func (h *MediaHandler) UploadR2(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	mediaType := c.PostForm("type")
	if mediaType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required (product, banner, category)"})
		return
	}

	// Validate file size (5MB limit)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file size exceeds 5MB limit"})
		return
	}

	// Validate content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Detect from file content
		buffer := make([]byte, 512)
		file.Read(buffer)
		file.Seek(0, 0) // Reset file pointer
		contentType = http.DetectContentType(buffer)
	}

	// Check if it's an allowed image type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/webp": true,
		"image/avif": true,
	}

	if !allowedTypes[contentType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type. Only JPEG, PNG, WebP, and AVIF are allowed"})
		return
	}

	// Generate unique key for R2
	key := h.R2Client.GenerateKey(mediaType, header.Filename)

	// Upload to R2
	uploadedKey, err := h.R2Client.UploadObject(c.Request.Context(), key, file, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload to R2"})
		return
	}

	// Store in media database
	var width, height int
	if strings.HasPrefix(contentType, "image/") {
		// Read file again for dimension detection
		file.Seek(0, 0)
		bodyBytes, _ := io.ReadAll(file)
		if img, _, err := image.DecodeConfig(bytes.NewReader(bodyBytes)); err == nil {
			width, height = img.Width, img.Height
		}
	}

	var mediaID int
	err = h.DB.QueryRow(c.Request.Context(),
		`INSERT INTO media (path, original_filename, mime_type, file_size_bytes, width, height, variants)
		VALUES ($1,$2,$3,$4,$5,$6,'{}'::jsonb) RETURNING id`,
		uploadedKey, header.Filename, contentType, header.Size, width, height).Scan(&mediaID)

	if err != nil {
		// If DB insert fails, try to delete from R2
		h.R2Client.DeleteObject(c.Request.Context(), uploadedKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store media record"})
		return
	}

	publicURL := h.R2Client.PublicURL(uploadedKey)
	c.JSON(http.StatusOK, UploadResponse{
		Key: uploadedKey,
		URL: publicURL,
	})
}

// PUT /api/admin/media/upload/:id?token=... (legacy - for backward compatibility)
func (h *MediaHandler) Upload(c *gin.Context) {
	id := c.Param("id")
	token := c.Query("token")
	if id == "" || token == "" || !h.verify(id, token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid upload token"})
		return
	}
	if err := os.MkdirAll(h.UploadDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload dir missing"})
		return
	}
	// Read entire body (for dev). For large files stream to disk.
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "read body failed"})
		return
	}
	ct := c.GetHeader("Content-Type")
	if ct == "" {
		ct = http.DetectContentType(body)
	}
	exts, _ := mime.ExtensionsByType(ct)
	ext := ".bin"
	if len(exts) > 0 {
		ext = exts[0]
	}
	filename := id + ext
	full := filepath.Join(h.UploadDir, filename)
	if err := os.WriteFile(full, body, 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
		return
	}
	// Try to decode image to get dimensions (jpeg/png/gif)
	var width, height int
	if strings.HasPrefix(ct, "image/") {
		if img, _, err := image.DecodeConfig(bytes.NewReader(body)); err == nil {
			width, height = img.Width, img.Height
		}
	}
	ctx := c.Request.Context()
	var mediaID int
	err = h.DB.QueryRow(ctx, `INSERT INTO media (path, original_filename, mime_type, file_size_bytes, width, height, variants)
		VALUES ($1,$2,$3,$4,$5,$6,'{}'::jsonb) RETURNING id`,
		"/uploads/"+filename, filename, ct, len(body), width, height).Scan(&mediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db insert failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": mediaID, "path": "/uploads/" + filename, "mime_type": ct, "width": width, "height": height})
}

// GET /api/admin/media?first=&after=
func (h *MediaHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	first := 1000 // Set a high default to effectively show all images
	if v := c.Query("first"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			first = n
		}
	}
	after := 0
	if v := c.Query("after"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			after = n
		}
	}
	rows, err := h.DB.Query(ctx, `SELECT id, path, mime_type, file_size_bytes, width, height, created_at FROM media WHERE id > $1 ORDER BY created_at DESC, id DESC LIMIT $2`, after, first)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	type item struct {
		ID        int       `json:"id"`
		Path      string    `json:"path"`
		URL       string    `json:"url"`
		Mime      string    `json:"mime_type"`
		Size      int64     `json:"file_size_bytes"`
		Width     *int      `json:"width"`
		Height    *int      `json:"height"`
		CreatedAt time.Time `json:"created_at"`
	}
	items := make([]item, 0)
	for rows.Next() {
		var it item
		var width, height *int
		if err := rows.Scan(&it.ID, &it.Path, &it.Mime, &it.Size, &width, &height, &it.CreatedAt); err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "scan failed"})
			return
		}
		it.Width = width
		it.Height = height

		// Generate proper URL based on path type
		if strings.HasPrefix(it.Path, "product/") || strings.HasPrefix(it.Path, "banner/") || strings.HasPrefix(it.Path, "category/") {
			// R2 path - use base64 encoded key in proxy URL to avoid slash issues
			encodedKey := strings.ReplaceAll(it.Path, "/", "_")
			it.URL = fmt.Sprintf("http://localhost:8080/api/public/media/%s", encodedKey)
		} else {
			// Legacy local path
			it.URL = it.Path
		}

		items = append(items, it)
	}
	var nextCursor *int
	if len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = &last.ID
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "nextCursor": nextCursor})
}

// DELETE /api/admin/media/:id
func (h *MediaHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	var path string
	err := h.DB.QueryRow(ctx, `DELETE FROM media WHERE id=$1 RETURNING path`, id).Scan(&path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Delete from R2 if it's an R2 path
	if h.R2Client != nil && (strings.HasPrefix(path, "product/") || strings.HasPrefix(path, "banner/") || strings.HasPrefix(path, "category/")) {
		if err := h.R2Client.DeleteObject(ctx, path); err != nil {
			// Log error but don't fail the request
			log.Printf("Warning: Failed to delete from R2: %v", err)
		}
	}

	// Remove local file if it's a local path
	if strings.HasPrefix(path, "/uploads/") {
		_ = os.Remove(filepath.Join(h.UploadDir, filepath.Base(path)))
	}

	c.Status(http.StatusNoContent)
}
