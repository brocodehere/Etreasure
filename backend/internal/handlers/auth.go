package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/etreasure/backend/internal/auth"
	"github.com/etreasure/backend/internal/config"
	"github.com/etreasure/backend/internal/email"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	DB    *pgxpool.Pool
	Cfg   config.Config
	Rd    *redis.Client
	Email *email.EmailService
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type signupRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FullName   string `json:"fullName"`
	RememberMe bool   `json:"rememberMe"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type verifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type resetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

type tokenResponse struct {
	AccessToken  string        `json:"accessToken"`
	RefreshToken string        `json:"refreshToken"`
	User         authUserModel `json:"user"`
}

type authUserModel struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Roles     []string  `json:"roles"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Basic rate limiting: 5 attempts per minute per IP
	ip := c.ClientIP()
	if h.Rd != nil {
		ctx := context.Background()
		key := fmt.Sprintf("login:ip:%s", ip)
		count, err := h.Rd.Incr(ctx, key).Result()
		if err == nil {
			if count == 1 {
				_ = h.Rd.Expire(ctx, key, time.Minute).Err()
			}
			if count > 5 {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, please try again later"})
				return
			}
		}
	}

	ctx := context.Background()
	var (
		id           int
		email        string
		fullName     sql.NullString
		passwordHash string
	)
	err := h.DB.QueryRow(ctx, `SELECT id, email, full_name, password_hash FROM users WHERE email = $1 AND is_active = TRUE`, req.Email).
		Scan(&id, &email, &fullName, &passwordHash)
	if err != nil {
		// Log DB lookup error for debugging (may be no rows or DB issue)
		log.Printf("auth.Login: user lookup failed for email=%s: %v", req.Email, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		// Password mismatch
		log.Printf("auth.Login: password mismatch for user id=%d email=%s", id, req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	rows, err := h.DB.Query(ctx, `SELECT r.name FROM roles r JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load roles"})
		return
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err == nil {
			roles = append(roles, r)
		}
	}

	access, err := auth.GenerateAccessToken(h.Cfg.JWTSecret, id, roles, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}
	refreshTTL := 30 * 24 * time.Hour
	refresh, err := auth.GenerateAccessToken(h.Cfg.RefreshSecret, id, roles, refreshTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	// Persist hashed refresh token for potential revocation/rotation
	if err := h.storeRefreshToken(ctx, id, refresh, false, ip, c.GetHeader("User-Agent"), time.Now().Add(refreshTTL)); err != nil {
		// Log but do not fail login if persistence issues occur
	}

	c.JSON(http.StatusOK, tokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User: authUserModel{
			ID:    id,
			Email: email,
			Name:  fullName.String,
			Roles: roles,
		},
	})
}

// Signup registers a new user and returns tokens
func (h *AuthHandler) Signup(c *gin.Context) {
	var req signupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := context.Background()
	// Check if user already exists
	var exists bool
	err := h.DB.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Insert user
	var userID int
	err = h.DB.QueryRow(ctx, `
		INSERT INTO users (email, password_hash, full_name, is_active, created_at)
		VALUES ($1, $2, $3, TRUE, NOW())
		RETURNING id
	`, req.Email, string(hashedPassword), req.FullName).Scan(&userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	// Default role: customer (if role exists)
	var roles []string
	if _, err := h.DB.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id)
		SELECT $1, id FROM roles WHERE name = 'customer'
		ON CONFLICT DO NOTHING
	`, userID); err == nil {
		roles = append(roles, "customer")
	}

	// Issue tokens
	access, err := auth.GenerateAccessToken(h.Cfg.JWTSecret, userID, roles, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}
	refreshTTL := 30 * 24 * time.Hour
	if !req.RememberMe {
		refreshTTL = 7 * 24 * time.Hour
	}
	refresh, err := auth.GenerateAccessToken(h.Cfg.RefreshSecret, userID, roles, refreshTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	ip := c.ClientIP()
	if err := h.storeRefreshToken(ctx, userID, refresh, req.RememberMe, ip, c.GetHeader("User-Agent"), time.Now().Add(refreshTTL)); err != nil {
	}

	c.JSON(http.StatusCreated, tokenResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		User: authUserModel{
			ID:    userID,
			Email: req.Email,
			Name:  req.FullName,
			Roles: roles,
		},
	})
}

// storeRefreshToken hashes a refresh token and stores it in the refresh_tokens table
func (h *AuthHandler) storeRefreshToken(ctx context.Context, userID int, token string, rememberMe bool, ip, userAgent string, expiresAt time.Time) error {
	hashBytes := sha256.Sum256([]byte(token))
	th := hex.EncodeToString(hashBytes[:])
	_, err := h.DB.Exec(ctx, `
		INSERT INTO refresh_tokens (user_id, token_hash, user_agent, ip_address, remember_me, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, th, userAgent, ip, rememberMe, expiresAt)
	return err
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var body struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	claims, err := auth.ParseToken(h.Cfg.RefreshSecret, body.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
		return
	}

	access, err := auth.GenerateAccessToken(h.Cfg.JWTSecret, claims.UserID, claims.Roles, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"accessToken": access})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Stateless JWT: on client we simply drop tokens. In production you may add token blacklist / rotation via DB.
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *AuthHandler) Me(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no authorization header"})
		return
	}

	// Extract token from "Bearer <token>"
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
		return
	}

	// Parse token to get user info
	claims, err := auth.ParseToken(h.Cfg.JWTSecret, tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	ctx := context.Background()
	var (
		id        int
		email     string
		fullName  sql.NullString
		createdAt time.Time
	)
	err = h.DB.QueryRow(ctx, `SELECT id, email, full_name, created_at FROM users WHERE id = $1`, claims.UserID).
		Scan(&id, &email, &fullName, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, authUserModel{
		ID:        id,
		Email:     email,
		Name:      fullName.String,
		Roles:     claims.Roles,
		CreatedAt: createdAt,
	})
}

// ForgotPassword - sends OTP to user's email if it exists in database
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := context.Background()
	var exists bool
	err := h.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND is_active = TRUE)", req.Email).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	if !exists {
		// Don't reveal that email doesn't exist for security
		c.JSON(http.StatusOK, gin.H{"message": "If email exists, OTP will be sent"})
		return
	}

	// Generate 6-digit OTP
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Store OTP in Redis with 10-minute expiry
	redisKey := fmt.Sprintf("otp:%s", req.Email)
	err = h.Rd.Set(ctx, redisKey, otp, 10*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store OTP"})
		return
	}

	// Send OTP email
	err = h.Email.SendOTPEmail(req.Email, otp)
	if err != nil {
		// Log error but don't fail the request
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent to your email"})
}

// VerifyOTP - verifies the OTP from Redis
func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req verifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := context.Background()
	redisKey := fmt.Sprintf("otp:%s", req.Email)

	// Get OTP from Redis
	storedOTP, err := h.Rd.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP not found or expired"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify OTP"})
		return
	}

	// Verify OTP
	if storedOTP != req.OTP {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid OTP"})
		return
	}

	// Mark OTP as verified by setting a verification flag
	verifyKey := fmt.Sprintf("verified:%s", req.Email)
	err = h.Rd.Set(ctx, verifyKey, "true", 5*time.Minute).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark verification"})
		return
	}

	// Delete the OTP after successful verification
	h.Rd.Del(ctx, redisKey)

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}

// ResetPassword - resets password after OTP verification
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	ctx := context.Background()

	// Check if OTP was verified
	verifyKey := fmt.Sprintf("verified:%s", req.Email)
	verified, err := h.Rd.Get(ctx, verifyKey).Result()
	if err == redis.Nil || verified != "true" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP not verified or verification expired"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	// Update password in database
	_, err = h.DB.Exec(ctx, "UPDATE users SET password_hash = $1, updated_at = NOW() WHERE email = $2 AND is_active = TRUE",
		string(hashedPassword), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update password"})
		return
	}

	// Clean up Redis
	h.Rd.Del(ctx, verifyKey)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
