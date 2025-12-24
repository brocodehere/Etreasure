package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type NewsletterSubscriber struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	SubscribedAt time.Time `json:"subscribed_at"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateNewsletterSubscriberRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type Banner struct {
	ID              uuid.UUID  `json:"id"`
	Title           string     `json:"title"`
	DesktopImageKey string     `json:"desktop_image_key"`
	DesktopImageURL string     `json:"desktop_image_url"`
	LaptopImageKey  string     `json:"laptop_image_key"`
	LaptopImageURL  string     `json:"laptop_image_url"`
	MobileImageKey  string     `json:"mobile_image_key"`
	MobileImageURL  string     `json:"mobile_image_url"`
	LinkURL         *string    `json:"link_url,omitempty"`
	IsActive        bool       `json:"is_active"`
	SortOrder       int        `json:"sort_order"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateBannerRequest struct {
	Title           string     `json:"title" binding:"required"`
	DesktopImageURL string     `json:"desktop_image_url" binding:"required"`
	LaptopImageURL  string     `json:"laptop_image_url"`
	MobileImageURL  string     `json:"mobile_image_url"`
	LinkURL         *string    `json:"link_url,omitempty"`
	IsActive        bool       `json:"is_active"`
	SortOrder       int        `json:"sort_order"`
	StartsAt        *time.Time `json:"starts_at,omitempty"`
	EndsAt          *time.Time `json:"ends_at,omitempty"`
}

type UpdateBannerRequest struct {
	Title           *string    `json:"title"`
	DesktopImageURL *string    `json:"desktop_image_url"`
	LaptopImageURL  *string    `json:"laptop_image_url"`
	MobileImageURL  *string    `json:"mobile_image_url"`
	LinkURL         *string    `json:"link_url"`
	IsActive        *bool      `json:"is_active"`
	SortOrder       *int       `json:"sort_order"`
	StartsAt        *time.Time `json:"starts_at"`
	EndsAt          *time.Time `json:"ends_at"`
}

// ListPublicBanners returns active banners for the storefront
func (h *Handler) ListPublicBanners(c *gin.Context) {
	now := time.Now().UTC()

	query := `
		SELECT id, title, link_url, is_active, sort_order, starts_at, ends_at, created_at, updated_at,
		       desktop_image_url, laptop_image_url, mobile_image_url
		FROM banners
		WHERE is_active = true 
			AND (starts_at IS NULL OR starts_at <= $1)
			AND (ends_at IS NULL OR ends_at >= $1)
		ORDER BY sort_order ASC, created_at DESC
		LIMIT 3
	`

	rows, err := h.DB.Query(c, query, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var banners []Banner
	for rows.Next() {
		var b Banner
		var desktopImageURL, laptopImageURL, mobileImageURL *string
		err := rows.Scan(&b.ID, &b.Title, &b.LinkURL, &b.IsActive, &b.SortOrder, &b.StartsAt, &b.EndsAt, &b.CreatedAt, &b.UpdatedAt,
			&desktopImageURL, &laptopImageURL, &mobileImageURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Set image URLs with fallbacks
		if desktopImageURL != nil && *desktopImageURL != "" {
			b.DesktopImageURL = *desktopImageURL
		} else {
			b.DesktopImageURL = "https://pub-1a3924a6c6994107be6fe9f3ed794c0a.r2.dev/banner-placeholder.webp"
		}

		if laptopImageURL != nil && *laptopImageURL != "" {
			b.LaptopImageURL = *laptopImageURL
		} else {
			b.LaptopImageURL = b.DesktopImageURL // Fallback to desktop
		}

		if mobileImageURL != nil && *mobileImageURL != "" {
			b.MobileImageURL = *mobileImageURL
		} else {
			b.MobileImageURL = b.DesktopImageURL // Fallback to desktop
		}

		banners = append(banners, b)
	}

	c.JSON(http.StatusOK, gin.H{
		"items": banners,
	})
}

func (h *Handler) CreateNewsletterSubscriber(c *gin.Context) {
	var req CreateNewsletterSubscriberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	now := time.Now().UTC()

	// Check if email already exists
	var existingCount int
	err := h.DB.QueryRow(c, "SELECT COUNT(*) FROM newsletter_subscribers WHERE email = $1", req.Email).Scan(&existingCount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if existingCount > 0 {
		// Email already subscribed - just return success
		c.JSON(http.StatusCreated, gin.H{
			"message": "Email already subscribed!",
			"email":   req.Email,
		})
	} else {
		// New subscriber - insert with only the columns that exist
		_, err = h.DB.Exec(c, `
			INSERT INTO newsletter_subscribers (email, subscribed_at)
			VALUES ($1, $2)
		`, req.Email, now)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to subscribe"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "Successfully subscribed to newsletter!",
			"email":   req.Email,
		})
	}
}

// Admin banner management methods
func (h *Handler) ListBanners(c *gin.Context) {
	ctx := c.Request.Context()

	query := `
		SELECT id, title, link_url, is_active, sort_order, starts_at, ends_at, created_at, updated_at,
		       desktop_image_url, laptop_image_url, mobile_image_url
		FROM banners
		ORDER BY sort_order ASC, created_at DESC
	`

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var banners []Banner
	for rows.Next() {
		var b Banner
		var desktopImageURL, laptopImageURL, mobileImageURL *string
		err := rows.Scan(&b.ID, &b.Title, &b.LinkURL, &b.IsActive, &b.SortOrder, &b.StartsAt, &b.EndsAt, &b.CreatedAt, &b.UpdatedAt,
			&desktopImageURL, &laptopImageURL, &mobileImageURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Set image URLs with fallbacks
		if desktopImageURL != nil && *desktopImageURL != "" {
			b.DesktopImageURL = *desktopImageURL
		} else {
			b.DesktopImageURL = "https://pub-1a3924a6c6994107be6fe9f3ed794c0a.r2.dev/banner-placeholder.webp"
		}

		if laptopImageURL != nil && *laptopImageURL != "" {
			b.LaptopImageURL = *laptopImageURL
		} else {
			b.LaptopImageURL = b.DesktopImageURL // Fallback to desktop
		}

		if mobileImageURL != nil && *mobileImageURL != "" {
			b.MobileImageURL = *mobileImageURL
		} else {
			b.MobileImageURL = b.DesktopImageURL // Fallback to desktop
		}

		banners = append(banners, b)
	}

	c.JSON(http.StatusOK, gin.H{
		"data": banners,
	})
}

func (h *Handler) CreateBanner(c *gin.Context) {
	ctx := c.Request.Context()

	var req CreateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()
	bannerID := uuid.New()

	query := `
		INSERT INTO banners (id, title, desktop_image_url, laptop_image_url, mobile_image_url, link_url, is_active, sort_order, starts_at, ends_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, title, desktop_image_url, laptop_image_url, mobile_image_url, link_url, is_active, sort_order, starts_at, ends_at, created_at, updated_at
	`

	var banner Banner
	err := h.DB.QueryRow(ctx, query,
		bannerID,
		req.Title,
		req.DesktopImageURL,
		req.LaptopImageURL,
		req.MobileImageURL,
		req.LinkURL,
		req.IsActive,
		req.SortOrder,
		req.StartsAt,
		req.EndsAt,
		now,
		now,
	).Scan(
		&banner.ID,
		&banner.Title,
		&banner.DesktopImageURL,
		&banner.LaptopImageURL,
		&banner.MobileImageURL,
		&banner.LinkURL,
		&banner.IsActive,
		&banner.SortOrder,
		&banner.StartsAt,
		&banner.EndsAt,
		&banner.CreatedAt,
		&banner.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create banner: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, banner)
}

func (h *Handler) GetBanner(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	_, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	query := `
		SELECT id, title, desktop_image_url, laptop_image_url, mobile_image_url, link_url, is_active, sort_order, starts_at, ends_at, created_at, updated_at
		FROM banners
		WHERE id = $1
	`

	var banner Banner
	var desktopImageURL, laptopImageURL, mobileImageURL *string
	err = h.DB.QueryRow(ctx, query, idStr).Scan(
		&banner.ID,
		&banner.Title,
		&desktopImageURL,
		&laptopImageURL,
		&mobileImageURL,
		&banner.LinkURL,
		&banner.IsActive,
		&banner.SortOrder,
		&banner.StartsAt,
		&banner.EndsAt,
		&banner.CreatedAt,
		&banner.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "banner not found"})
		return
	}

	// Set image URLs with fallbacks
	if desktopImageURL != nil && *desktopImageURL != "" {
		banner.DesktopImageURL = *desktopImageURL
	} else {
		banner.DesktopImageURL = "https://pub-1a3924a6c6994107be6fe9f3ed794c0a.r2.dev/banner-placeholder.webp"
	}

	if laptopImageURL != nil && *laptopImageURL != "" {
		banner.LaptopImageURL = *laptopImageURL
	} else {
		banner.LaptopImageURL = banner.DesktopImageURL // Fallback to desktop
	}

	if mobileImageURL != nil && *mobileImageURL != "" {
		banner.MobileImageURL = *mobileImageURL
	} else {
		banner.MobileImageURL = banner.DesktopImageURL // Fallback to desktop
	}

	c.JSON(http.StatusOK, banner)
}

func (h *Handler) UpdateBanner(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	_, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	var req UpdateBannerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build dynamic UPDATE query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}
	if req.DesktopImageURL != nil {
		setParts = append(setParts, fmt.Sprintf("desktop_image_url = $%d", argIndex))
		args = append(args, *req.DesktopImageURL)
		argIndex++
	}
	if req.LaptopImageURL != nil {
		setParts = append(setParts, fmt.Sprintf("laptop_image_url = $%d", argIndex))
		args = append(args, *req.LaptopImageURL)
		argIndex++
	}
	if req.MobileImageURL != nil {
		setParts = append(setParts, fmt.Sprintf("mobile_image_url = $%d", argIndex))
		args = append(args, *req.MobileImageURL)
		argIndex++
	}
	if req.LinkURL != nil {
		setParts = append(setParts, fmt.Sprintf("link_url = $%d", argIndex))
		args = append(args, *req.LinkURL)
		argIndex++
	}
	if req.IsActive != nil {
		setParts = append(setParts, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}
	if req.SortOrder != nil {
		setParts = append(setParts, fmt.Sprintf("sort_order = $%d", argIndex))
		args = append(args, *req.SortOrder)
		argIndex++
	}
	if req.StartsAt != nil {
		setParts = append(setParts, fmt.Sprintf("starts_at = $%d", argIndex))
		args = append(args, *req.StartsAt)
		argIndex++
	}
	if req.EndsAt != nil {
		setParts = append(setParts, fmt.Sprintf("ends_at = $%d", argIndex))
		args = append(args, *req.EndsAt)
		argIndex++
	}

	if len(setParts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	// Add updated_at and id
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now().UTC())
	argIndex++
	args = append(args, idStr)

	query := fmt.Sprintf(`
		UPDATE banners
		SET %s
		WHERE id = $%d
		RETURNING id, title, desktop_image_url, laptop_image_url, mobile_image_url, link_url, is_active, sort_order, starts_at, ends_at, created_at, updated_at
	`, strings.Join(setParts, ", "), argIndex)

	var banner Banner
	var desktopImageURL, laptopImageURL, mobileImageURL *string
	err = h.DB.QueryRow(ctx, query, args...).Scan(
		&banner.ID,
		&banner.Title,
		&desktopImageURL,
		&laptopImageURL,
		&mobileImageURL,
		&banner.LinkURL,
		&banner.IsActive,
		&banner.SortOrder,
		&banner.StartsAt,
		&banner.EndsAt,
		&banner.CreatedAt,
		&banner.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update banner: " + err.Error()})
		return
	}

	// Set image URLs with fallbacks
	if desktopImageURL != nil && *desktopImageURL != "" {
		banner.DesktopImageURL = *desktopImageURL
	} else {
		banner.DesktopImageURL = "https://pub-1a3924a6c6994107be6fe9f3ed794c0a.r2.dev/banner-placeholder.webp"
	}

	if laptopImageURL != nil && *laptopImageURL != "" {
		banner.LaptopImageURL = *laptopImageURL
	} else {
		banner.LaptopImageURL = banner.DesktopImageURL // Fallback to desktop
	}

	if mobileImageURL != nil && *mobileImageURL != "" {
		banner.MobileImageURL = *mobileImageURL
	} else {
		banner.MobileImageURL = banner.DesktopImageURL // Fallback to desktop
	}

	c.JSON(http.StatusOK, banner)
}

func (h *Handler) DeleteBanner(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")

	_, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}

	// First, get the banner to extract image URLs for R2 deletion
	var desktopImageURL, laptopImageURL, mobileImageURL *string
	err = h.DB.QueryRow(ctx, `SELECT desktop_image_url, laptop_image_url, mobile_image_url FROM banners WHERE id=$1`, idStr).Scan(&desktopImageURL, &laptopImageURL, &mobileImageURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "banner not found"})
		return
	}

	// Delete banner from DB
	_, err = h.DB.Exec(ctx, `DELETE FROM banners WHERE id=$1`, idStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	// Delete images from R2 if they are R2 paths
	imageURLs := []*string{desktopImageURL, laptopImageURL, mobileImageURL}
	for _, imageURL := range imageURLs {
		if imageURL != nil && h.R2Client != nil && (strings.HasPrefix(*imageURL, "product/") || strings.HasPrefix(*imageURL, "banner/") || strings.HasPrefix(*imageURL, "category/")) {
			if err := h.R2Client.DeleteObject(ctx, *imageURL); err != nil {
				log.Printf("Warning: Failed to delete from R2: %v", err)
			}
		}
	}

	c.Status(http.StatusNoContent)
}
