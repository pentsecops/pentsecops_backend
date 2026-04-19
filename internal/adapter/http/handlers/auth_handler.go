package handlers

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authUseCase     *usecases.AuthUseCase
	activityUseCase *usecases.ActivityUseCase
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authUseCase *usecases.AuthUseCase, activityUseCase *usecases.ActivityUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase:     authUseCase,
		activityUseCase: activityUseCase,
	}
}

// response encapsulates the API response
type response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LoginUser handles user login endpoint POST /api/v1/login
func (ah *AuthHandler) LoginUser(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/login - User login attempt")

	var req usecases.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] LoginUser - Invalid request format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		log.Printf("[ERROR] LoginUser - Missing credentials for email: %s", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Email and password are required",
			Error:   "missing credentials",
		})
	}

	// Attempt login
	loginResp, err := ah.authUseCase.LoginUser(&req)
	if err != nil {
		log.Printf("[ERROR] LoginUser - Authentication failed for email: %s, error: %v", req.Email, err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Login failed",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] LoginUser - User logged in successfully, email: %s, user_id: %s", req.Email, loginResp.UserID)

	// Log activity asynchronously
	go ah.activityUseCase.LogActivity(
		loginResp.UserID,
		domain.ActivityTypeUserLogin,
		fmt.Sprintf("User %s logged in", req.Email),
		map[string]interface{}{"email": req.Email},
		c.IP(),
		c.Get("User-Agent"),
	)

	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Login successful",
		Data:    loginResp,
	})
}

// LoginAdmin handles admin login endpoint POST /api/v1/admin/login
func (ah *AuthHandler) LoginAdmin(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/admin/login - Admin login attempt")

	var req usecases.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] LoginAdmin - Invalid request format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		log.Printf("[ERROR] LoginAdmin - Missing credentials for email: %s", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Email and password are required",
			Error:   "missing credentials",
		})
	}

	// Attempt admin login
	loginResp, err := ah.authUseCase.LoginAdmin(&req)
	if err != nil {
		log.Printf("[ERROR] LoginAdmin - Authentication failed for email: %s, error: %v", req.Email, err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Login failed",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] LoginAdmin - Admin logged in successfully, email: %s, admin_id: %s", req.Email, loginResp.AdminID)

	// Log activity asynchronously
	go ah.activityUseCase.LogActivity(
		loginResp.AdminID,
		domain.ActivityTypeUserLogin,
		fmt.Sprintf("Admin %s logged in", req.Email),
		map[string]interface{}{"email": req.Email, "role": "admin"},
		c.IP(),
		c.Get("User-Agent"),
	)

	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Admin login successful",
		Data:    loginResp,
	})
}

// RegisterAdmin handles admin registration endpoint POST /api/v1/admin/register
func (ah *AuthHandler) RegisterAdmin(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/admin/register - Admin registration attempt")

	var req usecases.AdminRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] RegisterAdmin - Invalid request format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Validate input
	if req.Email == "" || req.Password == "" {
		log.Printf("[ERROR] RegisterAdmin - Missing credentials for email: %s", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Email and password are required",
			Error:   "missing credentials",
		})
	}

	// Register admin
	registerResp, err := ah.authUseCase.RegisterAdmin(&req)
	if err != nil {
		log.Printf("[ERROR] RegisterAdmin - Registration failed for email: %s, error: %v", req.Email, err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Registration failed",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] RegisterAdmin - Admin registered successfully, email: %s, admin_id: %s", req.Email, registerResp.AdminID)
	return c.Status(fiber.StatusCreated).JSON(response{
		Success: true,
		Message: "Admin registered successfully",
		Data:    registerResp,
	})
}

// RegisterUser handles user registration by admin - POST /api/v1/admin/register/user
func (ah *AuthHandler) RegisterUser(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/admin/register/user - User registration by admin")

	// Verify admin authorization via token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] RegisterUser - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	// Validate token (must be admin)
	claims, err := ah.authUseCase.ValidateToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] RegisterUser - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	if claims.Role != "admin" {
		log.Printf("[ERROR] RegisterUser - Non-admin user attempted to register user, admin_id: %s, role: %s", claims.UserID, claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Success: false,
			Message: "Forbidden",
			Error:   "only admin can register users",
		})
	}

	var req usecases.UserRegisterRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] RegisterUser - Invalid request format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request format",
			Error:   err.Error(),
		})
	}

	// Validate input
	if req.Email == "" || req.Role == "" || req.FirstName == "" || req.LastName == "" {
		log.Printf("[ERROR] RegisterUser - Missing required fields for email: %s", req.Email)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Email, role, first_name, and last_name are required",
			Error:   "missing required fields",
		})
	}

	log.Printf("[INFO] RegisterUser - Creating %s user: %s (by admin: %s)", req.Role, req.Email, claims.UserID)

	// Register user
	registerResp, err := ah.authUseCase.RegisterUser(&req)
	if err != nil {
		log.Printf("[ERROR] RegisterUser - Registration failed for email: %s, error: %v", req.Email, err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "User registration failed",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] RegisterUser - User registered successfully, email: %s, user_id: %s, role: %s, by_admin: %s", req.Email, registerResp.UserID, req.Role, claims.UserID)

	return c.Status(fiber.StatusCreated).JSON(response{
		Success: true,
		Message: "User registered successfully. Temporary password sent to email.",
		Data:    registerResp,
	})
}

// extractTokenFromHeader extracts JWT token from Authorization header
func extractTokenFromHeader(c *fiber.Ctx) string {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Expected format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}
