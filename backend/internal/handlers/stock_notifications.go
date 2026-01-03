package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/etreasure/backend/internal/email"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type StockNotificationsHandler struct {
	DB    *pgxpool.Pool
	Rd    *redis.Client
	Email *email.EmailService
}

type createStockNotificationRequest struct {
	ProductID        *string `json:"productId"` // Changed to string UUID
	ProductSlug      string  `json:"productSlug" binding:"required"`
	NotificationType string  `json:"notificationType" binding:"required,oneof=email mobile"`
	Email            string  `json:"email,omitempty"`
	MobileNumber     string  `json:"mobileNumber,omitempty"`
}

// CreateStockNotification - creates a new stock notification request
func (h *StockNotificationsHandler) CreateStockNotification(c *gin.Context) {
	var req createStockNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Validate required fields based on notification type
	if req.NotificationType == "email" && req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email is required for email notifications"})
		return
	}
	if req.NotificationType == "mobile" && req.MobileNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mobile number is required for mobile notifications"})
		return
	}

	ctx := context.Background()

	// If productId is not provided, fetch it from the slug
	var productUUID string
	if req.ProductID != nil {
		productUUID = *req.ProductID
	} else {
		// Fetch product UUID from slug
		err := h.DB.QueryRow(ctx, `SELECT uuid_id FROM products WHERE slug = $1`, req.ProductSlug).Scan(&productUUID)
		if err != nil {
			if err == pgx.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			}
			return
		}
	}

	// Check if notification already exists for this product and contact
	var existingID int
	var query string
	var args []interface{}

	if req.NotificationType == "email" {
		query = `SELECT id FROM stock_notifications WHERE product_id = $1 AND email = $2 AND is_active = TRUE AND is_notified = FALSE`
		args = []interface{}{productUUID, req.Email}
	} else {
		query = `SELECT id FROM stock_notifications WHERE product_id = $1 AND mobile_number = $2 AND is_active = TRUE AND is_notified = FALSE`
		args = []interface{}{productUUID, req.MobileNumber}
	}

	err := h.DB.QueryRow(ctx, query, args...).Scan(&existingID)
	if err == nil {
		// Notification already exists
		c.JSON(http.StatusConflict, gin.H{"error": "You have already requested to be notified when this product is back in stock"})
		return
	} else if err != pgx.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Create new notification
	var contactField, contactValue string
	if req.NotificationType == "email" {
		contactField = "email"
		contactValue = req.Email
	} else {
		contactField = "mobile_number"
		contactValue = req.MobileNumber
	}

	query = fmt.Sprintf(`
		INSERT INTO stock_notifications (product_id, product_slug, %s, notification_type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, contactField)

	err = h.DB.QueryRow(ctx, query, productUUID, req.ProductSlug, contactValue, req.NotificationType).Scan(&existingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create notification request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification request created successfully",
		"id":      existingID,
	})
}

// SendStockNotifications - sends notifications when product is back in stock
func (h *StockNotificationsHandler) SendStockNotifications(productUUID string, productSlug string) error {
	ctx := context.Background()

	// Get product details for email
	var productTitle string
	var productImage *string
	var minPriceCents int
	err := h.DB.QueryRow(ctx, `
		SELECT p.title, 
		       (SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1) as image_path,
		       (SELECT MIN(price_cents) FROM product_variants WHERE product_id = p.uuid_id) as min_price
		FROM products p WHERE p.uuid_id = $1
	`, productUUID).Scan(&productTitle, &productImage, &minPriceCents)
	if err != nil {
		return fmt.Errorf("failed to fetch product details: %w", err)
	}

	// Get all active, non-notified notifications for this product
	rows, err := h.DB.Query(ctx, `
		SELECT id, email, mobile_number, notification_type
		FROM stock_notifications 
		WHERE product_id = $1 AND is_active = TRUE AND is_notified = FALSE
	`, productUUID)
	if err != nil {
		return fmt.Errorf("failed to fetch notifications: %w", err)
	}
	defer rows.Close()

	var notifications []struct {
		ID               int
		Email            *string
		MobileNumber     *string
		NotificationType string
	}

	for rows.Next() {
		var n struct {
			ID               int
			Email            *string
			MobileNumber     *string
			NotificationType string
		}
		err := rows.Scan(&n.ID, &n.Email, &n.MobileNumber, &n.NotificationType)
		if err != nil {
			continue
		}
		notifications = append(notifications, n)
	}

	if len(notifications) == 0 {
		return nil // No notifications to send
	}

	// Send email notifications
	for _, notification := range notifications {
		if notification.NotificationType == "email" && notification.Email != nil {
			err := h.sendStockNotificationEmail(*notification.Email, productSlug, productTitle, productImage, minPriceCents)
			if err != nil {
				fmt.Printf("Failed to send stock notification email to %s: %v\n", *notification.Email, err)
			}
		}
		// TODO: Implement SMS notifications for mobile numbers
	}

	// Mark all notifications as sent
	_, err = h.DB.Exec(ctx, `
		UPDATE stock_notifications 
		SET is_notified = TRUE, updated_at = NOW()
		WHERE product_id = $1 AND is_active = TRUE AND is_notified = FALSE
	`, productUUID)
	if err != nil {
		return fmt.Errorf("failed to mark notifications as sent: %w", err)
	}

	return nil
}

func (h *StockNotificationsHandler) sendStockNotificationEmail(email, productSlug, productTitle string, productImage *string, minPriceCents int) error {
	if h.Email == nil {
		return fmt.Errorf("email service not configured")
	}

	return h.Email.SendStockNotificationEmail(email, productSlug, productTitle, productImage, minPriceCents)
}
