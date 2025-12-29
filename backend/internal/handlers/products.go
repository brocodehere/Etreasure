package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductsHandler struct {
	DB          *pgxpool.Pool
	R2Client    *storage.R2Client
	ImageHelper *storage.ImageURLHelper
}

// Helper function to convert image path to public URL
func (h *ProductsHandler) formatImageURL(imagePath *string) *string {
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

type Product struct {
	UUIDID      uuid.UUID  `json:"uuid_id"`
	Slug        string     `json:"slug"`
	Title       string     `json:"title"`
	Subtitle    *string    `json:"subtitle,omitempty"`
	Description *string    `json:"description,omitempty"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	Published   bool       `json:"published"`
	PublishAt   *time.Time `json:"publish_at,omitempty"`
	UnpublishAt *time.Time `json:"unpublish_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ProductVariant struct {
	ID                  int       `json:"id"`
	ProductID           uuid.UUID `json:"product_id"`
	SKU                 string    `json:"sku"`
	Title               string    `json:"title"`
	PriceCents          int       `json:"price_cents"`
	CompareAtPriceCents *int      `json:"compare_at_price_cents"`
	Currency            string    `json:"currency"`
	StockQuantity       int       `json:"stock_quantity"`
}

type ProductImage struct {
	MediaID   int `json:"media_id"`
	SortOrder int `json:"sort_order"`
}

type UpsertProductRequest struct {
	Slug        string           `json:"slug"`
	Title       string           `json:"title"`
	Subtitle    *string          `json:"subtitle"`
	Description *string          `json:"description"`
	CategoryID  *uuid.UUID       `json:"category_id"`
	Published   bool             `json:"published"`
	PublishAt   *time.Time       `json:"publish_at"`
	UnpublishAt *time.Time       `json:"unpublish_at"`
	Variants    []ProductVariant `json:"variants"`
	Images      []ProductImage   `json:"images"`
}

// GET /api/admin/products
func (h *ProductsHandler) List(c *gin.Context) {
	ctx := c.Request.Context()

	rows, err := h.DB.Query(ctx, `SELECT uuid_id, slug, title, subtitle, description, category_id, published, publish_at, unpublish_at, created_at, updated_at
		FROM products
		ORDER BY created_at DESC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	items := make([]Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.UUIDID, &p.Slug, &p.Title, &p.Subtitle, &p.Description, &p.CategoryID, &p.Published, &p.PublishAt, &p.UnpublishAt, &p.CreatedAt, &p.UpdatedAt); err == nil {
			items = append(items, p)
		}
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

// POST /api/admin/products
func (h *ProductsHandler) Create(c *gin.Context) {
	ctx := c.Request.Context()
	var req UpsertProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	var uuidID uuid.UUID
	err := h.DB.QueryRow(ctx, `INSERT INTO products (slug, title, subtitle, description, category_id, published, publish_at, unpublish_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		RETURNING uuid_id`, req.Slug, req.Title, req.Subtitle, req.Description, req.CategoryID, req.Published, req.PublishAt, req.UnpublishAt).Scan(&uuidID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "create failed"})
		return
	}

	// Variants
	for _, v := range req.Variants {
		_, err := h.DB.Exec(ctx, `INSERT INTO product_variants (product_id, sku, title, price_cents, compare_at_price_cents, currency, stock_quantity)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT (sku) DO NOTHING`, uuidID, v.SKU, v.Title, v.PriceCents, v.CompareAtPriceCents, v.Currency, v.StockQuantity)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "variant insert failed"})
			return
		}
	}
	// Images
	for _, img := range req.Images {
		_, err := h.DB.Exec(ctx, `INSERT INTO product_images (product_id, media_id, sort_order) VALUES ($1,$2,$3)`, uuidID, img.MediaID, img.SortOrder)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "image link failed"})
			return
		}
	}
	c.JSON(http.StatusCreated, gin.H{"uuid_id": uuidID})
}

// GET /api/admin/products/:id
func (h *ProductsHandler) Get(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	uuidID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var p Product
	err = h.DB.QueryRow(ctx, `SELECT uuid_id, slug, title, subtitle, description, category_id, published, publish_at, unpublish_at, created_at, updated_at FROM products WHERE uuid_id=$1`, uuidID).
		Scan(&p.UUIDID, &p.Slug, &p.Title, &p.Subtitle, &p.Description, &p.CategoryID, &p.Published, &p.PublishAt, &p.UnpublishAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	// variants
	vrows, err := h.DB.Query(ctx, `SELECT id, product_id, sku, title, price_cents, compare_at_price_cents, currency, stock_quantity FROM product_variants WHERE product_id=$1 ORDER BY id`, uuidID)
	defer vrows.Close()
	var variants []ProductVariant
	for vrows.Next() {
		var v ProductVariant
		var cmp *int
		if err := vrows.Scan(&v.ID, &v.ProductID, &v.SKU, &v.Title, &v.PriceCents, &cmp, &v.Currency, &v.StockQuantity); err == nil {
			v.CompareAtPriceCents = cmp
			variants = append(variants, v)
		}
	}
	// images
	imgRows, err := h.DB.Query(ctx, `SELECT media_id, sort_order FROM product_images WHERE product_id=$1 ORDER BY sort_order, media_id`, uuidID)
	defer imgRows.Close()
	var images []ProductImage
	for imgRows.Next() {
		var im ProductImage
		if err := imgRows.Scan(&im.MediaID, &im.SortOrder); err == nil {
			images = append(images, im)
		}
	}

	c.JSON(http.StatusOK, gin.H{"product": p, "variants": variants, "images": images})
}

// PUT /api/admin/products/:id (replace arrays)
func (h *ProductsHandler) Update(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	uuidID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req UpsertProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	_, err = h.DB.Exec(ctx, `UPDATE products SET slug=$1,title=$2,subtitle=$3,description=$4,category_id=$5,published=$6,publish_at=$7,unpublish_at=$8,updated_at=NOW() WHERE uuid_id=$9`,
		req.Slug, req.Title, req.Subtitle, req.Description, req.CategoryID, req.Published, req.PublishAt, req.UnpublishAt, uuidID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	_, _ = h.DB.Exec(ctx, `DELETE FROM product_variants WHERE product_id=$1`, uuidID)
	for _, v := range req.Variants {
		_, err := h.DB.Exec(ctx, `INSERT INTO product_variants (product_id, sku, title, price_cents, compare_at_price_cents, currency, stock_quantity)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`, uuidID, v.SKU, v.Title, v.PriceCents, v.CompareAtPriceCents, v.Currency, v.StockQuantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "variant upsert failed"})
			return
		}
	}
	_, _ = h.DB.Exec(ctx, `DELETE FROM product_images WHERE product_id=$1`, uuidID)
	for _, img := range req.Images {
		_, _ = h.DB.Exec(ctx, `INSERT INTO product_images (product_id, media_id, sort_order) VALUES ($1,$2,$3)`, uuidID, img.MediaID, img.SortOrder)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// DELETE /api/admin/products/:id
func (h *ProductsHandler) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	idStr := c.Param("id")
	uuidID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	// First, get all media IDs for this product to delete from R2
	var mediaIDs []int
	rows, err := h.DB.Query(ctx, `SELECT media_id FROM product_images WHERE product_id=$1`, uuidID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var mediaID int
			if rows.Scan(&mediaID) == nil {
				mediaIDs = append(mediaIDs, mediaID)
			}
		}
	}

	// Delete product
	_, err = h.DB.Exec(ctx, `DELETE FROM products WHERE uuid_id=$1`, uuidID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		return
	}

	// Delete media from R2 if R2Client is available
	if h.R2Client != nil {
		for _, mediaID := range mediaIDs {
			var path string
			err := h.DB.QueryRow(ctx, `SELECT path FROM media WHERE id=$1`, mediaID).Scan(&path)
			if err == nil && (strings.HasPrefix(path, "product/") || strings.HasPrefix(path, "banner/") || strings.HasPrefix(path, "category/")) {
				if err := h.R2Client.DeleteObject(ctx, path); err != nil {
					log.Printf("Warning: Failed to delete from R2: %v", err)
				}
			}
		}
	}

	c.Status(http.StatusNoContent)
}

// POST /api/admin/products/import (stub)
func (h *ProductsHandler) Import(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
}

// PublicList exposes a simplified product list for the storefront with optional category filtering.
// GET /api/products?category_id=<uuid> or GET /api/products?category=<slug>
func (h *ProductsHandler) PublicList(c *gin.Context) {
	ctx := c.Request.Context()

	// Pagination
	limit := 12
	if l := strings.TrimSpace(c.Query("limit")); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	page := 1
	if p := strings.TrimSpace(c.Query("page")); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	offset := (page - 1) * limit

	search := strings.TrimSpace(c.Query("search"))
	sort := strings.TrimSpace(c.Query("sort"))
	minPriceStr := strings.TrimSpace(c.Query("min_price"))
	maxPriceStr := strings.TrimSpace(c.Query("max_price"))
	inStock := strings.TrimSpace(c.Query("in_stock")) == "1"

	var minPriceCents *int
	if minPriceStr != "" {
		if parsed, err := strconv.Atoi(minPriceStr); err == nil && parsed >= 0 {
			v := parsed * 100
			minPriceCents = &v
		}
	}
	var maxPriceCents *int
	if maxPriceStr != "" {
		if parsed, err := strconv.Atoi(maxPriceStr); err == nil && parsed >= 0 {
			v := parsed * 100
			maxPriceCents = &v
		}
	}

	// Check for category_id (UUID) first, then category (slug)
	categoryIDParam := strings.TrimSpace(c.Query("category_id"))
	categorySlugParam := strings.TrimSpace(c.Query("category"))

	var categoryFilter *uuid.UUID
	var categorySlugFilter string

	if categoryIDParam != "" {
		parsed, err := uuid.Parse(categoryIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		categoryFilter = &parsed
	} else if categorySlugParam != "" {
		categorySlugFilter = categorySlugParam
	}

	// Pick a representative variant price (lowest) and first image for listing cards.
	var baseQuery string
	args := []interface{}{}
	where := ""
	argN := 1

	if categorySlugFilter != "" {
		// When filtering by slug, join with categories table
		baseQuery = `
			SELECT p.uuid_id,
			       p.slug,
			       p.title,
			       p.description,
			       p.category_id,
			       pv.price_cents,
			       pv.currency,
			       m.path AS image_path,
			       p.created_at,
			       COALESCE((SELECT SUM(stock_quantity) FROM product_variants WHERE product_id = p.uuid_id), 0) as total_stock
			FROM products p
			LEFT JOIN LATERAL (
				SELECT price_cents, currency
				FROM product_variants
				WHERE product_id = p.uuid_id
				ORDER BY price_cents ASC
				LIMIT 1
			) pv ON true
			LEFT JOIN LATERAL (
				SELECT m.path
				FROM product_images pi
				JOIN media m ON m.id = pi.media_id
				WHERE pi.product_id = p.uuid_id
				ORDER BY pi.sort_order, pi.media_id
				LIMIT 1
			) m ON true
			INNER JOIN categories c ON c.uuid_id = p.category_id
			WHERE p.published = TRUE
			  AND (p.publish_at IS NULL OR p.publish_at <= NOW())
			  AND c.slug = $1
		`
		args = append(args, categorySlugFilter)
		argN = 2
	} else {
		// Regular query without category filtering or with category_id filtering
		baseQuery = `
			SELECT p.uuid_id,
			       p.slug,
			       p.title,
			       p.description,
			       p.category_id,
			       pv.price_cents,
			       pv.currency,
			       m.path AS image_path,
			       p.created_at,
			       COALESCE((SELECT SUM(stock_quantity) FROM product_variants WHERE product_id = p.uuid_id), 0) as total_stock
			FROM products p
			LEFT JOIN LATERAL (
				SELECT price_cents, currency
				FROM product_variants
				WHERE product_id = p.uuid_id
				ORDER BY price_cents ASC
				LIMIT 1
			) pv ON true
			LEFT JOIN LATERAL (
				SELECT m.path
				FROM product_images pi
				JOIN media m ON m.id = pi.media_id
				WHERE pi.product_id = p.uuid_id
				ORDER BY pi.sort_order, pi.media_id
				LIMIT 1
			) m ON true
			WHERE p.published = TRUE
			  AND (p.publish_at IS NULL OR p.publish_at <= NOW())
		`

		if categoryFilter != nil {
			baseQuery += " AND p.category_id = $1"
			args = append(args, *categoryFilter)
			argN = 2
		}
	}

	if search != "" {
		where += " AND lower(p.title) LIKE $" + strconv.Itoa(argN)
		args = append(args, "%"+strings.ToLower(search)+"%")
		argN++
	}
	if minPriceCents != nil {
		where += " AND COALESCE(pv.price_cents, 0) >= $" + strconv.Itoa(argN)
		args = append(args, *minPriceCents)
		argN++
	}
	if maxPriceCents != nil {
		where += " AND COALESCE(pv.price_cents, 0) <= $" + strconv.Itoa(argN)
		args = append(args, *maxPriceCents)
		argN++
	}
	if inStock {
		where += " AND EXISTS (SELECT 1 FROM product_variants v2 WHERE v2.product_id = p.uuid_id AND v2.stock_quantity > 0)"
	}

	orderBy := "p.created_at DESC"
	switch sort {
	case "price_asc":
		orderBy = "pv.price_cents ASC NULLS LAST, p.created_at DESC"
	case "price_desc":
		orderBy = "pv.price_cents DESC NULLS LAST, p.created_at DESC"
	case "name_asc":
		orderBy = "p.title ASC"
	case "name_desc":
		orderBy = "p.title DESC"
	default:
		orderBy = "p.created_at DESC"
	}

	query := baseQuery + where + " ORDER BY " + orderBy + " LIMIT $" + strconv.Itoa(argN) + " OFFSET $" + strconv.Itoa(argN+1)
	args = append(args, limit, offset)

	countQuery := "SELECT COUNT(1) FROM (" + baseQuery + where + ") as q"
	var total int
	if err := h.DB.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total); err != nil {
		total = 0
	}

	rows, err := h.DB.Query(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	type PublicProduct struct {
		ID            uuid.UUID  `json:"id"`
		Slug          string     `json:"slug"`
		Title         string     `json:"title"`
		Description   *string    `json:"description,omitempty"`
		CategoryID    *uuid.UUID `json:"category_id,omitempty"`
		PriceCents    *int       `json:"price_cents,omitempty"`
		Currency      *string    `json:"currency,omitempty"`
		ImageKey      *string    `json:"image_key,omitempty"`
		ImageURL      *string    `json:"image_url,omitempty"`
		StockQuantity int        `json:"stock_quantity"`
		CreatedAt     time.Time  `json:"created_at"`
	}

	items := make([]PublicProduct, 0)
	for rows.Next() {
		var p PublicProduct
		var imagePath *string
		if err := rows.Scan(&p.ID, &p.Slug, &p.Title, &p.Description, &p.CategoryID, &p.PriceCents, &p.Currency, &imagePath, &p.CreatedAt, &p.StockQuantity); err == nil {
			if h.ImageHelper != nil {
				p.ImageKey, p.ImageURL = h.ImageHelper.GetImageKeyAndURL(imagePath)

				// If no image URL is available, use fallback
				if p.ImageURL == nil || *p.ImageURL == "" {
					fallbackURL := h.ImageHelper.GetFallbackImageURL("product")
					p.ImageURL = &fallbackURL
				}
			}
			items = append(items, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total, "page": page, "limit": limit})
}

// PublicGet returns a single published product by slug for the storefront
// GET /api/products/:id (id is treated as slug here)
func (h *ProductsHandler) PublicGet(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("id")
	var (
		p          Product
		priceCents int
		currency   string
	)
	err := h.DB.QueryRow(ctx, `
		SELECT p.uuid_id, p.slug, p.title, p.subtitle, p.description, p.category_id, p.published, p.publish_at, p.unpublish_at, p.created_at, p.updated_at,
			COALESCE(MIN(v.price_cents), 0) AS price_cents,
			COALESCE(MIN(v.currency), 'INR') AS currency
		FROM products p
		LEFT JOIN product_variants v ON v.product_id = p.uuid_id
		WHERE p.slug = $1 AND p.published = TRUE AND (p.publish_at IS NULL OR p.publish_at <= NOW())
		GROUP BY p.uuid_id
	`, slug).
		Scan(&p.UUIDID, &p.Slug, &p.Title, &p.Subtitle, &p.Description, &p.CategoryID, &p.Published, &p.PublishAt, &p.UnpublishAt, &p.CreatedAt, &p.UpdatedAt, &priceCents, &currency)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	// Basic availability: in stock if any variant has stock_quantity > 0
	var availability string = "OutOfStock"
	var stockQty int
	if err := h.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(stock_quantity), 0) FROM product_variants WHERE product_id = $1
	`, p.UUIDID).Scan(&stockQty); err == nil && stockQty > 0 {
		availability = "InStock"
	}

	// Fetch all product images
	rows, err := h.DB.Query(ctx, `
		SELECT m.path, pi.sort_order
		FROM product_images pi
		JOIN media m ON m.id = pi.media_id
		WHERE pi.product_id = $1
		ORDER BY pi.sort_order, pi.media_id
	`, p.UUIDID)
	if err != nil && err != pgx.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load images"})
		return
	}
	defer rows.Close()

	var images []gin.H
	imageCount := 0
	for rows.Next() {
		var path string
		var sortOrder int
		if err := rows.Scan(&path, &sortOrder); err != nil {
			continue
		}
		imageCount++
		images = append(images, gin.H{
			"url":        h.formatImageURL(&path),
			"sort_order": sortOrder,
		})
	}

	// Get hero image (first image) for backward compatibility
	var heroURL *string
	if len(images) > 0 {
		heroPath := images[0]["url"].(*string)
		heroURL = heroPath
	} else {
		heroURL = nil
	}

	// Projection support via fields query param
	fieldsParam := c.Query("fields")
	if fieldsParam == "" {
		response := gin.H{
			"id":           p.UUIDID,
			"slug":         p.Slug,
			"title":        p.Title,
			"description":  p.Description,
			"price_cents":  priceCents,
			"currency":     currency,
			"availability": availability,
			"hero_image": gin.H{
				"url": heroURL,
			},
			"images": images,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	requested := map[string]bool{}
	for _, f := range strings.Split(fieldsParam, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			requested[f] = true
		}
	}

	resp := gin.H{}
	if requested["id"] {
		resp["id"] = p.UUIDID
	}
	if requested["slug"] {
		resp["slug"] = p.Slug
	}
	if requested["title"] {
		resp["title"] = p.Title
	}
	if requested["description"] {
		resp["description"] = p.Description
	}
	if requested["price_cents"] {
		resp["price_cents"] = priceCents
	}
	if requested["currency"] {
		resp["currency"] = currency
	}
	if requested["availability"] {
		resp["availability"] = availability
	}
	if requested["hero_image"] {
		var formattedHeroURL *string
		if len(images) > 0 {
			heroPath := images[0]["url"].(*string)
			formattedHeroURL = heroPath
		} else {
			formattedHeroURL = nil
		}
		resp["hero_image"] = gin.H{"url": formattedHeroURL}
	}
	if requested["images"] {
		resp["images"] = images
	} else {
	}

	c.JSON(http.StatusOK, resp)
}

// Related returns related products for storefront
// GET /api/products/:id/related
func (h *ProductsHandler) Related(c *gin.Context) {
	ctx := c.Request.Context()
	slug := c.Param("id")

	// Find product uuid_id and category_id
	var (
		productUUID uuid.UUID
		categoryID  *uuid.UUID
	)
	err := h.DB.QueryRow(ctx, `
		SELECT uuid_id, category_id
		FROM products
		WHERE slug = $1
	`, slug).Scan(&productUUID, &categoryID)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	// Prefer same-category products, fallback to any other products
	var rows pgx.Rows
	if categoryID != nil {
		rows, err = h.DB.Query(ctx, `
			SELECT p.uuid_id, p.slug, p.title,
				COALESCE(MIN(v.price_cents), 0) AS price_cents,
				COALESCE(MIN(v.currency), 'INR') AS currency
			FROM products p
			LEFT JOIN product_variants v ON v.product_id = p.uuid_id
			WHERE p.category_id = $1 AND p.uuid_id <> $2 AND p.published = TRUE
			GROUP BY p.uuid_id
			ORDER BY p.created_at DESC
			LIMIT 8`, *categoryID, productUUID)
	} else {
		rows, err = h.DB.Query(ctx, `
			SELECT p.uuid_id, p.slug, p.title,
				COALESCE(MIN(v.price_cents), 0) AS price_cents,
				COALESCE(MIN(v.currency), 'INR') AS currency
			FROM products p
			LEFT JOIN product_variants v ON v.product_id = p.uuid_id
			WHERE p.uuid_id <> $1 AND p.published = TRUE
			GROUP BY p.uuid_id
			ORDER BY p.created_at DESC
			LIMIT 8`, productUUID)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	type relatedProduct struct {
		UUIDID     uuid.UUID `json:"uuid_id"`
		Slug       string    `json:"slug"`
		Title      string    `json:"title"`
		PriceCents int       `json:"price_cents"`
		Currency   string    `json:"currency"`
	}

	var items []relatedProduct
	for rows.Next() {
		var rp relatedProduct
		if err := rows.Scan(&rp.UUIDID, &rp.Slug, &rp.Title, &rp.PriceCents, &rp.Currency); err == nil {
			items = append(items, rp)
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// Search performs a simple text search over product title and description for the storefront
// POST /api/products/search { "query": "..." }
func (h *ProductsHandler) Search(c *gin.Context) {
	ctx := c.Request.Context()
	var body struct {
		Query string `json:"query"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	q := "%" + body.Query + "%"
	rows, err := h.DB.Query(ctx, `
		SELECT uuid_id, slug, title, subtitle, description, category_id, published, publish_at, unpublish_at, created_at, updated_at
		FROM products
		WHERE published = TRUE
		  AND (publish_at IS NULL OR publish_at <= NOW())
		  AND (title ILIKE $1 OR description ILIKE $1)
		ORDER BY created_at DESC
		LIMIT 50`, q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	items := make([]Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.UUIDID, &p.Slug, &p.Title, &p.Subtitle, &p.Description, &p.CategoryID, &p.Published, &p.PublishAt, &p.UnpublishAt, &p.CreatedAt, &p.UpdatedAt); err == nil {
			items = append(items, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

// GetOutOfStockProducts returns all published products that are out of stock
// GET /api/admin/products/out-of-stock
func (h *ProductsHandler) GetOutOfStockProducts(c *gin.Context) {
	ctx := c.Request.Context()

	query := `
		SELECT p.uuid_id,
			   p.slug,
			   p.title,
			   p.subtitle,
			   p.description,
			   p.category_id,
			   p.published,
			   p.publish_at,
			   p.unpublish_at,
			   p.created_at,
			   p.updated_at,
			   COALESCE((SELECT SUM(stock_quantity) FROM product_variants WHERE product_id = p.uuid_id), 0) as total_stock,
			   (SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_path
		FROM products p
		WHERE p.published = true 
		  AND NOT EXISTS (
			SELECT 1 FROM product_variants pv 
			WHERE pv.product_id = p.uuid_id AND pv.stock_quantity > 0
		  )
		ORDER BY p.updated_at DESC
	`

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}
	defer rows.Close()

	type OutOfStockProduct struct {
		UUIDID      uuid.UUID        `json:"id"`
		Slug        string           `json:"slug"`
		Title       string           `json:"title"`
		Subtitle    *string          `json:"subtitle,omitempty"`
		Description *string          `json:"description,omitempty"`
		CategoryID  *uuid.UUID       `json:"category_id,omitempty"`
		Published   bool             `json:"published"`
		PublishAt   *time.Time       `json:"publish_at,omitempty"`
		UnpublishAt *time.Time       `json:"unpublish_at,omitempty"`
		CreatedAt   time.Time        `json:"created_at"`
		UpdatedAt   time.Time        `json:"updated_at"`
		TotalStock  int              `json:"total_stock"`
		ImageKey    *string          `json:"image_key,omitempty"`
		ImageURL    *string          `json:"image_url,omitempty"`
		Variants    []ProductVariant `json:"variants"`
	}

	var products []OutOfStockProduct
	for rows.Next() {
		var p OutOfStockProduct
		var imagePath *string
		if err := rows.Scan(&p.UUIDID, &p.Slug, &p.Title, &p.Subtitle, &p.Description, &p.CategoryID, &p.Published, &p.PublishAt, &p.UnpublishAt, &p.CreatedAt, &p.UpdatedAt, &p.TotalStock, &imagePath); err == nil {

			// Set image fields
			if imagePath != nil {
				p.ImageKey, p.ImageURL = h.ImageHelper.GetImageKeyAndURL(imagePath)

				// If no image URL is available, use fallback
				if p.ImageURL == nil || *p.ImageURL == "" {
					fallbackURL := h.ImageHelper.GetFallbackImageURL("product")
					p.ImageURL = &fallbackURL
				}
			}

			// Fetch variants for this product
			variantRows, err := h.DB.Query(ctx, `
				SELECT id, product_id, sku, title, price_cents, compare_at_price_cents, currency, stock_quantity
				FROM product_variants
				WHERE product_id = $1
				ORDER BY id
			`, p.UUIDID)
			if err == nil {
				defer variantRows.Close()
				for variantRows.Next() {
					var variant ProductVariant
					if err := variantRows.Scan(&variant.ID, &variant.ProductID, &variant.SKU, &variant.Title, &variant.PriceCents, &variant.CompareAtPriceCents, &variant.Currency, &variant.StockQuantity); err == nil {
						p.Variants = append(p.Variants, variant)
					}
				}
			}

			products = append(products, p)
		}
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

// UpdateVariantStock updates the stock quantity for a specific product variant
// PATCH /api/admin/products/:productId/variants/:variantId/stock
func (h *ProductsHandler) UpdateVariantStock(c *gin.Context) {
	ctx := c.Request.Context()
	productID := c.Param("productId")
	variantID, err := strconv.Atoi(c.Param("variantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid variant ID"})
		return
	}

	var req struct {
		StockQuantity int `json:"stock_quantity" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Update the stock quantity
	_, err = h.DB.Exec(ctx, `
		UPDATE product_variants 
		SET stock_quantity = $1, updated_at = NOW()
		WHERE id = $2 AND product_id = $3
	`, req.StockQuantity, variantID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update stock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "stock updated successfully", "stock_quantity": req.StockQuantity})
}
