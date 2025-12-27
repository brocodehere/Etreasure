package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/etreasure/backend/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Helper function to check if user is authenticated
func isAuthenticated(c *gin.Context) bool {
	token := c.GetHeader("Authorization")
	return token != "" && strings.HasPrefix(token, "Bearer ")
}

// Helper function to get user ID from token
func getUserID(c *gin.Context) string {
	token := c.GetHeader("Authorization")
	if token != "" && strings.HasPrefix(token, "Bearer ") {
		// Extract user ID from token (simplified - in real app you'd decode JWT)
		// For now, return empty string to force session-based approach
		return ""
	}
	return ""
}

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
		SELECT p.title, COALESCE(pv.price_cents, 0) as price_cents, COALESCE(pv.currency, 'INR') as currency, 
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

	// Get session ID from cookie or create new one
	sessionID, err := c.Cookie("session_id")
	log.Printf("Wishlist: Reading session_id cookie: %s, error: %v", sessionID, err)
	log.Printf("Wishlist: Request headers: %+v", c.Request.Header)
	log.Printf("Wishlist: Request origin: %s", c.GetHeader("Origin"))
	log.Printf("Wishlist: Request host: %s", c.Request.Host)
	log.Printf("Wishlist: Is HTTPS: %v", c.Request.TLS != nil)

	if err != nil || sessionID == "" {
		// Generate new session ID using UUID
		sessionID = fmt.Sprintf("session_%s", uuid.New().String())
		// Set secure flag based on whether request is HTTPS
		isSecure := c.Request.TLS != nil
		log.Printf("Wishlist: Setting new session_id cookie: %s, secure: %v, TLS: %v", sessionID, isSecure, c.Request.TLS != nil)

		// For production, set cookie domain to work across ethnictreasures.co.in
		cookieDomain := ""
		if isSecure {
			// In production (HTTPS), set domain to work across main site
			// Can be overridden with env var for flexibility
			cookieDomain = os.Getenv("COOKIE_DOMAIN")
			if cookieDomain == "" {
				cookieDomain = "ethnictreasures.co.in"
			}
		}

		log.Printf("Wishlist: Using cookie domain: %s", cookieDomain)

		c.SetCookie("session_id", sessionID, 86400*30, "/", cookieDomain, isSecure, false) // 30 days, HttpOnly=false for frontend access
		log.Printf("Wishlist: SetCookie called with session_id: %s, domain: %s, secure: %v", sessionID, cookieDomain, isSecure)

		// Also try to set the cookie manually as a fallback with proper domain
		domainPart := ""
		if cookieDomain != "" {
			domainPart = fmt.Sprintf("; Domain=%s", cookieDomain)
		}
		// For cross-domain requests, use SameSite=None when Secure
		sameSiteAttr := "SameSite=Lax"
		if isSecure && cookieDomain != "" {
			sameSiteAttr = "SameSite=None"
		}

		cookieString := fmt.Sprintf("session_id=%s; Path=/; Max-Age=%d; %s%s", sessionID, 86400*30, func() string {
			if isSecure {
				return "Secure; " + sameSiteAttr
			}
			return sameSiteAttr
		}(), domainPart)

		c.Header("Set-Cookie", cookieString)
		log.Printf("Wishlist: Manual Set-Cookie header: %s", cookieString)
	} else {
		log.Printf("Wishlist: Using existing session_id: %s", sessionID)
	}

	// Check if item is already in wishlist (using session_id for guest users)
	var existingID any
	err = h.DB.QueryRow(ctx, `
		SELECT id FROM wishlist 
		WHERE session_id = $1 AND product_id = $2::uuid
	`, sessionID, req.ProductID).Scan(&existingID)

	if err == nil {
		// Remove from wishlist
		_, err = h.DB.Exec(ctx, `
			DELETE FROM wishlist 
			WHERE session_id = $1 AND product_id = $2::uuid
		`, sessionID, req.ProductID)

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
			INSERT INTO wishlist (session_id, product_id, created_at)
			VALUES ($1, $2::uuid, NOW())
		`, sessionID, req.ProductID)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add to wishlist", "details": err.Error()})
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

// GetWishlist retrieves the user's wishlist
func (h *WishlistHandler) GetWishlist(c *gin.Context) {
	// Get session ID from cookie (for consistency with toggle)
	sessionID, err := c.Cookie("session_id")
	log.Printf("GetWishlist: Reading session_id cookie: %s, error: %v", sessionID, err)
	log.Printf("GetWishlist: Request headers: %+v", c.Request.Header)
	log.Printf("GetWishlist: Request origin: %s", c.GetHeader("Origin"))
	log.Printf("GetWishlist: Request host: %s", c.Request.Host)
	log.Printf("GetWishlist: All cookies: %+v", c.Request.Cookies())

	if err != nil || sessionID == "" {
		log.Printf("GetWishlist: No session found, returning empty wishlist")
		// Return empty wishlist for new users
		c.JSON(http.StatusOK, WishlistResponse{
			Items: []WishlistItem{},
			Count: 0,
		})
		return
	}

	log.Printf("GetWishlist: Using session_id: %s", sessionID)

	ctx := context.Background()

	// Query with proper JOIN to get wishlist items with product details (using session_id)
	query := `
		SELECT 
			w.id,
			w.product_id::text,
			p.title,
			COALESCE(pv.price_cents, 0) as price_cents,
			COALESCE(pv.currency, 'INR') as currency,
			w.created_at,
			(SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_url
		FROM wishlist w
		JOIN products p ON w.product_id = p.uuid_id
		LEFT JOIN product_variants pv ON p.uuid_id = pv.product_id
		WHERE w.session_id = $1
		ORDER BY w.created_at DESC
	`

	rows, err := h.DB.Query(ctx, query, sessionID)
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
		var id any
		var priceCents int
		var currency string
		var createdAt time.Time
		var imagePath *string
		var productID string

		err := rows.Scan(&id, &productID, &item.Title, &priceCents, &currency, &createdAt, &imagePath)
		if err != nil {
			continue
		}

		item.ID = fmt.Sprint(id)
		item.ProductID = productID
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

	// Get session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No session found"})
		return
	}

	// Delete item from wishlist
	_, err = h.DB.Exec(ctx, `
		DELETE FROM wishlist 
		WHERE session_id = $1 AND product_id = $2::uuid
	`, sessionID, productID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove item from wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Product removed from wishlist successfully",
		"product_id": productID,
	})
}
