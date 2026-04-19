package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/middleware"
)

// LogoutHandler handles logout-related HTTP requests
type LogoutHandler struct {
	logoutUseCase   *usecases.LogoutUseCase
	activityUseCase *usecases.ActivityUseCase
	tokenManager    *middleware.TokenManager
}

// NewLogoutHandler creates a new logout handler
func NewLogoutHandler(logoutUseCase *usecases.LogoutUseCase, tokenManager *middleware.TokenManager, activityUseCase *usecases.ActivityUseCase) *LogoutHandler {
	return &LogoutHandler{
		logoutUseCase:   logoutUseCase,
		activityUseCase: activityUseCase,
		tokenManager:    tokenManager,
	}
}

// Logout handles POST /api/v1/logout requests
func (lh *LogoutHandler) Logout(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/logout - Logout request")

	// Extract token from header
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] Logout: Missing authentication token")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: missing token",
		})
	}

	// Verify token is valid
	claims, err := lh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] Logout: Token validation failed: %v", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Unauthorized: invalid token",
		})
	}

	// Get token expiration time
	tokenExpiresAt := claims.Exp

	// Logout the user (invalidate token)
	err = lh.logoutUseCase.Logout(tokenString, claims.UserID, claims.Role, tokenExpiresAt)
	if err != nil {
		log.Printf("[ERROR] Logout: Failed to logout user: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to logout",
		})
	}

	log.Printf("[SUCCESS] Logout: User %s (role: %s) logged out successfully", claims.UserID, claims.Role)

	// Log activity asynchronously
	go lh.activityUseCase.LogActivity(
		claims.UserID,
		domain.ActivityTypeUserLogout,
		fmt.Sprintf("User %s logged out", claims.Email),
		map[string]interface{}{"role": claims.Role},
		c.IP(),
		c.Get("User-Agent"),
	)

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Logged out successfully",
	})
}
