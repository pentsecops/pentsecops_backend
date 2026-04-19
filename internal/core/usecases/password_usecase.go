package usecases

import (
	"fmt"

	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/pkg/utils"
)

// PasswordUseCase handles password-related operations
type PasswordUseCase struct {
	userRepo  domain.UserRepository
	adminRepo domain.AdminRepository
}

// NewPasswordUseCase creates a new password use case
func NewPasswordUseCase(
	userRepo domain.UserRepository,
	adminRepo domain.AdminRepository,
) *PasswordUseCase {
	return &PasswordUseCase{
		userRepo:  userRepo,
		adminRepo: adminRepo,
	}
}

// ChangePasswordRequest contains the request data for password change
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// ChangePasswordResponse contains the response after successful password change
type ChangePasswordResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
	Email   string `json:"email,omitempty"`
}

// ValidatePasswordChange validates the password change request
func (pu *PasswordUseCase) ValidatePasswordChange(req *ChangePasswordRequest) error {
	if req.CurrentPassword == "" {
		return fmt.Errorf("current password is required")
	}

	if req.NewPassword == "" {
		return fmt.Errorf("new password is required")
	}

	if req.ConfirmPassword == "" {
		return fmt.Errorf("confirm password is required")
	}

	if req.NewPassword != req.ConfirmPassword {
		return fmt.Errorf("new password and confirm password do not match")
	}

	if len(req.NewPassword) < 8 {
		return fmt.Errorf("new password must be at least 8 characters")
	}

	if req.CurrentPassword == req.NewPassword {
		return fmt.Errorf("new password cannot be the same as current password")
	}

	return nil
}

// ChangeUserPassword changes password for a regular user (pentester/stakeholder)
func (pu *PasswordUseCase) ChangeUserPassword(userID string, req *ChangePasswordRequest) error {
	// Validate request
	if err := pu.ValidatePasswordChange(req); err != nil {
		return err
	}

	// Get user
	user, err := pu.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Verify current password
	isValid, err := utils.VerifyPassword(req.CurrentPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}
	if !isValid {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	user.PasswordHash = newPasswordHash
	if err := pu.userRepo.UpdateUser(user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// ChangeAdminPassword changes password for an admin
func (pu *PasswordUseCase) ChangeAdminPassword(adminID string, req *ChangePasswordRequest) error {
	// Validate request
	if err := pu.ValidatePasswordChange(req); err != nil {
		return err
	}

	// Get admin
	admin, err := pu.adminRepo.GetAdminByID(adminID)
	if err != nil {
		return fmt.Errorf("admin not found: %w", err)
	}

	// Verify current password
	isValid, err := utils.VerifyPassword(req.CurrentPassword, admin.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to verify password: %w", err)
	}
	if !isValid {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	newPasswordHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Update password
	admin.PasswordHash = newPasswordHash
	if err := pu.adminRepo.UpdateAdmin(admin); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}
