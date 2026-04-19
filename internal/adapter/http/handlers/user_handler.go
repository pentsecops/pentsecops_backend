package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/usecases"
)

// ProfileHandler handles profile HTTP requests
type ProfileHandler struct {
	profileUseCase *usecases.ProfileUseCase
	authUseCase    *usecases.AuthUseCase
}

// NewProfileHandler creates a new profile handler
func NewProfileHandler(profileUseCase *usecases.ProfileUseCase, authUseCase *usecases.AuthUseCase) *ProfileHandler {
	return &ProfileHandler{
		profileUseCase: profileUseCase,
		authUseCase:    authUseCase,
	}
}

// GetProfile handles user profile retrieval endpoint GET /api/v1/profile
func (ph *ProfileHandler) GetProfile(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/profile - Profile retrieval attempt")

	// Extract token from header
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] GetProfile - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	// Validate token
	claims, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] GetProfile - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	log.Printf("[INFO] GetProfile - Retrieving profile for user_id: %s, role: %s", claims.UserID, claims.Role)

	// Get user profile based on role
	profileData, err := ph.profileUseCase.GetUserProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] GetProfile - Profile not found for user_id: %s, error: %v", claims.UserID, err)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "Profile not found",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetProfile - Profile retrieved successfully for user_id: %s, role: %s", claims.UserID, claims.Role)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Profile retrieved successfully",
		Data:    profileData,
	})
}
