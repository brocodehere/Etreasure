package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Roles     []string  `json:"roles"`
}

type UserCustomer struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	FirstName  string    `json:"first_name,omitempty"`
	LastName   string    `json:"last_name,omitempty"`
	FullName   string    `json:"full_name"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	OrderCount int       `json:"order_count"`
}

type CreateUserRequest struct {
	Email     string   `json:"email" binding:"required,email"`
	FirstName *string  `json:"first_name,omitempty"`
	LastName  *string  `json:"last_name,omitempty"`
	Password  string   `json:"password" binding:"required,min=8"`
	IsActive  *bool    `json:"is_active,omitempty"`
	Roles     []string `json:"roles"`
}

type UpdateUserRequest struct {
	Email     *string   `json:"email,omitempty"`
	FirstName *string   `json:"first_name,omitempty"`
	LastName  *string   `json:"last_name,omitempty"`
	Password  *string   `json:"password,omitempty"`
	IsActive  *bool     `json:"is_active,omitempty"`
	Roles     *[]string `json:"roles,omitempty"`
}

type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *Handler) ListUsers(c *gin.Context) {
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
		SELECT u.id, u.email, u.first_name, u.last_name, u.is_active, u.created_at, u.updated_at
		FROM users u
		WHERE ($1::timestamp IS NULL OR u.updated_at < $1::timestamp)
		ORDER BY u.updated_at DESC
		LIMIT $2
	`, cursorTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		u.Roles, err = h.loadUserRoles(c, u.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		users = append(users, u)
	}

	var nextCursor *string
	if len(users) == limit {
		next := users[len(users)-1].UpdatedAt.Format(time.RFC3339Nano)
		nextCursor = &next
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        users,
		"next_cursor": nextCursor,
	})
}

func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	_, err = h.DB.Exec(c, `
		INSERT INTO users (id, email, first_name, last_name, password_hash, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, id, req.Email, req.FirstName, req.LastName, string(hashedPassword), isActive, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Assign roles
	if len(req.Roles) > 0 {
		for _, roleName := range req.Roles {
			_, err = h.DB.Exec(c, `
				INSERT INTO user_roles (user_id, role_id)
				SELECT $1, id FROM roles WHERE name = $2
				ON CONFLICT DO NOTHING
			`, id, roleName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	user := User{
		ID:        id,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  isActive,
		CreatedAt: now,
		UpdatedAt: now,
		Roles:     req.Roles,
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetUser(c *gin.Context) {
	id := c.Param("id")
	var u User
	err := h.DB.QueryRow(c, `
		SELECT id, email, first_name, last_name, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u.Roles, err = h.loadUserRoles(c, u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req UpdateUserRequest
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
	if req.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		sets = append(sets, "password_hash = $"+strconv.Itoa(argIdx))
		args = append(args, string(hashedPassword))
		argIdx++
	}
	if req.IsActive != nil {
		sets = append(sets, "is_active = $"+strconv.Itoa(argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(sets) > 0 {
		sets = append(sets, "updated_at = NOW()")
		query := "UPDATE users SET " + joinString(sets, ", ") + " WHERE id = $" + strconv.Itoa(argIdx)
		args = append(args, id)
		_, err := h.DB.Exec(c, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Update roles if provided
	if req.Roles != nil {
		// Delete existing roles
		_, err := h.DB.Exec(c, "DELETE FROM user_roles WHERE user_id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Insert new roles
		for _, roleName := range *req.Roles {
			_, err = h.DB.Exec(c, `
				INSERT INTO user_roles (user_id, role_id)
				SELECT $1, id FROM roles WHERE name = $2
				ON CONFLICT DO NOTHING
			`, id, roleName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	// Return updated user
	var u User
	err := h.DB.QueryRow(c, `
		SELECT id, email, first_name, last_name, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	u.Roles, _ = h.loadUserRoles(c, id)
	c.JSON(http.StatusOK, u)
}

func (h *Handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	// Prevent deleting self
	currentUserID := c.GetString("userID")
	if currentUserID == id {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot delete yourself"})
		return
	}
	_, err := h.DB.Exec(c, "DELETE FROM users WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) loadUserRoles(c *gin.Context, userID string) ([]string, error) {
	rows, err := h.DB.Query(c, `
		SELECT r.name
		FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}
	return roles, nil
}

func (h *Handler) ListRoles(c *gin.Context) {
	rows, err := h.DB.Query(c, `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		ORDER BY name ASC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		roles = append(roles, r)
	}
	c.JSON(http.StatusOK, gin.H{"data": roles})
}

// ListUserCustomers - fetch customers from users table excluding admin roles
func (h *Handler) ListUserCustomers(c *gin.Context) {
	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	// Query all users without filtering
	query := `
		SELECT u.id, u.email, '', '', u.full_name, u.is_active, u.created_at, NOW(), 0
		FROM users u
		ORDER BY u.created_at DESC
		LIMIT $1
	`

	rows, err := h.DB.Query(c, query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var customers []UserCustomer
	for rows.Next() {
		var cust UserCustomer
		err := rows.Scan(&cust.ID, &cust.Email, &cust.FirstName, &cust.LastName, &cust.FullName,
			&cust.IsActive, &cust.CreatedAt, &cust.UpdatedAt, &cust.OrderCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Add all users without filtering
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

// GetCustomerOrders - get all orders for a specific customer
func (h *Handler) GetCustomerOrders(c *gin.Context) {
	customerID := c.Param("id")

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
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
		SELECT o.id, o.order_number, o.status, o.total_cents, o.currency, 
		       o.placed_at, o.fulfilled_at, o.cancelled_at, o.refund_cents,
		       COUNT(oi.id) as item_count
		FROM orders o
		LEFT JOIN order_items oi ON o.id = oi.order_id
		WHERE o.user_id = $1
		AND ($2::timestamp IS NULL OR o.updated_at < $2::timestamp)
		GROUP BY o.id, o.order_number, o.status, o.total_cents, o.currency, 
		         o.placed_at, o.fulfilled_at, o.cancelled_at, o.refund_cents
		ORDER BY o.updated_at DESC
		LIMIT $3
	`, customerID, cursorTime, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type CustomerOrder struct {
		ID          string     `json:"id"`
		OrderNumber string     `json:"order_number"`
		Status      string     `json:"status"`
		TotalCents  int        `json:"total_cents"`
		Currency    string     `json:"currency"`
		PlacedAt    time.Time  `json:"placed_at"`
		FulfilledAt *time.Time `json:"fulfilled_at,omitempty"`
		CancelledAt *time.Time `json:"cancelled_at,omitempty"`
		RefundCents int        `json:"refund_cents"`
		ItemCount   int        `json:"item_count"`
	}

	var orders []CustomerOrder
	for rows.Next() {
		var order CustomerOrder
		err := rows.Scan(&order.ID, &order.OrderNumber, &order.Status, &order.TotalCents,
			&order.Currency, &order.PlacedAt, &order.FulfilledAt, &order.CancelledAt,
			&order.RefundCents, &order.ItemCount)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		orders = append(orders, order)
	}

	var nextCursor *string
	if len(orders) == limit {
		next := orders[len(orders)-1].PlacedAt.Format(time.RFC3339Nano)
		nextCursor = &next
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        orders,
		"next_cursor": nextCursor,
	})
}
