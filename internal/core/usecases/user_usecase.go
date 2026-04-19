package usecases

import (
	"fmt"

	"github.com/pentsecops/backend/internal/core/domain"
)

// ProfileUseCase handles profile retrieval business logic
type ProfileUseCase struct {
	userRepo        domain.UserRepository
	pentesterRepo   domain.PentesterRepository
	stakeholderRepo domain.StakeholderRepository
}

// NewProfileUseCase creates a new profile use case
func NewProfileUseCase(
	userRepo domain.UserRepository,
	pentesterRepo domain.PentesterRepository,
	stakeholderRepo domain.StakeholderRepository,
) *ProfileUseCase {
	return &ProfileUseCase{
		userRepo:        userRepo,
		pentesterRepo:   pentesterRepo,
		stakeholderRepo: stakeholderRepo,
	}
}

// PentesterProfileResponse contains full pentester profile data
type PentesterProfileResponse struct {
	User    *domain.UserResponse      `json:"user"`
	Profile *domain.PentesterResponse `json:"profile"`
}

// StakeholderProfileResponse contains full stakeholder profile data
type StakeholderProfileResponse struct {
	User    *domain.UserResponse        `json:"user"`
	Profile *domain.StakeholderResponse `json:"profile"`
}

// GetPentesterProfile retrieves a pentester's full profile by user ID
func (pu *ProfileUseCase) GetPentesterProfile(userID string) (*PentesterProfileResponse, error) {
	// Get user information
	user, err := pu.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify user is a pentester
	if user.Role != "pentester" {
		return nil, fmt.Errorf("user is not a pentester")
	}

	// Get pentester profile
	pentester, err := pu.pentesterRepo.GetPentesterByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("pentester profile not found: %w", err)
	}

	return &PentesterProfileResponse{
		User:    user.ToResponse(),
		Profile: pentester.ToResponse(),
	}, nil
}

// GetStakeholderProfile retrieves a stakeholder's full profile by user ID
func (pu *ProfileUseCase) GetStakeholderProfile(userID string) (*StakeholderProfileResponse, error) {
	// Get user information
	user, err := pu.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify user is a stakeholder
	if user.Role != "stakeholder" {
		return nil, fmt.Errorf("user is not a stakeholder")
	}

	// Get stakeholder profile
	stakeholder, err := pu.stakeholderRepo.GetStakeholderByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("stakeholder profile not found: %w", err)
	}

	return &StakeholderProfileResponse{
		User:    user.ToResponse(),
		Profile: stakeholder.ToResponse(),
	}, nil
}

// GetUserProfile retrieves a user's profile based on their role
func (pu *ProfileUseCase) GetUserProfile(userID string) (interface{}, error) {
	// Get user information
	user, err := pu.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Retrieve profile based on role
	if user.Role == "pentester" {
		return pu.GetPentesterProfile(userID)
	} else if user.Role == "stakeholder" {
		return pu.GetStakeholderProfile(userID)
	}

	return nil, fmt.Errorf("unknown user role: %s", user.Role)
}
