package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type InventoryItem struct {
	ID        string    `json:"id"`
	ProductID string    `json:"product_id"`
	VariantID *string   `json:"variant_id,omitempty"`
	SKU       string    `json:"sku"`
	Quantity  int       `json:"quantity"`
	Reserved  int       `json:"reserved"`
	Available int       `json:"available"`
	Location  *string   `json:"location,omitempty"`
	CostPrice *float64  `json:"cost_price,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInventoryItemRequest struct {
	ProductID string   `json:"product_id" binding:"required"`
	VariantID *string  `json:"variant_id,omitempty"`
	SKU       string   `json:"sku" binding:"required"`
	Quantity  int      `json:"quantity" binding:"required,min=0"`
	Location  *string  `json:"location,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
}

type UpdateInventoryItemRequest struct {
	Quantity  *int     `json:"quantity,omitempty"`
	Location  *string  `json:"location,omitempty"`
	CostPrice *float64 `json:"cost_price,omitempty"`
}

type AdjustmentRequest struct {
	Quantity int    `json:"quantity" binding:"required"` // positive to add, negative to subtract
	Reason   string `json:"reason" binding:"required"`
}

func (h *Handler) ListInventory(c *gin.Context) {
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

	rows, err := h.DB.Query(c, `
		SELECT id, product_id, variant_id, sku, quantity, reserved, available, location, cost_price, updated_at
		FROM inventory_items
		WHERE ($1::timestamp IS NULL OR updated_at < $1::timestamp)
		ORDER BY updated_at DESC
		LIMIT $2
	`, cursorTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []InventoryItem
	for rows.Next() {
		var it InventoryItem
		err := rows.Scan(
			&it.ID, &it.ProductID, &it.VariantID, &it.SKU,
			&it.Quantity, &it.Reserved, &it.Available, &it.Location, &it.CostPrice, &it.UpdatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, it)
	}

	var nextCursor *string
	if len(items) == limit {
		next := items[len(items)-1].UpdatedAt.Format(time.RFC3339Nano)
		nextCursor = &next
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        items,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) CreateInventoryItem(c *gin.Context) {
	var req CreateInventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	_, err := h.DB.Exec(c, `
		INSERT INTO inventory_items (id, product_id, variant_id, sku, quantity, reserved, available, location, cost_price, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $5, $6, $7, $8)
	`, id, req.ProductID, req.VariantID, req.SKU, req.Quantity, req.Location, req.CostPrice, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	item := InventoryItem{
		ID:        id,
		ProductID: req.ProductID,
		VariantID: req.VariantID,
		SKU:       req.SKU,
		Quantity:  req.Quantity,
		Reserved:  0,
		Available: req.Quantity,
		Location:  req.Location,
		CostPrice: req.CostPrice,
		UpdatedAt: now,
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) GetInventoryItem(c *gin.Context) {
	id := c.Param("id")
	var it InventoryItem
	err := h.DB.QueryRow(c, `
		SELECT id, product_id, variant_id, sku, quantity, reserved, available, location, cost_price, updated_at
		FROM inventory_items
		WHERE id = $1
	`, id).Scan(
		&it.ID, &it.ProductID, &it.VariantID, &it.SKU,
		&it.Quantity, &it.Reserved, &it.Available, &it.Location, &it.CostPrice, &it.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, it)
}

func (h *Handler) UpdateInventoryItem(c *gin.Context) {
	id := c.Param("id")
	var req UpdateInventoryItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sets := []string{}
	args := []any{1, id}
	argIdx := 2

	if req.Quantity != nil {
		sets = append(sets, "quantity = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Quantity)
		argIdx++
	}
	if req.Location != nil {
		sets = append(sets, "location = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Location)
		argIdx++
	}
	if req.CostPrice != nil {
		sets = append(sets, "cost_price = $"+strconv.Itoa(argIdx))
		args = append(args, *req.CostPrice)
		argIdx++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	sets = append(sets, "updated_at = NOW()")
	query := "UPDATE inventory_items SET " + joinString(sets, ", ") + " WHERE id = $" + strconv.Itoa(argIdx)
	args = append(args, id)

	tx, err := h.DB.Begin(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(c)

	_, err = tx.Exec(c, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Recalculate available = quantity - reserved
	_, err = tx.Exec(c, "UPDATE inventory_items SET available = quantity - reserved WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated item
	var it InventoryItem
	err = h.DB.QueryRow(c, `
		SELECT id, product_id, variant_id, sku, quantity, reserved, available, location, cost_price, updated_at
		FROM inventory_items
		WHERE id = $1
	`, id).Scan(
		&it.ID, &it.ProductID, &it.VariantID, &it.SKU,
		&it.Quantity, &it.Reserved, &it.Available, &it.Location, &it.CostPrice, &it.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, it)
}

func (h *Handler) DeleteInventoryItem(c *gin.Context) {
	id := c.Param("id")
	_, err := h.DB.Exec(c, "DELETE FROM inventory_items WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) AdjustInventory(c *gin.Context) {
	id := c.Param("id")
	var req AdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.DB.Begin(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback(c)

	// Get current quantity
	var currentQty int
	err = tx.QueryRow(c, "SELECT quantity FROM inventory_items WHERE id = $1", id).Scan(&currentQty)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "inventory item not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newQty := currentQty + req.Quantity
	if newQty < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "adjustment would result in negative quantity"})
		return
	}

	// Update quantity
	_, err = tx.Exec(c, `
		UPDATE inventory_items
		SET quantity = $1, available = quantity - reserved, updated_at = NOW()
		WHERE id = $2
	`, newQty, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log adjustment (optional, could be a separate table)
	_, err = tx.Exec(c, `
		INSERT INTO audit_log (id, user_id, user_email, action, resource_type, resource_id, details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
	`, uuid.New().String(), c.GetString("userID"), c.GetString("userEmail"), "adjust_inventory", "inventory_item", id,
		map[string]interface{}{"adjustment": req.Quantity, "reason": req.Reason, "new_quantity": newQty},
		c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		// Non-critical, ignore
	}

	if err := tx.Commit(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated item
	var it InventoryItem
	err = h.DB.QueryRow(c, `
		SELECT id, product_id, variant_id, sku, quantity, reserved, available, location, cost_price, updated_at
		FROM inventory_items
		WHERE id = $1
	`, id).Scan(
		&it.ID, &it.ProductID, &it.VariantID, &it.SKU,
		&it.Quantity, &it.Reserved, &it.Available, &it.Location, &it.CostPrice, &it.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, it)
}
