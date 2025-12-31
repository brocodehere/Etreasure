package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CartHandler struct {
	DB          *pgxpool.Pool
	ImageHelper *storage.ImageURLHelper
}

type AddToCartRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price,omitempty"` // Allow price override for discounts
}

type CartItem struct {
	ID              string  `json:"id"`
	ProductID       string  `json:"product_id"`
	Title           string  `json:"title"`
	Price           float64 `json:"price"`
	Quantity        int     `json:"quantity"`
	ImageURL        string  `json:"image_url"`
	SKU             string  `json:"sku"`
	DiscountedPrice *string `json:"discounted_price,omitempty"`
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
	var currentStock int

	err := h.DB.QueryRow(ctx, `
		SELECT p.title, pv.price_cents, pv.currency, 
		       (SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_url,
		       pv.id as variant_id,
		       COALESCE(pv.stock_quantity, 0) as stock_quantity
		FROM products p
		LEFT JOIN product_variants pv ON p.uuid_id = pv.product_id
		WHERE p.uuid_id = $1::uuid
		ORDER BY pv.id
		LIMIT 1
	`, req.ProductID).Scan(&title, &priceCents, &currency, &imageURL, &variantID, &currentStock)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check stock availability
	if currentStock < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":         "insufficient stock",
			"current_stock": currentStock,
			"requested":     req.Quantity,
			"variant_id":    variantID,
		})
		return
	}

	// Get session ID from cookie or create new one
	sessionID, err := c.Cookie("session_id")

	if err != nil || sessionID == "" {
		// Generate new session ID using UUID
		sessionID = fmt.Sprintf("session_%s", uuid.New().String())
		// Set secure flag based on whether request is HTTPS
		// Check both TLS and X-Forwarded-Proto header (for proxy setups)
		isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

		// For production, set cookie domain to work across ethnictreasures.co.in
		cookieDomain := ""
		// Only set domain for production/remote environments, not localhost
		requestHost := c.Request.Host
		headerHost := c.GetHeader("Host")
		origin := c.GetHeader("Origin")
		referer := c.GetHeader("Referer")

		// Check if this is a production request from ethnictreasures.co.in
		isProduction := (origin == "https://ethnictreasures.co.in") ||
			(referer != "" && strings.Contains(referer, "ethnictreasures.co.in")) ||
			(headerHost != "" && strings.Contains(headerHost, "ethnictreasures.co.in")) ||
			(requestHost != "" && strings.Contains(requestHost, "ethnictreasures.co.in"))

		// Check if this is localhost
		isLocalhost := requestHost == "localhost:8080" ||
			headerHost == "localhost:8080" ||
			requestHost == "127.0.0.1:8080" ||
			headerHost == "127.0.0.1:8080" ||
			requestHost == "localhost:4321" ||
			headerHost == "localhost:4321" ||
			origin == "http://localhost:4321" ||
			origin == "http://localhost:3000"

		if isProduction && !isLocalhost {
			// For cross-origin requests, don't set domain to let browser handle cookies naturally
			cookieDomain = os.Getenv("COOKIE_DOMAIN")
			if cookieDomain == "" {
				// Don't set domain for cross-origin - let browser handle it
				cookieDomain = ""
			}
		}

		// Industry standard: Set single cookie with proper attributes for cross-domain
		// For cross-domain scenarios, we need Domain=ethnictreasures.co.in with SameSite=None; Secure
		// For localhost scenarios, we need no domain with SameSite=Lax

		var cookieString string
		if cookieDomain != "" && isSecure {
			// Production: Cross-domain cookie with explicit domain
			cookieString = fmt.Sprintf("session_id=%s; Path=/; Domain=%s; Max-Age=%d; HttpOnly=false; SameSite=None; Secure",
				sessionID, cookieDomain, 86400*30)
		} else if isSecure {
			// Cross-origin or localhost: Secure cookie without domain
			cookieString = fmt.Sprintf("session_id=%s; Path=/; Max-Age=%d; HttpOnly=false; SameSite=None; Secure",
				sessionID, 86400*30)
		} else {
			// Localhost: Simple cookie
			cookieString = fmt.Sprintf("session_id=%s; Path=/; Max-Age=%d; HttpOnly=false; SameSite=Lax",
				sessionID, 86400*30)
		}

		c.Header("Set-Cookie", cookieString)
	} else {
		// Use existing session
	}

	// Use provided price or get from database
	var finalPrice float64
	if req.Price > 0 {
		// Use the provided price (discounted price)
		finalPrice = req.Price
	} else {
		// Use regular price from database
		finalPrice = float64(priceCents) / 100.0
	}

	// Check if item already exists in cart
	var existingQuantity int
	err = h.DB.QueryRow(ctx, `
		SELECT quantity FROM cart 
		WHERE session_id = $1 AND product_id = $2::uuid AND variant_id = $3
	`, sessionID, req.ProductID, variantID).Scan(&existingQuantity)

	newTotalQuantity := existingQuantity + req.Quantity

	if err == nil {
		// Check if new total quantity exceeds available stock
		if newTotalQuantity > currentStock {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":           "insufficient stock",
				"current_stock":   currentStock,
				"existing_cart":   existingQuantity,
				"requested":       req.Quantity,
				"total_requested": newTotalQuantity,
				"variant_id":      variantID,
			})
			return
		}
		// Update existing item with discounted price
		_, err = h.DB.Exec(ctx, `
			UPDATE cart 
			SET quantity = quantity + $1, discounted_price = $2, updated_at = NOW()
			WHERE session_id = $3 AND product_id = $4::uuid AND variant_id = $5
		`, req.Quantity, fmt.Sprintf("%.2f", finalPrice), sessionID, req.ProductID, variantID)
	} else {
		// Insert new item with discounted price
		_, err = h.DB.Exec(ctx, `
			INSERT INTO cart (session_id, product_id, variant_id, quantity, discounted_price, updated_at)
			VALUES ($1, $2::uuid, $3, $4, $5, NOW())
		`, sessionID, req.ProductID, variantID, req.Quantity, fmt.Sprintf("%.2f", finalPrice))
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Product added to cart successfully",
		"product_id": req.ProductID,
		"quantity":   req.Quantity + existingQuantity,
		"price":      finalPrice,
		"currency":   currency,
	})
}

// GetCart retrieves the current user's cart
func (h *CartHandler) GetCart(c *gin.Context) {
	ctx := context.Background()

	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")

	if err != nil || sessionID == "" {
		// Return empty cart for new users
		c.JSON(http.StatusOK, CartResponse{
			Items: []CartItem{},
			Total: 0.0,
			Count: 0,
		})
		return
	}

	query := `
			SELECT 
			c.id,
			c.product_id::text,
			p.title,
			pv.sku,
			CASE 
				WHEN c.discounted_price IS NOT NULL THEN c.discounted_price::numeric * 100
				ELSE pv.price_cents
			END as price_cents,
			pv.currency,
			c.quantity,
			c.discounted_price,
			(SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_url
		FROM cart c
		JOIN products p ON c.product_id = p.uuid_id
		JOIN product_variants pv ON c.variant_id = pv.id
		WHERE c.session_id = $1
	`

	rows, err := h.DB.Query(ctx, query, sessionID)
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
	total := 0.0
	count := 0
	for rows.Next() {
		var item CartItem
		var id any
		var priceCents int
		var currency string
		var quantity int
		var imagePath *string
		var productID string
		var sku *string
		var discountedPrice *string

		err := rows.Scan(&id, &productID, &item.Title, &sku, &priceCents, &currency, &quantity, &discountedPrice, &imagePath)
		if err != nil {
			continue
		}

		item.ID = fmt.Sprint(id)
		item.ProductID = productID
		item.Price = float64(priceCents) / 100.0
		item.Quantity = quantity

		// Set SKU if available
		if sku != nil {
			item.SKU = *sku
		}

		// Set discounted price if available
		if discountedPrice != nil {
			item.DiscountedPrice = discountedPrice
		}

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

	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
		return
	}

	// Delete item from cart
	_, err = h.DB.Exec(ctx, `
		DELETE FROM cart 
		WHERE id = $1 AND session_id = $2
	`, itemID, sessionID)

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

	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
		return
	}

	// Delete all items from cart
	_, err = h.DB.Exec(ctx, `
		DELETE FROM cart 
		WHERE session_id = $1
	`, sessionID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
	})
}
