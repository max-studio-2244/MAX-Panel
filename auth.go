package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Role              string    `json:"role"`
	IsAdmin           bool      `json:"is_admin"`
	TwoFactorEnabled  bool      `json:"two_factor_enabled"`
	TwoFactorSecret   string    `json:"two_factor_secret,omitempty"`
	APIKey            string    `json:"api_key,omitempty"`
	LastLogin         time.Time `json:"last_login"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	IsAdmin   bool   `json:"is_admin"`
}

func login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user User
	var hashedPassword string
	err := db.QueryRow(`
		SELECT id, username, email, first_name, last_name, role, is_admin, password
		FROM users WHERE username = ? OR email = ?
	`, req.Username, req.Username).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, 
		&user.LastName, &user.Role, &user.IsAdmin, &hashedPassword,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
	}

	// Update last login
	_, err = db.Exec("UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Generate session token (simplified)
	token := generateToken()
	
	// Log activity
	logActivity(user.ID, "user.login", "User logged in", c.IP(), c.Get("User-Agent"))

	return c.JSON(fiber.Map{
		"user":  user,
		"token": token,
	})
}

func register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	// Generate user ID and API key
	userID := generateID()
	apiKey := generateToken()

	// Insert user
	_, err = db.Exec(`
		INSERT INTO users (id, username, email, password, first_name, last_name, is_admin, api_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, userID, req.Username, req.Email, string(hashedPassword), req.FirstName, req.LastName, req.IsAdmin, apiKey)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Log activity
	logActivity(userID, "user.register", "User registered", c.IP(), c.Get("User-Agent"))

	return c.JSON(fiber.Map{
		"id":      userID,
		"message": "User created successfully",
	})
}

func logout(c *fiber.Ctx) error {
	// In a real implementation, you would invalidate the session token
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

func getMe(c *fiber.Ctx) error {
	// In a real implementation, you would get user from session token
	// For now, return a mock admin user
	user := User{
		ID:        "admin",
		Username:  "admin",
		Email:     "admin@maxpanel.com",
		FirstName: "Admin",
		LastName:  "User",
		Role:      "admin",
		IsAdmin:   true,
	}

	return c.JSON(user)
}

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func forgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	
	// Check if user exists
	var userID string
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal if email exists or not
			return c.JSON(fiber.Map{"message": "If the email exists, a reset link has been sent"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// Generate reset token
	resetToken := generateToken()
	
	// Store reset token (in production, use a separate table with expiration)
	_, err = db.Exec("UPDATE users SET api_key = ? WHERE id = ?", resetToken, userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// In production, send email with reset link
	// For now, just log the token
	println("Password reset token for", req.Email, ":", resetToken)
	
	logActivity(userID, "password.reset_requested", "Password reset requested", c.IP(), c.Get("User-Agent"))
	
	return c.JSON(fiber.Map{"message": "If the email exists, a reset link has been sent"})
}

func resetPassword(c *fiber.Ctx) error {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	
	// Find user by reset token
	var userID string
	err := db.QueryRow("SELECT id FROM users WHERE api_key = ?", req.Token).Scan(&userID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired reset token"})
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
	}
	
	// Update password and clear reset token
	_, err = db.Exec("UPDATE users SET password = ?, api_key = ? WHERE id = ?", string(hashedPassword), generateToken(), userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	logActivity(userID, "password.reset_completed", "Password reset completed", c.IP(), c.Get("User-Agent"))
	
	return c.JSON(fiber.Map{"message": "Password reset successfully"})
}

func verifyTwoFA(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	
	// In production, verify TOTP code against user's secret
	// For now, accept any 6-digit code
	if len(req.Code) != 6 {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid 2FA code"})
	}
	
	// Generate session token
	token := generateToken()
	
	logActivity(req.UserID, "auth.2fa_verified", "2FA verification successful", c.IP(), c.Get("User-Agent"))
	
	return c.JSON(fiber.Map{"token": token})
}

func enableTwoFA(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "2FA setup coming soon"})
}

func disableTwoFA(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "2FA disable coming soon"})
}

func logActivity(userID, action, description, ipAddress, userAgent string) {
	activityID := generateID()
	_, err := db.Exec(`
		INSERT INTO activity_logs (id, user_id, action, description, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?)
	`, activityID, userID, action, description, ipAddress, userAgent)
	
	if err != nil {
		// Log error but don't fail the request
		println("Failed to log activity:", err.Error())
	}
}