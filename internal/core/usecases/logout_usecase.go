package usecases

import (
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/adapter/repositories"
)

// LogoutUseCase handles logout business logic
type LogoutUseCase struct {
	logoutRepo *repositories.LogoutRepository
}

// NewLogoutUseCase creates a new logout use case
func NewLogoutUseCase(logoutRepo *repositories.LogoutRepository) *LogoutUseCase {
	return &LogoutUseCase{
		logoutRepo: logoutRepo,
	}
}

// Logout invalidates the user's token by adding it to the blacklist
func (lu *LogoutUseCase) Logout(token, userID, userType string, tokenExpiresAt time.Time) error {
	if token == "" {
		log.Printf("[ERROR] Logout: Token is required")
		return fmt.Errorf("token is required")
	}

	if userID == "" {
		log.Printf("[ERROR] Logout: User ID is required")
		return fmt.Errorf("user ID is required")
	}

	if userType == "" {
		log.Printf("[ERROR] Logout: User type is required")
		return fmt.Errorf("user type is required")
	}

	// Add token to blacklist
	err := lu.logoutRepo.BlacklistToken(token, userID, userType, tokenExpiresAt)
	if err != nil {
		log.Printf("[ERROR] Logout: Failed to blacklist token: %v", err)
		return fmt.Errorf("failed to logout: %w", err)
	}

	log.Printf("[SUCCESS] Logout: User %s (type: %s) successfully logged out", userID, userType)
	return nil
}

// CleanExpiredTokens removes expired tokens from the blacklist (can be called periodically)
func (lu *LogoutUseCase) CleanExpiredTokens() error {
	rowsDeleted, err := lu.logoutRepo.CleanExpiredTokens()
	if err != nil {
		log.Printf("[ERROR] CleanExpiredTokens: Failed to clean tokens: %v", err)
		return err
	}

	log.Printf("[INFO] CleanExpiredTokens: Removed %d expired tokens from blacklist", rowsDeleted)
	return nil
}
