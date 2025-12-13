package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type Setting struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Type        string    `json:"type"` // string | number | boolean | json
	Description *string   `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateSettingRequest struct {
	Key         string  `json:"key" binding:"required"`
	Value       string  `json:"value" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=string number boolean json"`
	Description *string `json:"description,omitempty"`
}

type UpdateSettingRequest struct {
	Value       *string `json:"value,omitempty"`
	Type        *string `json:"type,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (h *Handler) ListSettings(c *gin.Context) {
	rows, err := h.DB.Query(c, `
		SELECT key, value, type, description, updated_at
		FROM settings
		ORDER BY key ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var settings []Setting
	for rows.Next() {
		var s Setting
		err := rows.Scan(&s.Key, &s.Value, &s.Type, &s.Description, &s.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		settings = append(settings, s)
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *Handler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	var s Setting
	err := h.DB.QueryRow(c, `
		SELECT key, value, type, description, updated_at
		FROM settings
		WHERE key = $1
	`, key).Scan(&s.Key, &s.Value, &s.Type, &s.Description, &s.UpdatedAt)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) CreateSetting(c *gin.Context) {
	var req CreateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now().UTC()

	_, err := h.DB.Exec(c, `
		INSERT INTO settings (key, value, type, description, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, req.Key, req.Value, req.Type, req.Description, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	setting := Setting{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Description: req.Description,
		UpdatedAt:   now,
	}

	c.JSON(http.StatusCreated, setting)
}

func (h *Handler) UpdateSetting(c *gin.Context) {
	key := c.Param("key")
	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sets := []string{}
	args := []any{1, key}
	argIdx := 2

	if req.Value != nil {
		sets = append(sets, "value = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Value)
		argIdx++
	}
	if req.Type != nil {
		sets = append(sets, "type = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Type)
		argIdx++
	}
	if req.Description != nil {
		sets = append(sets, "description = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Description)
		argIdx++
	}

	if len(sets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	sets = append(sets, "updated_at = NOW()")
	query := "UPDATE settings SET " + joinString(sets, ", ") + " WHERE key = $" + strconv.Itoa(argIdx)
	args = append(args, key)

	_, err := h.DB.Exec(c, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return updated setting
	var s Setting
	err = h.DB.QueryRow(c, `
		SELECT key, value, type, description, updated_at
		FROM settings
		WHERE key = $1
	`, key).Scan(&s.Key, &s.Value, &s.Type, &s.Description, &s.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (h *Handler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	_, err := h.DB.Exec(c, "DELETE FROM settings WHERE key = $1", key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) GetPublicSettings(c *gin.Context) {
	// Return only safe public settings for storefront
	rows, err := h.DB.Query(c, `
		SELECT key, value, type
		FROM settings
		WHERE key LIKE 'public.%' OR key IN ('store_name','store_description','currency','contact_email')
		ORDER BY key ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var key, val, typ string
		if err := rows.Scan(&key, &val, &typ); err != nil {
			continue
		}
		var parsed interface{}
		switch typ {
		case "number":
			var f float64
			if err := json.Unmarshal([]byte(val), &f); err == nil {
				parsed = f
			} else {
				parsed = val
			}
		case "boolean":
			var b bool
			if err := json.Unmarshal([]byte(val), &b); err == nil {
				parsed = b
			} else {
				parsed = val
			}
		case "json":
			var j interface{}
			if err := json.Unmarshal([]byte(val), &j); err == nil {
				parsed = j
			} else {
				parsed = val
			}
		default:
			parsed = val
		}
		result[key] = parsed
	}
	c.JSON(http.StatusOK, result)
}
