package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WishlistHandler struct {
	DB          *pgxpool.Pool
	ImageHelper *storage.ImageURLHelper
}

type ToggleWishlistRequest struct {
	ProductID string `json:"product_id" binding:"required"`
}

type WishlistItem struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	ImageURL  string  `json:"image_url"`
	AddedAt   string  `json:"added_at"`
}

type WishlistResponse struct {
	Items []WishlistItem `json:"items"`
	Count int            `json:"count"`
}

// ToggleWishlist adds or removes a product from the wishlist
func (h *WishlistHandler) ToggleWishlist(c *gin.Context) {
	var req ToggleWishlistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Check if product exists
	var title string
	var priceCents int
	var currency string
	var imageURL string
	err := h.DB.QueryRow(ctx, `
		SELECT p.title, pv.price_cents, pv.currency, 
		       (SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_url
		FROM products p
		LEFT JOIN product_variants pv ON p.uuid_id = pv.product_id
		WHERE p.uuid_id = $1::uuid
		ORDER BY pv.id
		LIMIT 1
	`, req.ProductID).Scan(&title, &priceCents, &currency, &imageURL)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Get authenticated user ID from middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Check if item is already in wishlist
	var existingID int
	err = h.DB.QueryRow(ctx, `
		SELECT id FROM wishlist 
		WHERE user_id = $1 AND product_id = $2::uuid
	`, userID, req.ProductID).Scan(&existingID)

	if err == nil {
		// Remove from wishlist
		_, err = h.DB.Exec(ctx, `
			DELETE FROM wishlist 
			WHERE user_id = $1 AND product_id = $2::uuid
		`, userID, req.ProductID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove from wishlist"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":     "Product removed from wishlist successfully",
			"product_id":  req.ProductID,
			"in_wishlist": false,
		})
	} else {
		// Add to wishlist
		_, err = h.DB.Exec(ctx, `
			INSERT INTO wishlist (user_id, product_id, created_at)
			VALUES ($1, $2::uuid, NOW())
		`, userID, req.ProductID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to wishlist"})
			return
		}

		// Convert cents to price
		price := float64(priceCents) / 100.0

		c.JSON(http.StatusOK, gin.H{
			"message":     "Product added to wishlist successfully",
			"product_id":  req.ProductID,
			"in_wishlist": true,
			"price":       price,
			"currency":    currency,
		})
	}
}

// GetWishlist retrieves the authenticated user's wishlist
func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	ctx := context.Background()

	// Query with proper JOIN to get wishlist items with product details
	query := `
		SELECT 
			w.id,
			w.product_id,
			p.title,
			COALESCE(pv.price_cents, 0) as price_cents,
			COALESCE(pv.currency, 'USD') as currency,
			w.created_at,
			(SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_url
		FROM wishlist w
		JOIN products p ON w.product_id = p.uuid_id
		LEFT JOIN product_variants pv ON p.uuid_id = pv.product_id
		WHERE w.user_id = $1
		ORDER BY w.created_at DESC
	`

	rows, err := h.DB.Query(ctx, query, userID)
	if err != nil {
		// Check if it's a table doesn't exist error
		if strings.Contains(err.Error(), "does not exist") {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Wishlist table not found", "details": "Database schema issue"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch wishlist", "details": err.Error()})
		}
		return
	}
	defer rows.Close()

	// Process rows
	var items []WishlistItem
	for rows.Next() {
		var item WishlistItem
		var id int
		var priceCents int
		var currency string
		var createdAt time.Time
		var imagePath *string

		err := rows.Scan(&id, &item.ProductID, &item.Title, &priceCents, &currency, &createdAt, &imagePath)
		if err != nil {
			continue
		}

		item.ID = fmt.Sprintf("%d", id)
		item.Price = float64(priceCents) / 100.0
		item.AddedAt = createdAt.Format("2006-01-02T15:04:05Z")

		// Construct proper image URL using ImageHelper
		if h.ImageHelper != nil {
			_, imageURL := h.ImageHelper.GetImageKeyAndURL(imagePath)

			// If no image URL is available, use fallback
			if imageURL == nil || *imageURL == "" {
				fallbackURL := h.ImageHelper.GetFallbackImageURL("product")
				item.ImageURL = fallbackURL
			} else {
				item.ImageURL = *imageURL
			}
		} else {
			// Fallback if ImageHelper is not available
			if imagePath != nil {
				item.ImageURL = *imagePath
			} else {
				item.ImageURL = "/product-placeholder.webp"
			}
		}

		items = append(items, item)
	}

	// Ensure items is never null
	if items == nil {
		items = []WishlistItem{}
	}

	wishlist := WishlistResponse{
		Items: items,
		Count: len(items),
	}

	c.JSON(http.StatusOK, wishlist)
}

// RemoveFromWishlist removes an item from the wishlist
func (h *WishlistHandler) RemoveFromWishlist(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
		return
	}

	ctx := context.Background()

	// Get authenticated user ID from middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Delete item from wishlist
	_, err := h.DB.Exec(ctx, `
		DELETE FROM wishlist 
		WHERE user_id = $1 AND product_id = $2::uuid
	`, userID, productID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Product removed from wishlist successfully",
		"product_id": productID,
	})
}
