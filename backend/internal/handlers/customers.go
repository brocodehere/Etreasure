package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Customer struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Addresses []Address `json:"addresses"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Address struct {
	ID         *string `json:"id,omitempty"`
	CustomerID *string `json:"customer_id,omitempty"`
	Type       string  `json:"type"` // shipping | billing
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Company    *string `json:"company,omitempty"`
	Address1   string  `json:"address1"`
	Address2   *string `json:"address2,omitempty"`
	City       string  `json:"city"`
	Province   *string `json:"province,omitempty"`
	Country    string  `json:"country"`
	PostalCode string  `json:"postal_code"`
	Phone      *string `json:"phone,omitempty"`
}

type CreateCustomerRequest struct {
	Email     string    `json:"email" binding:"required,email"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Addresses []Address `json:"addresses"`
}

type UpdateCustomerRequest struct {
	Email     *string    `json:"email,omitempty"`
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	Addresses *[]Address `json:"addresses,omitempty"`
}

func (h *Handler) ListCustomers(c *gin.Context) {
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
		SELECT id, email, first_name, last_name, phone, created_at, updated_at
		FROM customers
		WHERE ($1::timestamp IS NULL OR updated_at < $1::timestamp)
		ORDER BY updated_at DESC
		LIMIT $2
	`, cursorTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var customers []Customer
	for rows.Next() {
		var cust Customer
		err := rows.Scan(&cust.ID, &cust.Email, &cust.FirstName, &cust.LastName, &cust.Phone, &cust.CreatedAt, &cust.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cust.Addresses, err = h.loadCustomerAddresses(c, cust.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		customers = append(customers, cust)
	}

	var nextCursor *string
	if len(customers) == limit {
		next := customers[len(customers)-1].UpdatedAt.Format(time.RFC3339Nano)
		nextCursor = &next
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        customers,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) CreateCustomer(c *gin.Context) {
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()

	_, err := h.DB.Exec(c, `
		INSERT INTO customers (id, email, first_name, last_name, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, id, req.Email, req.FirstName, req.LastName, req.Phone, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert addresses if provided
	for _, addr := range req.Addresses {
		addrID := uuid.New().String()
		_, err = h.DB.Exec(c, `
			INSERT INTO addresses (id, customer_id, type, first_name, last_name, company,
								  address1, address2, city, province, country, postal_code, phone)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`, addrID, id, addr.Type, addr.FirstName, addr.LastName, addr.Company,
			addr.Address1, addr.Address2, addr.City, addr.Province, addr.Country, addr.PostalCode, addr.Phone)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	customer := Customer{
		ID:        id,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Addresses: req.Addresses,
		CreatedAt: now,
		UpdatedAt: now,
	}

	c.JSON(http.StatusCreated, customer)
}

func (h *Handler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	var cust Customer
	err := h.DB.QueryRow(c, `
		SELECT id, email, first_name, last_name, phone, created_at, updated_at
		FROM customers
		WHERE id = $1
	`, id).Scan(&cust.ID, &cust.Email, &cust.FirstName, &cust.LastName, &cust.Phone, &cust.CreatedAt, &cust.UpdatedAt)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cust.Addresses, err = h.loadCustomerAddresses(c, cust.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cust)
}

func (h *Handler) UpdateCustomer(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sets := []string{}
	args := []any{1, id}
	argIdx := 2

	if req.Email != nil {
		sets = append(sets, "email = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.FirstName != nil {
		sets = append(sets, "first_name = $"+strconv.Itoa(argIdx))
		args = append(args, *req.FirstName)
		argIdx++
	}
	if req.LastName != nil {
		sets = append(sets, "last_name = $"+strconv.Itoa(argIdx))
		args = append(args, *req.LastName)
		argIdx++
	}
	if req.Phone != nil {
		sets = append(sets, "phone = $"+strconv.Itoa(argIdx))
		args = append(args, *req.Phone)
		argIdx++
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = NOW()")
		query := "UPDATE customers SET " + joinString(sets, ", ") + " WHERE id = $" + strconv.Itoa(argIdx)
		args = append(args, id)
		_, err := h.DB.Exec(c, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Update addresses if provided
	if req.Addresses != nil {
		// Delete existing addresses
		_, err := h.DB.Exec(c, "DELETE FROM addresses WHERE customer_id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Insert new addresses
		for _, addr := range *req.Addresses {
			addrID := uuid.New().String()
			_, err = h.DB.Exec(c, `
				INSERT INTO addresses (id, customer_id, type, first_name, last_name, company,
									  address1, address2, city, province, country, postal_code, phone)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
			`, addrID, id, addr.Type, addr.FirstName, addr.LastName, addr.Company,
				addr.Address1, addr.Address2, addr.City, addr.Province, addr.Country, addr.PostalCode, addr.Phone)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Return updated customer
	var cust Customer
	err := h.DB.QueryRow(c, `
		SELECT id, email, first_name, last_name, phone, created_at, updated_at
		FROM customers
		WHERE id = $1
	`, id).Scan(&cust.ID, &cust.Email, &cust.FirstName, &cust.LastName, &cust.Phone, &cust.CreatedAt, &cust.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	cust.Addresses, _ = h.loadCustomerAddresses(c, cust.ID)
	c.JSON(http.StatusOK, cust)
}

func (h *Handler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	_, err := h.DB.Exec(c, "DELETE FROM customers WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) loadCustomerAddresses(c *gin.Context, customerID string) ([]Address, error) {
	rows, err := h.DB.Query(c, `
		SELECT id, customer_id, type, first_name, last_name, company,
			   address1, address2, city, province, country, postal_code, phone
		FROM addresses
		WHERE customer_id = $1
	`, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addrs []Address
	for rows.Next() {
		var addr Address
		err := rows.Scan(&addr.ID, &addr.CustomerID, &addr.Type, &addr.FirstName, &addr.LastName,
			&addr.Company, &addr.Address1, &addr.Address2, &addr.City, &addr.Province,
			&addr.Country, &addr.PostalCode, &addr.Phone)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

// GetCustomerProfile - get current logged-in user's profile
func (h *Handler) GetCustomerProfile(c *gin.Context) {
	// Get user ID from auth context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var cust Customer
	err := h.DB.QueryRow(c, `
		SELECT id, email, first_name, last_name, phone, created_at, updated_at
		FROM customers
		WHERE id = $1
	`, userID).Scan(&cust.ID, &cust.Email, &cust.FirstName, &cust.LastName, &cust.Phone, &cust.CreatedAt, &cust.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "profile not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	addrs, err := h.loadCustomerAddresses(c, cust.ID)
	if err == nil {
		cust.Addresses = addrs
	}

	c.JSON(http.StatusOK, gin.H{"data": cust})
}

// ListCustomerAddresses - list addresses for current user
func (h *Handler) ListCustomerAddresses(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	addrs, err := h.loadCustomerAddresses(c, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": addrs})
}

// AddCustomerAddress - add a new address for current user
func (h *Handler) AddCustomerAddress(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req Address
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	addrID := uuid.New().String()
	userIDStr := userID.(string)
	_, err := h.DB.Exec(c, `
		INSERT INTO addresses (id, customer_id, type, first_name, last_name, company, address1, address2, city, province, country, postal_code, phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, addrID, userIDStr, req.Type, req.FirstName, req.LastName, req.Company, req.Address1, req.Address2,
		req.City, req.Province, req.Country, req.PostalCode, req.Phone)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	req.ID = &addrID
	req.CustomerID = &userIDStr
	c.JSON(http.StatusCreated, gin.H{"data": req})
}
