package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartHandler struct {
	DB          *pgxpool.Pool
	ImageHelper *storage.ImageURLHelper
}

type AddToCartRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity"`
}

type CartItem struct {
	ID        string  `json:"id"`
	ProductID string  `json:"product_id"`
	Title     string  `json:"title"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	ImageURL  string  `json:"image_url"`
}

type CartResponse struct {
	Items []CartItem `json:"items"`
	Total float64    `json:"total"`
	Count int        `json:"count"`
}

// AddToCart adds a product to the cart
func (h *CartHandler) AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	// Check if product exists and get variant info
	var title string
	var priceCents int
	var currency string
	var imageURL string
	var variantID int

	err := h.DB.QueryRow(ctx, `
		SELECT p.title, pv.price_cents, pv.currency, 
		       (SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_url,
		       pv.id as variant_id
		FROM products p
		LEFT JOIN product_variants pv ON p.uuid_id = pv.product_id
		WHERE p.uuid_id = $1::uuid
		ORDER BY pv.id
		LIMIT 1
	`, req.ProductID).Scan(&title, &priceCents, &currency, &imageURL, &variantID)

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

	// Check if item already exists in cart
	var existingQuantity int
	err = h.DB.QueryRow(ctx, `
		SELECT quantity FROM cart 
		WHERE user_id = $1 AND product_id = $2::uuid AND variant_id = $3
	`, userID, req.ProductID, variantID).Scan(&existingQuantity)

	if err == nil {
		// Update existing item
		_, err = h.DB.Exec(ctx, `
			UPDATE cart 
			SET quantity = quantity + $1, updated_at = NOW()
			WHERE user_id = $2 AND product_id = $3::uuid AND variant_id = $4
		`, req.Quantity, userID, req.ProductID, variantID)
	} else {
		// Insert new item
		_, err = h.DB.Exec(ctx, `
			INSERT INTO cart (user_id, product_id, variant_id, quantity, updated_at)
			VALUES ($1, $2::uuid, $3, $4, NOW())
		`, userID, req.ProductID, variantID, req.Quantity)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to cart"})
		return
	}

	// Convert cents to price
	price := float64(priceCents) / 100.0

	c.JSON(http.StatusOK, gin.H{
		"message":    "Product added to cart successfully",
		"product_id": req.ProductID,
		"quantity":   req.Quantity + existingQuantity,
		"price":      price,
		"currency":   currency,
	})
}

// GetCart retrieves the current user's cart
func (h *CartHandler) GetCart(c *gin.Context) {
	ctx := context.Background()

	// Get authenticated user ID from middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	query := `
		SELECT 
			c.id,
			c.product_id,
			p.title,
			pv.price_cents,
			pv.currency,
			c.quantity,
			(SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_url
		FROM cart c
		JOIN products p ON c.product_id = p.uuid_id
		JOIN product_variants pv ON c.variant_id = pv.id
		WHERE c.user_id = $1
	`

	rows, err := h.DB.Query(ctx, query, userID)
	if err != nil {
		// Check if it's a table doesn't exist error
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "relation") {
			c.JSON(http.StatusOK, CartResponse{
				Items: []CartItem{},
				Total: 0.0,
				Count: 0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart", "details": err.Error()})
		return
	}
	defer rows.Close()

	items := make([]CartItem, 0)
	var total float64
	var count int

	for rows.Next() {
		var item CartItem
		var id int
		var priceCents int
		var currency string
		var imagePath *string

		err := rows.Scan(&id, &item.ProductID, &item.Title, &priceCents, &currency, &item.Quantity, &imagePath)
		if err != nil {
			continue
		}

		item.ID = fmt.Sprintf("%d", id)
		item.Price = float64(priceCents) / 100.0

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
		total += item.Price * float64(item.Quantity)
		count += item.Quantity
	}

	cart := CartResponse{
		Items: items,
		Total: total,
		Count: count,
	}

	c.JSON(http.StatusOK, cart)
}

// RemoveFromCart removes an item from the cart
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	itemID := c.Param("id")
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Item ID is required"})
		return
	}

	ctx := context.Background()

	// Get authenticated user ID from middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Delete item from cart
	_, err := h.DB.Exec(ctx, `
		DELETE FROM cart 
		WHERE id = $1 AND user_id = $2
	`, itemID, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item removed from cart successfully",
		"item_id": itemID,
	})
}

// ClearCart clears all items from the cart
func (h *CartHandler) ClearCart(c *gin.Context) {
	ctx := context.Background()

	// Get authenticated user ID from middleware
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Delete all items from cart
	_, err := h.DB.Exec(ctx, `
		DELETE FROM cart 
		WHERE user_id = $1
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
	})
}
