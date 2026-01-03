package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoriesHandler struct {
	DB          *pgxpool.Pool
	R2Client    *storage.R2Client
	ImageHelper *storage.ImageURLHelper
}

// Helper function to convert image path to public URL
// Helper function to convert image path to public URL
func (h *CategoriesHandler) formatImageURL(imagePath *string) *string {
	if imagePath == nil || h.R2Client == nil {
		return imagePath
	}

	path := *imagePath
	// If it's an R2 key (starts with product/, banner/, category/), convert to public URL
	if strings.HasPrefix(path, "product/") || strings.HasPrefix(path, "banner/") || strings.HasPrefix(path, "category/") {
		url := h.R2Client.PublicURL(path)
		return &url
	}
	// If it's a local path starting with /uploads/, convert to full URL
	if strings.HasPrefix(path, "/uploads/") {
		url := "https://etreasure-1.onrender.com" + path
		return &url
	}
	// If it's already a full URL, keep as is
	if strings.HasPrefix(path, "http") {
		return imagePath
	}
	// Otherwise, assume it's an R2 path and convert
	url := h.R2Client.PublicURL(path)
	return &url
}

type Category struct {
	UUIDID      uuid.UUID  `json:"uuid_id"`
	Slug        string     `json:"slug"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	SortOrder   int        `json:"sort_order"`
	ImageID     *int       `json:"image_id,omitempty"`
	ImagePath   *string    `json:"image_path,omitempty"`
	ImageURL    *string    `json:"image_url,omitempty"`
}

type UpsertCategoryRequest struct {
	Slug        string     `json:"slug" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   *int       `json:"sort_order"`
	ImageID     *int       `json:"image_id"`
}

// GET /api/admin/categories
func (h *CategoriesHandler) List(c *gin.Context) {
	ctx := c.Request.Context()
	rows, err := h.DB.Query(ctx, `
		SELECT c.uuid_id, c.slug, c.name, c.description, c.parent_id, c.sort_order, c.image_id, m.path
		FROM categories c
		LEFT JOIN media m ON c.image_id = m.id
		ORDER BY c.parent_id NULLS FIRST, c.sort_order, c.uuid_id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()
	var items []Category
	for rows.Next() {
		var it Category
		if err := rows.Scan(&it.UUIDID, &it.Slug, &it.Name, &it.Description, &it.ParentID, &it.SortOrder, &it.ImageID, &it.ImagePath); err == nil {
			// Use ImageHelper to format image URL
			if h.ImageHelper != nil {
				it.ImageURL = h.ImageHelper.FormatImageURL(it.ImagePath)
			}
			items = append(items, it)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// POST /api/admin/categories
func (h *CategoriesHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req UpsertCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	var uuidID uuid.UUID
	err := h.DB.QueryRow(ctx, `INSERT INTO categories (slug, name, description, parent_id, sort_order, image_id) VALUES ($1,$2,$3,$4,COALESCE($5,0),$6) RETURNING uuid_id`,
		req.Slug, req.Name, req.Description, req.ParentID, req.SortOrder, req.ImageID).Scan(&uuidID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"uuid_id": uuidID})
}

// PUT /api/admin/categories/:id
func (h *CategoriesHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	uuidID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req UpsertCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	_, err = h.DB.Exec(ctx, `UPDATE categories SET slug=$1,name=$2,description=$3,parent_id=$4,sort_order=COALESCE($5,sort_order),image_id=$6 WHERE uuid_id=$7`,
		req.Slug, req.Name, req.Description, req.ParentID, req.SortOrder, req.ImageID, uuidID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /api/admin/categories/:id
func (h *CategoriesHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	uuidID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// First, get the category's image_id for R2 deletion
	var imageID *int
	err = h.DB.QueryRow(ctx, `SELECT image_id FROM categories WHERE uuid_id=$1`, uuidID).Scan(&imageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "category not found"})
		return
	}

	// Delete category
	_, err = h.DB.Exec(ctx, `DELETE FROM categories WHERE uuid_id=$1`, uuidID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	// Delete media from R2 if category had an image
	if h.R2Client != nil && imageID != nil {
		var path string
		err := h.DB.QueryRow(ctx, `SELECT path FROM media WHERE id=$1`, *imageID).Scan(&path)
		if err == nil && (strings.HasPrefix(path, "product/") || strings.HasPrefix(path, "banner/") || strings.HasPrefix(path, "category/")) {
			if err := h.R2Client.DeleteObject(ctx, path); err != nil {
				log.Printf("Warning: Failed to delete from R2: %v", err)
			}
		}
	}

	c.Status(http.StatusNoContent)
}

// GET /api/public/categories
func (h *CategoriesHandler) PublicList(c *gin.Context) {
	ctx := c.Request.Context()
	items := make([]Category, 0)
	rows, err := h.DB.Query(ctx, `
		SELECT c.uuid_id, c.slug, c.name, c.description, c.parent_id, c.sort_order, c.image_id, m.path
		FROM categories c
		LEFT JOIN media m ON c.image_id = m.id
		ORDER BY c.parent_id NULLS FIRST, c.sort_order, c.uuid_id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var it Category
		if err := rows.Scan(&it.UUIDID, &it.Slug, &it.Name, &it.Description, &it.ParentID, &it.SortOrder, &it.ImageID, &it.ImagePath); err == nil {
			// Use ImageHelper to format image URL
			if h.ImageHelper != nil {
				it.ImageURL = h.ImageHelper.FormatImageURL(it.ImagePath)
			}
			items = append(items, it)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
