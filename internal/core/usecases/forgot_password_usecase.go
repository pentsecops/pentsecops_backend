package usecases

import (
	"fmt"
	"math/rand"

	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/pkg/services"
	"github.com/pentsecops/backend/pkg/utils"
)

type ForgotPasswordUseCase struct {
	forgotPasswordRepo domain.ForgotPasswordRequestRepository
	userRepo           domain.UserRepository
	pentesterRepo      domain.PentesterRepository
	stakeholderRepo    domain.StakeholderRepository
	emailService       *services.EmailService
}

func NewForgotPasswordUseCase(
	forgotPasswordRepo domain.ForgotPasswordRequestRepository,
	userRepo domain.UserRepository,
	pentesterRepo domain.PentesterRepository,
	stakeholderRepo domain.StakeholderRepository,
	emailService *services.EmailService,
) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		forgotPasswordRepo: forgotPasswordRepo,
		userRepo:           userRepo,
		pentesterRepo:      pentesterRepo,
		stakeholderRepo:    stakeholderRepo,
		emailService:       emailService,
	}
}

// RequestPasswordReset creates a password reset request
func (fpu *ForgotPasswordUseCase) RequestPasswordReset(email string) error {
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Check if user exists
	user, err := fpu.userRepo.GetUserByEmail(email)
	if err != nil {
		// For security, don't reveal if email exists or not
		return nil
	}

	if user == nil {
		// For security, don't reveal if email exists or not
		return nil
	}

	// Create the reset request with user_id
	_, err = fpu.forgotPasswordRepo.CreateRequest(user.ID, email)
	if err != nil {
		return fmt.Errorf("failed to create reset request: %w", err)
	}

	return nil
}

// GetPendingRequests retrieves all pending password reset requests (admin only)
func (fpu *ForgotPasswordUseCase) GetPendingRequests() ([]domain.ForgotPasswordRequest, error) {
	requests, err := fpu.forgotPasswordRepo.GetPendingRequests()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending requests: %w", err)
	}

	return requests, nil
}

// ApproveRequest approves a password reset request and sends new password to user
func (fpu *ForgotPasswordUseCase) ApproveRequest(requestID string) (*domain.ForgotPasswordRequest, error) {
	// Get the request
	request, err := fpu.forgotPasswordRepo.GetRequestByID(requestID)
	if err != nil {
		return nil, fmt.Errorf("request not found: %w", err)
	}

	// Generate temporary password (12 characters: uppercase, lowercase, numbers, special)
	tempPassword := fpu.generateTemporaryPassword()

	// Hash the password
	passwordHash, err := utils.HashPassword(tempPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Get user by ID
	user, err := fpu.userRepo.GetUserByID(request.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update user password
	user.PasswordHash = passwordHash
	err = fpu.userRepo.UpdateUser(user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user password: %w", err)
	}

	// Update request status to accepted with temp password
	err = fpu.forgotPasswordRepo.UpdateRequestStatus(requestID, "accepted", tempPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to update request: %w", err)
	}

	// Get user's first name from profile and send email
	var firstName string
	if user.Role == "pentester" {
		pentester, err := fpu.pentesterRepo.GetPentesterByUserID(user.ID)
		if err == nil && pentester != nil {
			firstName = pentester.FirstName
		} else {
			firstName = "User"
		}
	} else if user.Role == "stakeholder" {
		stakeholder, err := fpu.stakeholderRepo.GetStakeholderByUserID(user.ID)
		if err == nil && stakeholder != nil {
			firstName = stakeholder.FirstName
		} else {
			firstName = "User"
		}
	} else {
		firstName = "User"
	}

	// Send email with new password
	if fpu.emailService.IsEnabled() {
		err = fpu.emailService.SendForgotPasswordApprovedEmail(request.UserEmail, firstName, tempPassword)
		if err != nil {
			// Log error but don't fail - email service is optional
			fmt.Printf("[WARNING] Failed to send password reset email: %v\n", err)
		}
	}

	request.Status = "accepted"
	request.TemporaryPassword = tempPassword

	return request, nil
}

// RejectRequest rejects a password reset request
func (fpu *ForgotPasswordUseCase) RejectRequest(requestID string) error {
	// Verify the request exists
	_, err := fpu.forgotPasswordRepo.GetRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("request not found: %w", err)
	}

	// Update request status to rejected
	err = fpu.forgotPasswordRepo.RejectRequest(requestID)
	if err != nil {
		return fmt.Errorf("failed to reject request: %w", err)
	}

	return nil
}

// generateTemporaryPassword generates a 12-character password with mixed case, numbers, and special chars
func (fpu *ForgotPasswordUseCase) generateTemporaryPassword() string {
	const (
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		numbers   = "0123456789"
		special   = "!@#$%^&*"
	)

	all := uppercase + lowercase + numbers + special

	password := make([]byte, 12)

	// Ensure at least one character from each category
	password[0] = uppercase[rand.Intn(len(uppercase))]
	password[1] = lowercase[rand.Intn(len(lowercase))]
	password[2] = numbers[rand.Intn(len(numbers))]
	password[3] = special[rand.Intn(len(special))]

	// Fill the rest randomly
	for i := 4; i < 12; i++ {
		password[i] = all[rand.Intn(len(all))]
	}

	// Shuffle the password
	shuffled := []byte(password)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return string(shuffled)
}
