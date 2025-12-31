package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/etreasure/backend/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RazorpayHandler struct {
	DB  *pgxpool.Pool
	Cfg config.Config
}

type createPaymentRequest struct {
	Items []struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	} `json:"items"`
	Customer struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	} `json:"customer"`
	ShippingAddress map[string]any `json:"shipping_address"`
}

type razorpayOrderResponse struct {
	ID     string `json:"id"`
	Amount int    `json:"amount"`
	Status string `json:"status"`
}

// CreatePayment validates the cart, creates an internal pending order and a Razorpay order
func (h *RazorpayHandler) CreatePayment(c *gin.Context) {
	if h.Cfg.RazorpayKeyID == "" || h.Cfg.RazorpaySecret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment not configured"})
		return
	}

	var req createPaymentRequest
	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := c.Request.Context()

	// Get user_id from context if user is logged in
	var userID *int
	if val, exists := c.Get("user_id"); exists {
		if uid, ok := val.(int); ok {
			userID = &uid
			log.Printf("CreatePayment: Found user_id in context: %d", uid)
		} else {
			log.Printf("CreatePayment: user_id in context but not int type: %T", val)
		}
	} else {
		log.Printf("CreatePayment: No user_id found in context - user may not be logged in")
	}

	// Compute amount from the server-side cart for this session.
	// This keeps UI cart total and Razorpay amount consistent and avoids client-side manipulation.
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = strings.TrimSpace(c.GetHeader("X-Session-Id"))
		if sessionID == "" {
			sessionID = fmt.Sprintf("session_%d", len(body)+len(c.Request.RemoteAddr))
		}
		// Set secure flag based on whether request is HTTPS
		// Check both TLS and X-Forwarded-Proto header (for proxy setups)
		isSecure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"

		// For production, set cookie domain to work across ethnictreasures.co.in
		cookieDomain := ""
		// Only set domain for production/remote environments, not localhost
		isLocalhost := c.Request.Host == "localhost:8080" ||
			c.GetHeader("Host") == "localhost:8080" ||
			c.Request.Host == "127.0.0.1:8080" ||
			c.GetHeader("Host") == "127.0.0.1:8080"

		if !isLocalhost {
			cookieDomain = os.Getenv("COOKIE_DOMAIN")
			if cookieDomain == "" {
				cookieDomain = "ethnictreasures.co.in"
			}
		}

		// Industry standard: Set single cookie with proper attributes for cross-domain
		// For cross-domain scenarios, we need Domain=ethnictreasures.co.in with SameSite=None; Secure
		// For localhost scenarios, we need no domain with SameSite=Lax

		var cookieString string
		if cookieDomain != "" && isSecure {
			// Production: Cross-domain cookie
			cookieString = fmt.Sprintf("session_id=%s; Path=/; Domain=%s; Max-Age=%d; HttpOnly=false; SameSite=None; Secure",
				sessionID, cookieDomain, 86400*30)
		} else {
			// Localhost: Simple cookie
			cookieString = fmt.Sprintf("session_id=%s; Path=/; Max-Age=%d; HttpOnly=false; SameSite=Lax%s",
				sessionID, 86400*30, func() string {
					if isSecure {
						return "; Secure"
					}
					return ""
				}())
		}

		c.Header("Set-Cookie", cookieString)
	}

	var subtotal int
	err = h.DB.QueryRow(ctx, `
		SELECT COALESCE(SUM(pv.price_cents * c.quantity), 0)
		FROM cart c
		JOIN product_variants pv ON c.variant_id = pv.id
		WHERE c.session_id = $1
	`, sessionID).Scan(&subtotal)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") || strings.Contains(err.Error(), "relation") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cart", "details": err.Error()})
		return
	}
	if subtotal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cart is empty"})
		return
	}

	// Keep tax and shipping at 0 to match the cart UI.
	tax := 0
	shipping := 0
	total := subtotal

	// Extract shipping address details
	shippingAddr := req.ShippingAddress
	shippingAddrLine1 := ""
	shippingCity := ""
	shippingState := ""
	shippingPinCode := ""

	if val, exists := shippingAddr["address"]; exists {
		shippingAddrLine1 = fmt.Sprintf("%v", val)
	}
	if val, exists := shippingAddr["city"]; exists {
		shippingCity = fmt.Sprintf("%v", val)
	}
	if val, exists := shippingAddr["state"]; exists {
		shippingState = fmt.Sprintf("%v", val)
	}
	if val, exists := shippingAddr["pin_code"]; exists {
		shippingPinCode = fmt.Sprintf("%v", val)
	}

	// Insert complete order with customer details
	var orderID string
	err = h.DB.QueryRow(ctx, `
    INSERT INTO orders (
        order_number, status, currency, total_price, subtotal, tax_amount, shipping_amount,
        customer_name, customer_email, customer_phone,
        shipping_name, shipping_email, shipping_phone, shipping_address_line1,
        shipping_city, shipping_state, shipping_country, shipping_pin_code,
        billing_name, billing_email, billing_phone, billing_address_line1,
        billing_city, billing_state, billing_country, billing_pin_code,
        payment_method, user_id
    ) VALUES (
        gen_random_uuid()::text, 'pending_payment', 'INR', $1, $2, $3, $4,
        $5, $6, $7, $8, $9, $10, $11, $12, $13, 'India', $14,
        $5, $6, $7, $11, $12, $13, 'India', $14, 'razorpay', $15
    ) RETURNING id
  `,
		float64(total)/100.0, float64(subtotal)/100.0, float64(tax)/100.0, float64(shipping)/100.0,
		req.Customer.Name, req.Customer.Email, req.Customer.Phone,
		req.Customer.Name, req.Customer.Email, req.Customer.Phone, shippingAddrLine1,
		shippingCity, shippingState, shippingPinCode, userID).Scan(&orderID)

	if userID != nil {
		log.Printf("CreatePayment: Storing order with user_id: %d", *userID)
	} else {
		log.Printf("CreatePayment: Storing order with NULL user_id (user not logged in)")
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order", "details": err.Error()})
		return
	}

	// Insert order line items from cart
	rows, err := h.DB.Query(ctx, `
		SELECT c.product_id::text, c.variant_id, c.quantity
		FROM cart c
		WHERE c.session_id = $1
	`, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load cart items", "details": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var productID string
		var variantID int
		var qty int
		if err := rows.Scan(&productID, &variantID, &qty); err != nil {
			continue
		}

		var productTitle, productSKU, productImageURL string
		var itemPriceCents int
		err := h.DB.QueryRow(ctx, `
			SELECT p.title, pv.sku,
				   (SELECT m.path FROM product_images pi JOIN media m ON pi.media_id = m.id WHERE pi.product_id = p.uuid_id ORDER BY pi.sort_order LIMIT 1),
				   pv.price_cents
			FROM products p
			JOIN product_variants pv ON p.uuid_id = pv.product_id
			WHERE pv.id = $1
		`, variantID).Scan(&productTitle, &productSKU, &productImageURL, &itemPriceCents)
		if err != nil {
			continue
		}

		_, err = h.DB.Exec(ctx, `
			INSERT INTO order_line_items (
				order_id, product_id, variant_id, product_title, product_sku,
				product_image_url, quantity, price, total
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, orderID, productID, variantID, productTitle, productSKU, productImageURL,
			qty, float64(itemPriceCents)/100.0, float64(itemPriceCents*qty)/100.0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create order line item", "details": err.Error()})
			return
		}
	}

	// Create Razorpay order via REST
	client := &http.Client{}
	payload := map[string]any{
		"amount":          total,
		"currency":        "INR",
		"receipt":         orderID,
		"payment_capture": 1,
	}
	body, _ = json.Marshal(payload)
	reqHTTP, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.razorpay.com/v1/orders", bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to init payment"})
		return
	}
	reqHTTP.SetBasicAuth(h.Cfg.RazorpayKeyID, h.Cfg.RazorpaySecret)
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(reqHTTP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create payment"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment gateway error"})
		return
	}

	var rpResp razorpayOrderResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid payment response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order_id":          orderID,
		"razorpay_order_id": rpResp.ID,
		"amount":            total,
		"currency":          "INR",
		"key":               h.Cfg.RazorpayKeyID,
	})
}

type verifyPaymentRequest struct {
	OrderID           string `json:"order_id"`
	RazorpayOrderID   string `json:"razorpay_order_id"`
	RazorpayPaymentID string `json:"razorpay_payment_id"`
	RazorpaySignature string `json:"razorpay_signature"`
}

// VerifyPayment verifies Razorpay signature and marks order as paid
func (h *RazorpayHandler) VerifyPayment(c *gin.Context) {
	log.Printf("VerifyPayment: Starting payment verification")

	if h.Cfg.RazorpaySecret == "" {
		log.Printf("VerifyPayment: Razorpay secret not configured")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "payment not configured"})
		return
	}

	var req verifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("VerifyPayment: Invalid payload - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	log.Printf("VerifyPayment: Processing order %s, razorpay_order_id %s", req.OrderID, req.RazorpayOrderID)

	data := req.RazorpayOrderID + "|" + req.RazorpayPaymentID
	mac := hmac.New(sha256.New, []byte(h.Cfg.RazorpaySecret))
	mac.Write([]byte(data))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(req.RazorpaySignature)) {
		log.Printf("VerifyPayment: Invalid signature - expected %s, got %s", expected, req.RazorpaySignature)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature"})
		return
	}

	log.Printf("VerifyPayment: Signature verified successfully")

	ctx := c.Request.Context()

	// Get user_id from context if user is logged in
	var userID *int
	if val, exists := c.Get("user_id"); exists {
		if uid, ok := val.(int); ok {
			userID = &uid
			log.Printf("VerifyPayment: Found user_id in context: %d", uid)
		} else {
			log.Printf("VerifyPayment: user_id in context but not int type: %T", val)
		}
	} else {
		log.Printf("VerifyPayment: No user_id found in context - user may not be logged in")
	}

	// Start transaction for stock management
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		log.Printf("VerifyPayment: Failed to start transaction - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	log.Printf("VerifyPayment: Transaction started successfully")

	// Mark order as paid and store payment details
	_, err = tx.Exec(ctx, `
    UPDATE orders SET 
        status = 'paid', 
        updated_at = NOW(),
        razorpay_order_id = $2,
        razorpay_payment_id = $3,
        razorpay_signature = $4,
        user_id = $5
    WHERE id = $1
  `, req.OrderID, req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature, userID)
	if err != nil {
		log.Printf("VerifyPayment: Failed to update order - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order"})
		return
	}

	if userID != nil {
		log.Printf("VerifyPayment: Updated order %s with user_id: %d", req.OrderID, *userID)
	} else {
		log.Printf("VerifyPayment: Updated order %s with NULL user_id (user not logged in)", req.OrderID)
	}

	log.Printf("VerifyPayment: Order %s updated to paid status", req.OrderID)

	// Stock management removed from payment verification
	// Stock is now only managed when adding items to cart
	log.Printf("VerifyPayment: Skipping stock deduction - stock managed at cart level")

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("VerifyPayment: Failed to commit transaction - %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit transaction"})
		return
	}

	log.Printf("VerifyPayment: Transaction committed successfully for order %s", req.OrderID)

	// Clear cart after successful payment (session-based cart)
	if sessionID, errCookie := c.Cookie("session_id"); errCookie == nil && sessionID != "" {
		log.Printf("VerifyPayment: Clearing cart for session %s", sessionID)
		_, _ = h.DB.Exec(ctx, `DELETE FROM cart WHERE session_id = $1`, sessionID)
	}

	log.Printf("VerifyPayment: Payment verification completed successfully for order %s", req.OrderID)
	c.JSON(http.StatusOK, gin.H{"order_id": req.OrderID, "status": "paid"})
}
