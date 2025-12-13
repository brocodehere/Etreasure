package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Order struct {
	ID             string  `json:"id"`
	OrderNumber    string  `json:"order_number"`
	UserID         *int    `json:"user_id" db:"user_id"`
	Status         string  `json:"status"` // pending | processing | shipped | delivered | cancelled
	Currency       string  `json:"currency"`
	TotalPrice     float64 `json:"total_price"`
	Subtotal       float64 `json:"subtotal"`
	TaxAmount      float64 `json:"tax_amount"`
	ShippingAmount float64 `json:"shipping_amount"`
	DiscountAmount float64 `json:"discount_amount"`
	// Customer Details
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
	// Shipping Details
	ShippingName         string `json:"shipping_name"`
	ShippingEmail        string `json:"shipping_email"`
	ShippingPhone        string `json:"shipping_phone"`
	ShippingAddressLine1 string `json:"shipping_address_line1"`
	ShippingAddressLine2 string `json:"shipping_address_line2"`
	ShippingCity         string `json:"shipping_city"`
	ShippingState        string `json:"shipping_state"`
	ShippingCountry      string `json:"shipping_country"`
	ShippingPinCode      string `json:"shipping_pin_code"`
	// Billing Details
	BillingName         string `json:"billing_name"`
	BillingEmail        string `json:"billing_email"`
	BillingPhone        string `json:"billing_phone"`
	BillingAddressLine1 string `json:"billing_address_line1"`
	BillingAddressLine2 string `json:"billing_address_line2"`
	BillingCity         string `json:"billing_city"`
	BillingState        string `json:"billing_state"`
	BillingCountry      string `json:"billing_country"`
	BillingPinCode      string `json:"billing_pin_code"`
	// Payment Details
	PaymentMethod     string  `json:"payment_method"`
	PaymentID         *string `json:"payment_id,omitempty"`
	RazorpayOrderID   *string `json:"razorpay_order_id,omitempty"`
	RazorpayPaymentID *string `json:"razorpay_payment_id,omitempty"`
	RazorpaySignature *string `json:"razorpay_signature,omitempty"`
	// Tracking
	TrackingNumber    *string    `json:"tracking_number,omitempty"`
	TrackingProvider  *string    `json:"tracking_provider,omitempty"`
	EstimatedDelivery *time.Time `json:"estimated_delivery,omitempty"`
	// Metadata
	Notes     *string         `json:"notes,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	LineItems []OrderLineItem `json:"line_items"`
}

type OrderLineItem struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	ProductID string  `json:"product_id"`
	VariantID *string `json:"variant_id,omitempty"`
	Title     string  `json:"title"`
	SKU       string  `json:"sku"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
	Total     float64 `json:"total"`
}

type CreateOrderRequest struct {
	UserID       *int                  `json:"user_id,omitempty"`
	Currency     string                `json:"currency" binding:"required"`
	LineItems    []CreateOrderLineItem `json:"line_items" binding:"required,min=1"`
	ShippingAddr *Address              `json:"shipping_address,omitempty"`
	BillingAddr  *Address              `json:"billing_address,omitempty"`
	Notes        *string               `json:"notes,omitempty"`
}

type CreateOrderLineItem struct {
	ProductID string  `json:"product_id" binding:"required"`
	VariantID *string `json:"variant_id,omitempty"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	Price     float64 `json:"price" binding:"required,min=0"`
}

type UpdateOrderRequest struct {
	Status       *string  `json:"status,omitempty"`
	ShippingAddr *Address `json:"shipping_address,omitempty"`
	BillingAddr  *Address `json:"billing_address,omitempty"`
	Notes        *string  `json:"notes,omitempty"`
}

func (h *Handler) ListOrders(c *gin.Context) {

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	cursor := c.Query("cursor")
	var cursorTime time.Time
	if cursor != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, cursor); err == nil {
			cursorTime = parsed
		}
	}

	// Simple query without cursor filtering for debugging
	var query string
	var args []interface{}

	if cursor != "" {
		query = `
			SELECT 
				id, order_number, user_id, status, currency, total_price, subtotal,
				tax_amount, shipping_amount, discount_amount,
				COALESCE(customer_name, 'Guest Customer') as customer_name,
				COALESCE(customer_email, 'guest@example.com') as customer_email,
				COALESCE(customer_phone, '0000000000') as customer_phone,
				COALESCE(payment_method, 'cod') as payment_method,
				payment_id, razorpay_order_id, razorpay_payment_id, razorpay_signature,
				notes, created_at, updated_at
			FROM orders
			WHERE updated_at < $1
			ORDER BY updated_at DESC
			LIMIT $2
		`
		args = []interface{}{cursorTime, limit}
	} else {
		query = `
			SELECT 
				id, order_number, user_id, status, currency, total_price, subtotal,
				tax_amount, shipping_amount, discount_amount,
				COALESCE(customer_name, 'Guest Customer') as customer_name,
				COALESCE(customer_email, 'guest@example.com') as customer_email,
				COALESCE(customer_phone, '0000000000') as customer_phone,
				COALESCE(payment_method, 'cod') as payment_method,
				payment_id, razorpay_order_id, razorpay_payment_id, razorpay_signature,
				notes, created_at, updated_at
			FROM orders
			ORDER BY updated_at DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	rows, err := h.DB.Query(c, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var orders []Order
	orderCount := 0
	for rows.Next() {
		orderCount++
		var o Order
		err := rows.Scan(
			&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.Currency,
			&o.TotalPrice, &o.Subtotal, &o.TaxAmount, &o.ShippingAmount,
			&o.DiscountAmount,
			&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
			&o.PaymentMethod, &o.PaymentID, &o.RazorpayOrderID, &o.RazorpayPaymentID, &o.RazorpaySignature,
			&o.Notes, &o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Set default values for missing fields
		o.ShippingName = ""
		o.ShippingEmail = ""
		o.ShippingPhone = ""
		o.ShippingAddressLine1 = ""
		o.ShippingAddressLine2 = ""
		o.ShippingCity = ""
		o.ShippingState = ""
		o.ShippingCountry = "India"
		o.ShippingPinCode = ""
		o.BillingName = ""
		o.BillingEmail = ""
		o.BillingPhone = ""
		o.BillingAddressLine1 = ""
		o.BillingAddressLine2 = ""
		o.BillingCity = ""
		o.BillingState = ""
		o.BillingCountry = "India"
		o.BillingPinCode = ""
		o.TrackingNumber = nil
		o.TrackingProvider = nil
		o.EstimatedDelivery = nil

		// Load line items
		o.LineItems, err = h.loadOrderLineItems(c, o.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		orders = append(orders, o)
	}

	var nextCursor *string
	if len(orders) == limit {
		next := orders[len(orders)-1].UpdatedAt.Format(time.RFC3339Nano)
		nextCursor = &next
	}

	response := gin.H{
		"data": orders,
	}

	if nextCursor != nil {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	orderNumber := generateOrderNumber()

	var subtotal, totalPrice float64
	for _, item := range req.LineItems {
		subtotal += item.Price * float64(item.Quantity)
	}
	totalPrice = subtotal // TODO: add tax and shipping later

	shippingAddrJSON, _ := json.Marshal(req.ShippingAddr)
	billingAddrJSON, _ := json.Marshal(req.BillingAddr)

	_, err := h.DB.Exec(c, `
		INSERT INTO orders (id, order_number, user_id, status, currency, total_price, subtotal,
							tax_amount, shipping_amount, discount_amount, shipping_address, billing_address,
							notes, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', $4, $5, $6, 0, 0, 0, $7, $8, $9, $10, $10)
	`, id, orderNumber, req.UserID, req.Currency, totalPrice, subtotal,
		shippingAddrJSON, billingAddrJSON, req.Notes, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert line items
	for _, item := range req.LineItems {
		lineItemID := uuid.New().String()
		_, err = h.DB.Exec(c, `
			INSERT INTO order_line_items (id, order_id, product_id, variant_id, title, sku, quantity, price, total)
			VALUES ($1, $2, $3, $4, '', '', $5, $6, $7)
		`, lineItemID, id, item.ProductID, item.VariantID, item.Quantity, item.Price, item.Price*float64(item.Quantity))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	order := Order{
		ID:             id,
		OrderNumber:    orderNumber,
		UserID:         req.UserID,
		Status:         "pending",
		Currency:       req.Currency,
		TotalPrice:     totalPrice,
		Subtotal:       subtotal,
		TaxAmount:      0,
		ShippingAmount: 0,
		DiscountAmount: 0,
		Notes:          req.Notes,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	order.LineItems, _ = h.loadOrderLineItems(c, id)

	c.JSON(http.StatusCreated, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	var o Order
	err := h.DB.QueryRow(c, `
		SELECT 
			id, order_number, user_id, status, currency, total_price, subtotal,
			tax_amount, shipping_amount, discount_amount,
			COALESCE(customer_name, 'Guest Customer') as customer_name,
			COALESCE(customer_email, 'guest@example.com') as customer_email,
			COALESCE(customer_phone, '0000000000') as customer_phone,
			COALESCE(shipping_name, '') as shipping_name,
			COALESCE(shipping_email, '') as shipping_email,
			COALESCE(shipping_phone, '') as shipping_phone,
			COALESCE(shipping_address_line1, '') as shipping_address_line1,
			COALESCE(shipping_address_line2, '') as shipping_address_line2,
			COALESCE(shipping_city, '') as shipping_city,
			COALESCE(shipping_state, '') as shipping_state,
			COALESCE(shipping_country, 'India') as shipping_country,
			COALESCE(shipping_pin_code, '') as shipping_pin_code,
			COALESCE(billing_name, '') as billing_name,
			COALESCE(billing_email, '') as billing_email,
			COALESCE(billing_phone, '') as billing_phone,
			COALESCE(billing_address_line1, '') as billing_address_line1,
			COALESCE(billing_address_line2, '') as billing_address_line2,
			COALESCE(billing_city, '') as billing_city,
			COALESCE(billing_state, '') as billing_state,
			COALESCE(billing_country, 'India') as billing_country,
			COALESCE(billing_pin_code, '') as billing_pin_code,
			COALESCE(payment_method, 'cod') as payment_method,
			payment_id, razorpay_order_id, razorpay_payment_id, razorpay_signature,
			tracking_number, tracking_provider, estimated_delivery,
			notes, created_at, updated_at
		FROM orders
		WHERE id = $1
	`, id).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.Currency,
		&o.TotalPrice, &o.Subtotal, &o.TaxAmount, &o.ShippingAmount,
		&o.DiscountAmount,
		&o.CustomerName, &o.CustomerEmail, &o.CustomerPhone,
		&o.ShippingName, &o.ShippingEmail, &o.ShippingPhone, &o.ShippingAddressLine1,
		&o.ShippingAddressLine2, &o.ShippingCity, &o.ShippingState, &o.ShippingCountry, &o.ShippingPinCode,
		&o.BillingName, &o.BillingEmail, &o.BillingPhone, &o.BillingAddressLine1,
		&o.BillingAddressLine2, &o.BillingCity, &o.BillingState, &o.BillingCountry, &o.BillingPinCode,
		&o.PaymentMethod, &o.PaymentID, &o.RazorpayOrderID, &o.RazorpayPaymentID, &o.RazorpaySignature,
		&o.TrackingNumber, &o.TrackingProvider, &o.EstimatedDelivery,
		&o.Notes, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	o.LineItems, err = h.loadOrderLineItems(c, o.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, o)
}

func (h *Handler) UpdateOrder(c *gin.Context) {
	id := c.Param("id")
	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sets := []string{}
	args := []any{1, id}
	argIdx := 2

	if req.Status != nil {
		sets = append(sets, "status = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Status)
		argIdx++
	}
	if req.Notes != nil {
		sets = append(sets, "notes = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Notes)
		argIdx++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	sets = append(sets, "updated_at = NOW()")
	query := "UPDATE orders SET " + joinString(sets, ", ") + " WHERE id = $" + strconv.Itoa(argIdx)
	args = append(args, id)

	_, err := h.DB.Exec(c, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated order
	var o Order
	var shippingAddrJSON, billingAddrJSON []byte
	err = h.DB.QueryRow(c, `
		SELECT id, order_number, user_id, status, currency, total_price, subtotal,
			   tax_amount, shipping_amount, discount_amount, shipping_address, billing_address,
			   notes, created_at, updated_at
		FROM orders
		WHERE id = $1
	`, id).Scan(
		&o.ID, &o.OrderNumber, &o.UserID, &o.Status, &o.Currency,
		&o.TotalPrice, &o.Subtotal, &o.TaxAmount, &o.ShippingAmount,
		&o.DiscountAmount, &shippingAddrJSON, &billingAddrJSON, &o.Notes,
		&o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	o.LineItems, _ = h.loadOrderLineItems(c, o.ID)

	c.JSON(http.StatusOK, o)
}

func (h *Handler) DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	_, err := h.DB.Exec(c, "DELETE FROM orders WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) loadOrderLineItems(c *gin.Context, orderID string) ([]OrderLineItem, error) {
	// Try with new column names first
	rows, err := h.DB.Query(c, `
		SELECT id, order_id, product_id, variant_id, product_title, product_sku, quantity, unit_price, total_price, product_image_url
		FROM order_line_items
		WHERE order_id = $1
	`, orderID)
	if err != nil {
		// Fallback to old column names if new ones don't exist
		rows, err = h.DB.Query(c, `
			SELECT id, order_id, product_id, variant_id, title, sku, quantity, price, total, NULL as product_image_url
			FROM order_line_items
			WHERE order_id = $1
		`, orderID)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()

	var items []OrderLineItem
	for rows.Next() {
		var li OrderLineItem
		var productImageURL *string
		var variantID *int
		var title, sku string
		var price, totalPrice float64

		// Try scanning with new column names first
		err = rows.Scan(&li.ID, &li.OrderID, &li.ProductID, &variantID, &title, &sku,
			&li.Quantity, &price, &totalPrice, &productImageURL)
		if err != nil {
			// Fallback to old column names
			err = rows.Scan(&li.ID, &li.OrderID, &li.ProductID, &variantID, &title, &sku,
				&li.Quantity, &price, &totalPrice, &productImageURL)
			if err != nil {
				return nil, err
			}
		}

		li.Title = title
		li.SKU = sku
		li.Price = price
		li.Total = totalPrice

		if variantID != nil {
			variantIDStr := fmt.Sprintf("%d", *variantID)
			li.VariantID = &variantIDStr
		}
		items = append(items, li)
	}
	return items, nil
}

func generateOrderNumber() string {
	// Simple order number: ORD-YYYYMMDD-XXXX where XXXX is random
	now := time.Now().UTC()
	prefix := now.Format("20060102")
	suffix := strings.ToUpper(uuid.New().String()[:4])
	return "ORD-" + prefix + "-" + suffix
}
