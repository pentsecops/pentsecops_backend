package usecases

import (
	"encoding/json"
	"fmt"

	"github.com/pentsecops/backend/internal/core/domain"
)

// UserManagementUseCase handles user management business logic for admins and users
type UserManagementUseCase struct {
	userRepo        domain.UserRepository
	pentesterRepo   domain.PentesterRepository
	stakeholderRepo domain.StakeholderRepository
}

// NewUserManagementUseCase creates a new user management use case
func NewUserManagementUseCase(
	userRepo domain.UserRepository,
	pentesterRepo domain.PentesterRepository,
	stakeholderRepo domain.StakeholderRepository,
) *UserManagementUseCase {
	return &UserManagementUseCase{
		userRepo:        userRepo,
		pentesterRepo:   pentesterRepo,
		stakeholderRepo: stakeholderRepo,
	}
}

// GetUserDetails retrieves complete user details including profile
type UserDetails struct {
	User      *domain.UserResponse `json:"user"`
	Profile   interface{}          `json:"profile,omitempty"`
	ProfileID string               `json:"profile_id,omitempty"`
}

// GetUserByID retrieves user details by ID
func (uc *UserManagementUseCase) GetUserByID(userID string) (*UserDetails, error) {
	user, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	details := &UserDetails{
		User: user.ToResponse(),
	}

	// Get profile based on role
	if user.Role == "pentester" {
		pentester, err := uc.pentesterRepo.GetPentesterByUserID(userID)
		if err == nil && pentester != nil {
			details.Profile = pentester
			details.ProfileID = pentester.ID
		}
	} else if user.Role == "stakeholder" {
		stakeholder, err := uc.stakeholderRepo.GetStakeholderByUserID(userID)
		if err == nil && stakeholder != nil {
			details.Profile = stakeholder
			details.ProfileID = stakeholder.ID
		}
	}

	return details, nil
}

// UpdateUserStatusRequest contains status update request
type UpdateUserStatusRequest struct {
	Status string `json:"status"` // active, freeze, suspended
}

// UpdateUserStatus changes the status of a user (freeze/unfreeze/suspend)
func (uc *UserManagementUseCase) UpdateUserStatus(userID string, status string) error {
	// Validate status
	validStatuses := map[string]bool{
		"active":    true,
		"freeze":    true,
		"suspended": true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s (must be active, freeze, or suspended)", status)
	}

	// Verify user exists
	_, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Update status
	err = uc.userRepo.UpdateUserStatus(userID, status)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	return nil
}

// DeleteUserRequest contains data for deleting a user
type DeleteUserRequest struct {
	UserID string `json:"user_id"`
}

// DeleteUser removes a user and their associated profile data
func (uc *UserManagementUseCase) DeleteUser(userID string) error {
	// Verify user exists
	_, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Note: Profile deletion is handled by CASCADE DELETE in database
	// when we delete the user

	// Delete user (this will cascade delete profile via foreign key)
	if err := uc.userRepo.DeleteUser(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// ListAllUsersResponse contains paginated or complete list of users
type ListAllUsersResponse struct {
	Total int            `json:"total"`
	Users []*UserDetails `json:"users"`
}

// GetAllUsers retrieves all users with their profiles
func (uc *UserManagementUseCase) GetAllUsers() (*ListAllUsersResponse, error) {
	users, err := uc.userRepo.GetAllUsers()
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	var userDetails []*UserDetails
	for _, user := range users {
		details := &UserDetails{
			User: user.ToResponse(),
		}

		// Get profile based on role
		if user.Role == "pentester" {
			pentester, err := uc.pentesterRepo.GetPentesterByUserID(user.ID)
			if err == nil && pentester != nil {
				details.Profile = pentester
				details.ProfileID = pentester.ID
			}
		} else if user.Role == "stakeholder" {
			stakeholder, err := uc.stakeholderRepo.GetStakeholderByUserID(user.ID)
			if err == nil && stakeholder != nil {
				details.Profile = stakeholder
				details.ProfileID = stakeholder.ID
			}
		}

		userDetails = append(userDetails, details)
	}

	return &ListAllUsersResponse{
		Total: len(userDetails),
		Users: userDetails,
	}, nil
}

// UpdateProfileRequest contains generic profile update request
type UpdateProfileRequest struct {
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Email           string   `json:"email,omitempty"`
	Company         string   `json:"company,omitempty"`          // For stakeholders
	Address         string   `json:"address,omitempty"`          // For stakeholders
	Specialization  string   `json:"specialization,omitempty"`   // For pentesters
	ExperienceYears int      `json:"experience_years,omitempty"` // For pentesters
	Skills          []string `json:"skills,omitempty"`           // For pentesters
	Certifications  []string `json:"certifications,omitempty"`   // For pentesters
	About           string   `json:"about,omitempty"`            // For stakeholders
}

// UpdateUserProfile updates a user's profile (for the user themselves)
func (uc *UserManagementUseCase) UpdateUserProfile(userID string, req *UpdateProfileRequest) error {
	// Get user to determine role
	user, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.Role == "pentester" {
		// Get pentester profile
		pentester, err := uc.pentesterRepo.GetPentesterByUserID(userID)
		if err != nil {
			return fmt.Errorf("pentester profile not found: %w", err)
		}

		// Update fields
		if req.FirstName != "" {
			pentester.FirstName = req.FirstName
		}
		if req.LastName != "" {
			pentester.LastName = req.LastName
		}
		if req.Email != "" {
			pentester.Email = req.Email
		}
		if req.Specialization != "" {
			pentester.Specialization = req.Specialization
		}
		if req.ExperienceYears > 0 {
			pentester.ExperienceYears = req.ExperienceYears
		}

		// Handle skills array
		if len(req.Skills) > 0 {
			skillsJSON, _ := json.Marshal(req.Skills)
			pentester.Skills = string(skillsJSON)
		}

		// Handle certifications array
		if len(req.Certifications) > 0 {
			certificationsJSON, _ := json.Marshal(req.Certifications)
			pentester.Certifications = string(certificationsJSON)
		}

		// Update pentester profile
		err = uc.pentesterRepo.UpdatePentester(pentester)
		if err != nil {
			return fmt.Errorf("failed to update pentester profile: %w", err)
		}

	} else if user.Role == "stakeholder" {
		// Get stakeholder profile
		stakeholder, err := uc.stakeholderRepo.GetStakeholderByUserID(userID)
		if err != nil {
			return fmt.Errorf("stakeholder profile not found: %w", err)
		}

		// Update fields
		if req.FirstName != "" {
			stakeholder.FirstName = req.FirstName
		}
		if req.LastName != "" {
			stakeholder.LastName = req.LastName
		}
		if req.Email != "" {
			stakeholder.Email = req.Email
		}
		if req.Company != "" {
			stakeholder.Company = req.Company
		}
		if req.Address != "" {
			stakeholder.Address = req.Address
		}

		// Update stakeholder profile
		err = uc.stakeholderRepo.UpdateStakeholder(stakeholder)
		if err != nil {
			return fmt.Errorf("failed to update stakeholder profile: %w", err)
		}
	}

	return nil
}

// GetUserProfile retrieves the profile of a user (for the user themselves)
func (uc *UserManagementUseCase) GetUserProfile(userID string) (*UserDetails, error) {
	return uc.GetUserByID(userID)
}

// AdminUpdateUserProfile updates a user's profile (admin operation)
func (uc *UserManagementUseCase) AdminUpdateUserProfile(userID string, req *UpdateProfileRequest) error {
	// Get user to determine role
	user, err := uc.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if user.Role == "pentester" {
		// Get pentester profile
		pentester, err := uc.pentesterRepo.GetPentesterByUserID(userID)
		if err != nil {
			return fmt.Errorf("pentester profile not found: %w", err)
		}

		// Update fields
		if req.FirstName != "" {
			pentester.FirstName = req.FirstName
		}
		if req.LastName != "" {
			pentester.LastName = req.LastName
		}
		if req.Email != "" {
			pentester.Email = req.Email
		}
		if req.Specialization != "" {
			pentester.Specialization = req.Specialization
		}
		if req.ExperienceYears > 0 {
			pentester.ExperienceYears = req.ExperienceYears
		}

		// Handle skills array
		if len(req.Skills) > 0 {
			skillsJSON, _ := json.Marshal(req.Skills)
			pentester.Skills = string(skillsJSON)
		}

		// Handle certifications array
		if len(req.Certifications) > 0 {
			certificationsJSON, _ := json.Marshal(req.Certifications)
			pentester.Certifications = string(certificationsJSON)
		}

		// Update pentester profile
		err = uc.pentesterRepo.UpdatePentester(pentester)
		if err != nil {
			return fmt.Errorf("failed to update pentester profile: %w", err)
		}

	} else if user.Role == "stakeholder" {
		// Get stakeholder profile
		stakeholder, err := uc.stakeholderRepo.GetStakeholderByUserID(userID)
		if err != nil {
			return fmt.Errorf("stakeholder profile not found: %w", err)
		}

		// Update fields
		if req.FirstName != "" {
			stakeholder.FirstName = req.FirstName
		}
		if req.LastName != "" {
			stakeholder.LastName = req.LastName
		}
		if req.Email != "" {
			stakeholder.Email = req.Email
		}
		if req.Company != "" {
			stakeholder.Company = req.Company
		}
		if req.Address != "" {
			stakeholder.Address = req.Address
		}

		// Update stakeholder profile
		err = uc.stakeholderRepo.UpdateStakeholder(stakeholder)
		if err != nil {
			return fmt.Errorf("failed to update stakeholder profile: %w", err)
		}
	}

	return nil
}

// PaginatedListAllUsersResponse contains paginated list of users
type PaginatedListAllUsersResponse struct {
	Total      int64          `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int64          `json:"total_pages"`
	Users      []*UserDetails `json:"users"`
}

// GetAllUsersPaginated retrieves paginated users with their profiles
func (uc *UserManagementUseCase) GetAllUsersPaginated(page, limit int) (*PaginatedListAllUsersResponse, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	users, total, err := uc.userRepo.GetAllUsersPaginated(page, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated users: %w", err)
	}

	var userDetails []*UserDetails
	for _, user := range users {
		details := &UserDetails{
			User: user.ToResponse(),
		}

		// Get profile based on role
		if user.Role == "pentester" {
			pentester, err := uc.pentesterRepo.GetPentesterByUserID(user.ID)
			if err == nil && pentester != nil {
				details.Profile = pentester
				details.ProfileID = pentester.ID
			}
		} else if user.Role == "stakeholder" {
			stakeholder, err := uc.stakeholderRepo.GetStakeholderByUserID(user.ID)
			if err == nil && stakeholder != nil {
				details.Profile = stakeholder
				details.ProfileID = stakeholder.ID
			}
		}

		userDetails = append(userDetails, details)
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	return &PaginatedListAllUsersResponse{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		Users:      userDetails,
	}, nil
}
