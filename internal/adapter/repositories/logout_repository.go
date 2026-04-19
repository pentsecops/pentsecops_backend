package repositories

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// LogoutRepository handles token blacklist operations
type LogoutRepository struct {
	db *sql.DB
}

// NewLogoutRepository creates a new logout repository
func NewLogoutRepository(db *sql.DB) *LogoutRepository {
	return &LogoutRepository{db: db}
}

// BlacklistToken adds a token to the blacklist (revokes it)
func (lr *LogoutRepository) BlacklistToken(token, userID, userType string, expiresAt time.Time) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if userType == "" {
		return fmt.Errorf("user type is required")
	}

	// Hash the token before storing (security best practice)
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashStr := hex.EncodeToString(tokenHash[:])

	query := `
		INSERT INTO token_blacklist (token_hash, user_id, user_type, expires_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT(token_hash) DO NOTHING
	`

	result, err := lr.db.Exec(query, tokenHashStr, userID, userType, expiresAt)
	if err != nil {
		log.Printf("[ERROR] Failed to blacklist token: %v", err)
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Failed to get rows affected: %v", err)
		return err
	}

	if rowsAffected > 0 {
		log.Printf("[SUCCESS] Token blacklisted for user %s (type: %s)", userID, userType)
	} else {
		log.Printf("[WARN] Token already blacklisted for user %s (type: %s)", userID, userType)
	}

	return nil
}

// IsTokenBlacklisted checks if a token is in the blacklist
func (lr *LogoutRepository) IsTokenBlacklisted(token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token is required")
	}

	// Hash the token
	tokenHash := sha256.Sum256([]byte(token))
	tokenHashStr := hex.EncodeToString(tokenHash[:])

	query := `
		SELECT EXISTS (
			SELECT 1 FROM token_blacklist
			WHERE token_hash = $1 AND expires_at > CURRENT_TIMESTAMP
		)
	`

	var exists bool
	err := lr.db.QueryRow(query, tokenHashStr).Scan(&exists)
	if err != nil {
		log.Printf("[ERROR] Failed to check token blacklist: %v", err)
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	return exists, nil
}

// CleanExpiredTokens removes expired tokens from blacklist
func (lr *LogoutRepository) CleanExpiredTokens() (int64, error) {
	query := `DELETE FROM token_blacklist WHERE expires_at <= CURRENT_TIMESTAMP`

	result, err := lr.db.Exec(query)
	if err != nil {
		log.Printf("[ERROR] Failed to clean expired tokens: %v", err)
		return 0, fmt.Errorf("failed to clean expired tokens: %w", err)
	}

	rowsDeleted, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Failed to get rows affected: %v", err)
		return 0, err
	}

	if rowsDeleted > 0 {
		log.Printf("[INFO] Cleaned %d expired tokens from blacklist", rowsDeleted)
	}

	return rowsDeleted, nil
}
