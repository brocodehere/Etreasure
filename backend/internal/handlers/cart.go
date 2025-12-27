package handlers

import (
	"context"
	"fmt"
	"log"
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

	// Get session ID from cookie or create new one
	sessionID, err := c.Cookie("session_id")
	log.Printf("Cart: Reading session_id cookie: %s, error: %v", sessionID, err)
	log.Printf("Cart: Request headers: %+v", c.Request.Header)
	log.Printf("Cart: Request origin: %s", c.GetHeader("Origin"))
	log.Printf("Cart: Request host: %s", c.Request.Host)
	log.Printf("Cart: Is HTTPS: %v", c.Request.TLS != nil)

	if err != nil || sessionID == "" {
		// Generate new session ID using UUID
		sessionID = fmt.Sprintf("session_%s", uuid.New().String())
		// Set secure flag based on whether request is HTTPS
		// Check both TLS and X-Forwarded-Proto header (for proxy setups)
		isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
		log.Printf("Cart: Setting new session_id cookie: %s, secure: %v, TLS: %v, X-Forwarded-Proto: %s", sessionID, isSecure, c.Request.TLS != nil, c.GetHeader("X-Forwarded-Proto"))

		// For production, set cookie domain to work across ethnictreasures.co.in
		cookieDomain := ""
		// Only set domain for production/remote environments, not localhost
		isLocalhost := c.Request.Host == "localhost:8080" ||
			c.GetHeader("Host") == "localhost:8080" ||
			c.Request.Host == "127.0.0.1:8080" ||
			c.GetHeader("Host") == "127.0.0.1:8080"

		log.Printf("Cart: Request host: %s, isLocalhost: %v", c.Request.Host, isLocalhost)

		if !isLocalhost {
			// For production (etreasure-1.onrender.com), set domain to frontend domain
			cookieDomain = os.Getenv("COOKIE_DOMAIN")
			if cookieDomain == "" {
				cookieDomain = "ethnictreasures.co.in"
			}
		}

		log.Printf("Cart: Using cookie domain: %s", cookieDomain)

		c.SetCookie("session_id", sessionID, 86400*30, "/", cookieDomain, isSecure, false) // 30 days, HttpOnly=false for frontend access
		log.Printf("Cart: SetCookie called with session_id: %s, domain: %s, secure: %v", sessionID, cookieDomain, isSecure)

		// Also try to set the cookie manually with proper domain for cross-domain
		domainPart := ""
		if cookieDomain != "" {
			domainPart = fmt.Sprintf("; Domain=%s", cookieDomain)
		}
		// For cross-domain requests, use SameSite=None when domain is set
		sameSiteAttr := "SameSite=Lax"
		if cookieDomain != "" && isSecure {
			sameSiteAttr = "SameSite=None"
		}

		// Ensure we have the proper cookie attributes for cross-domain
		cookieString := fmt.Sprintf("session_id=%s; Path=/; Max-Age=%d; HttpOnly=false; %s%s", sessionID, 86400*30, func() string {
			if isSecure {
				return "Secure; " + sameSiteAttr
			}
			return sameSiteAttr
		}(), domainPart)

		c.Header("Set-Cookie", cookieString)
		log.Printf("Cart: Manual Set-Cookie: %s", cookieString)

		// Log all response headers to debug
		log.Printf("Cart: All response headers: %+v", c.Writer.Header())

		// Test if cookie can be read immediately after setting
		testSessionID, testErr := c.Cookie("session_id")
		log.Printf("Cart: Test reading cookie after setting: %s, error: %v", testSessionID, testErr)
	} else {
		log.Printf("Cart: Using existing session_id: %s", sessionID)
	}

	// Check if item already exists in cart
	var existingQuantity int
	err = h.DB.QueryRow(ctx, `
		SELECT quantity FROM cart 
		WHERE session_id = $1 AND product_id = $2::uuid AND variant_id = $3
	`, sessionID, req.ProductID, variantID).Scan(&existingQuantity)

	if err == nil {
		// Update existing item
		_, err = h.DB.Exec(ctx, `
			UPDATE cart 
			SET quantity = quantity + $1, updated_at = NOW()
			WHERE session_id = $2 AND product_id = $3::uuid AND variant_id = $4
		`, req.Quantity, sessionID, req.ProductID, variantID)
	} else {
		// Insert new item
		_, err = h.DB.Exec(ctx, `
			INSERT INTO cart (session_id, product_id, variant_id, quantity, updated_at)
			VALUES ($1, $2::uuid, $3, $4, NOW())
		`, sessionID, req.ProductID, variantID, req.Quantity)
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

	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	log.Printf("GetCart: Reading session_id cookie: %s, error: %v", sessionID, err)
	log.Printf("GetCart: Request headers: %+v", c.Request.Header)
	log.Printf("GetCart: Request origin: %s", c.GetHeader("Origin"))
	log.Printf("GetCart: Request host: %s", c.Request.Host)
	log.Printf("GetCart: All cookies: %+v", c.Request.Cookies())

	if err != nil || sessionID == "" {
		log.Printf("GetCart: No session found, returning empty cart")
		// Return empty cart for new users
		c.JSON(http.StatusOK, CartResponse{
			Items: []CartItem{},
			Total: 0.0,
			Count: 0,
		})
		return
	}

	log.Printf("GetCart: Using session_id: %s", sessionID)

	query := `
		SELECT 
			c.id,
			c.product_id::text,
			p.title,
			pv.price_cents,
			pv.currency,
			c.quantity,
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

		err := rows.Scan(&id, &productID, &item.Title, &priceCents, &currency, &quantity, &imagePath)
		if err != nil {
			continue
		}

		item.ID = fmt.Sprint(id)
		item.ProductID = productID
		item.Price = float64(priceCents) / 100.0
		item.Quantity = quantity

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
