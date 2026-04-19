package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/middleware"
)

// PasswordHandler handles password-related HTTP requests
type PasswordHandler struct {
	passwordUseCase *usecases.PasswordUseCase
	tokenManager    *middleware.TokenManager
}

// NewPasswordHandler creates a new password handler
func NewPasswordHandler(
	passwordUseCase *usecases.PasswordUseCase,
	tokenManager *middleware.TokenManager,
) *PasswordHandler {
	return &PasswordHandler{
		passwordUseCase: passwordUseCase,
		tokenManager:    tokenManager,
	}
}

// ChangePassword handles password change for authenticated users (admin or regular users)
// POST /api/v1/change-password
func (ph *PasswordHandler) ChangePassword(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/change-password - User changing password")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] ChangePassword - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := ph.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] ChangePassword - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	var req usecases.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] ChangePassword - Invalid request format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   fmt.Sprintf("invalid request: %v", err),
		})
	}

	// Change password based on role
	var changeErr error
	if claims.Role == "admin" {
		log.Printf("[INFO] ChangePassword - Admin changing password, admin_id: %s", claims.UserID)
		changeErr = ph.passwordUseCase.ChangeAdminPassword(claims.UserID, &req)
	} else {
		log.Printf("[INFO] ChangePassword - User changing password, user_id: %s, role: %s", claims.UserID, claims.Role)
		changeErr = ph.passwordUseCase.ChangeUserPassword(claims.UserID, &req)
	}

	if changeErr != nil {
		log.Printf("[ERROR] ChangePassword - Password change failed: %v", changeErr)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   changeErr.Error(),
		})
	}

	log.Printf("[SUCCESS] ChangePassword - Password changed successfully, user_id: %s, role: %s", claims.UserID, claims.Role)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Password changed successfully",
		Data: usecases.ChangePasswordResponse{
			Success: true,
			Message: "Password updated successfully",
			UserID:  claims.UserID,
			Email:   claims.Email,
		},
	})
}
