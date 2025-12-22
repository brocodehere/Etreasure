package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Offer struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Description   *string   `json:"description,omitempty"`
	DiscountType  string    `json:"discount_type"` // percentage | fixed
	DiscountValue float64   `json:"discount_value"`
	AppliesTo     string    `json:"applies_to"` // all | products | categories | collections
	AppliesToIds  []string  `json:"applies_to_ids"`
	MinOrderAmt   *float64  `json:"min_order_amount,omitempty"`
	UsageLimit    *int      `json:"usage_limit,omitempty"`
	UsageCount    int       `json:"usage_count"`
	IsActive      bool      `json:"is_active"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateOfferRequest struct {
	Title         string   `json:"title" binding:"required"`
	Description   *string  `json:"description,omitempty"`
	DiscountType  string   `json:"discount_type" binding:"required,oneof=percentage fixed"`
	DiscountValue float64  `json:"discount_value" binding:"required,min=0"`
	AppliesTo     string   `json:"applies_to" binding:"required,oneof=all products categories collections"`
	AppliesToIds  []string `json:"applies_to_ids"`
	MinOrderAmt   *float64 `json:"min_order_amount,omitempty"`
	UsageLimit    *int     `json:"usage_limit,omitempty"`
	IsActive      bool     `json:"is_active"`
	StartsAt      *string  `json:"starts_at" binding:"required"`
	EndsAt        *string  `json:"ends_at" binding:"required"`
}

type UpdateOfferRequest struct {
	Title         *string    `json:"title,omitempty"`
	Description   *string    `json:"description,omitempty"`
	DiscountType  *string    `json:"discount_type,omitempty"`
	DiscountValue *float64   `json:"discount_value,omitempty"`
	AppliesTo     *string    `json:"applies_to,omitempty"`
	AppliesToIds  *[]string  `json:"applies_to_ids,omitempty"`
	MinOrderAmt   *float64   `json:"min_order_amount,omitempty"`
	UsageLimit    *int       `json:"usage_limit,omitempty"`
	IsActive      *bool      `json:"is_active,omitempty"`
	StartsAt      *time.Time `json:"starts_at,omitempty"`
	EndsAt        *time.Time `json:"ends_at,omitempty"`
}

func (h *Handler) ListOffers(c *gin.Context) {
	ctx := c.Request.Context()

	// Simple query without prepared statement conflicts
	query := `
		SELECT id, title, description, discount_type, discount_value, applies_to, applies_to_ids,
		       min_order_amount, usage_limit, usage_count, is_active, starts_at, ends_at, created_at, updated_at
		FROM offers
		WHERE is_active = true AND starts_at <= NOW() AND ends_at >= NOW()
		ORDER BY updated_at DESC
		LIMIT 50
	`

	rows, err := h.DB.Query(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch offers"})
		return
	}
	defer rows.Close()

	var offers []Offer
	for rows.Next() {
		var o Offer
		var appliesToIdsStr *string
		err := rows.Scan(
			&o.ID, &o.Title, &o.Description, &o.DiscountType, &o.DiscountValue,
			&o.AppliesTo, &appliesToIdsStr, &o.MinOrderAmt, &o.UsageLimit,
			&o.UsageCount, &o.IsActive, &o.StartsAt, &o.EndsAt,
			&o.CreatedAt, &o.UpdatedAt,
		)
		if err != nil {
			continue
		}
		if appliesToIdsStr != nil && *appliesToIdsStr != "" {
			o.AppliesToIds = strings.Split(*appliesToIdsStr, ",")
		}
		offers = append(offers, o)
	}

	response := gin.H{
		"items":       offers,
		"total":       len(offers),
		"page":        1,
		"limit":       50,
		"next_cursor": nil,
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) CreateOffer(c *gin.Context) {
	var req CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse time strings
	var startsAt, endsAt time.Time
	var err error

	if req.StartsAt != nil {
		startsAt, err = time.Parse(time.RFC3339, *req.StartsAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid starts_at format, expected RFC3339"})
			return
		}
	}

	if req.EndsAt != nil {
		endsAt, err = time.Parse(time.RFC3339, *req.EndsAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ends_at format, expected RFC3339"})
			return
		}
	}

	id := uuid.New()
	now := time.Now().UTC()

	var appliesToIdsStr *string
	if len(req.AppliesToIds) > 0 {
		joined := strings.Join(req.AppliesToIds, ",")
		appliesToIdsStr = &joined
	}

	_, err = h.DB.Exec(c, `
		INSERT INTO offers (id, title, description, discount_type, discount_value, applies_to, applies_to_ids,
						   min_order_amount, usage_limit, usage_count, is_active, starts_at, ends_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $14)
	`, id, req.Title, req.Description, req.DiscountType, req.DiscountValue,
		req.AppliesTo, appliesToIdsStr, req.MinOrderAmt, req.UsageLimit, 0,
		req.IsActive, startsAt, endsAt, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	offer := Offer{
		ID:            id,
		Title:         req.Title,
		Description:   req.Description,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		AppliesTo:     req.AppliesTo,
		AppliesToIds:  req.AppliesToIds,
		MinOrderAmt:   req.MinOrderAmt,
		UsageLimit:    req.UsageLimit,
		UsageCount:    0,
		IsActive:      req.IsActive,
		StartsAt:      startsAt,
		EndsAt:        endsAt,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	c.JSON(http.StatusCreated, offer)
}

func (h *Handler) GetOffer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var o Offer
	var appliesToIdsStr string
	err = h.DB.QueryRow(c, `
		SELECT id, title, description, discount_type, discount_value, applies_to, applies_to_ids,
			   min_order_amount, usage_limit, usage_count, is_active, starts_at, ends_at, created_at, updated_at
		FROM offers
		WHERE id = $1
	`, id).Scan(
		&o.ID, &o.Title, &o.Description, &o.DiscountType, &o.DiscountValue,
		&o.AppliesTo, &appliesToIdsStr, &o.MinOrderAmt, &o.UsageLimit,
		&o.UsageCount, &o.IsActive, &o.StartsAt, &o.EndsAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "offer not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if appliesToIdsStr != "" {
		o.AppliesToIds = strings.Split(appliesToIdsStr, ",")
	}

	c.JSON(http.StatusOK, o)
}

func (h *Handler) UpdateOffer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req UpdateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sets := []string{}
	args := []any{1, id}
	argIdx := 2

	if req.Title != nil {
		sets = append(sets, "title = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Title)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, "description = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Description)
		argIdx++
	}
	if req.DiscountType != nil {
		sets = append(sets, "discount_type = $"+strconv.Itoa(argIdx))
		args = append(args, *req.DiscountType)
		argIdx++
	}
	if req.DiscountValue != nil {
		sets = append(sets, "discount_value = $"+strconv.Itoa(argIdx))
		args = append(args, *req.DiscountValue)
		argIdx++
	}
	if req.AppliesTo != nil {
		sets = append(sets, "applies_to = $"+strconv.Itoa(argIdx))
		args = append(args, *req.AppliesTo)
		argIdx++
	}
	if req.AppliesToIds != nil {
		sets = append(sets, "applies_to_ids = $"+strconv.Itoa(argIdx))
		if len(*req.AppliesToIds) > 0 {
			args = append(args, strings.Join(*req.AppliesToIds, ","))
		} else {
			args = append(args, "")
		}
		argIdx++
	}
	if req.MinOrderAmt != nil {
		sets = append(sets, "min_order_amount = $"+strconv.Itoa(argIdx))
		args = append(args, *req.MinOrderAmt)
		argIdx++
	}
	if req.UsageLimit != nil {
		sets = append(sets, "usage_limit = $"+strconv.Itoa(argIdx))
		args = append(args, *req.UsageLimit)
		argIdx++
	}
	if req.IsActive != nil {
		sets = append(sets, "is_active = $"+strconv.Itoa(argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.StartsAt != nil {
		sets = append(sets, "starts_at = $"+strconv.Itoa(argIdx))
		args = append(args, *req.StartsAt)
		argIdx++
	}
	if req.EndsAt != nil {
		sets = append(sets, "ends_at = $"+strconv.Itoa(argIdx))
		args = append(args, *req.EndsAt)
		argIdx++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	sets = append(sets, "updated_at = NOW()")

	query := "UPDATE offers SET " + joinString(sets, ", ") + " WHERE id = $" + strconv.Itoa(argIdx)
	args = append(args, id)

	_, err = h.DB.Exec(c, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var o Offer
	var appliesToIdsStr string
	err = h.DB.QueryRow(c, `
		SELECT id, title, description, discount_type, discount_value, applies_to, applies_to_ids,
			   min_order_amount, usage_limit, usage_count, is_active, starts_at, ends_at, created_at, updated_at
		FROM offers
		WHERE id = $1
	`, id).Scan(
		&o.ID, &o.Title, &o.Description, &o.DiscountType, &o.DiscountValue,
		&o.AppliesTo, &appliesToIdsStr, &o.MinOrderAmt, &o.UsageLimit,
		&o.UsageCount, &o.IsActive, &o.StartsAt, &o.EndsAt, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if appliesToIdsStr != "" {
		o.AppliesToIds = strings.Split(appliesToIdsStr, ",")
	}

	c.JSON(http.StatusOK, o)
}

func (h *Handler) DeleteOffer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	_, err = h.DB.Exec(c, "DELETE FROM offers WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
