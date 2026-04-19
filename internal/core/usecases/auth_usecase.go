package usecases

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/pkg/middleware"
	"github.com/pentsecops/backend/pkg/services"
	"github.com/pentsecops/backend/pkg/utils"
)

// AuthUseCase handles authentication business logic
type AuthUseCase struct {
	userRepo        domain.UserRepository
	adminRepo       domain.AdminRepository
	pentesterRepo   domain.PentesterRepository
	stakeholderRepo domain.StakeholderRepository
	tokenManager    *middleware.TokenManager
	emailService    *services.EmailService
}

// NewAuthUseCase creates a new authentication use case
func NewAuthUseCase(
	userRepo domain.UserRepository,
	adminRepo domain.AdminRepository,
	pentesterRepo domain.PentesterRepository,
	stakeholderRepo domain.StakeholderRepository,
	tokenManager *middleware.TokenManager,
	emailService *services.EmailService,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:        userRepo,
		adminRepo:       adminRepo,
		pentesterRepo:   pentesterRepo,
		stakeholderRepo: stakeholderRepo,
		tokenManager:    tokenManager,
		emailService:    emailService,
	}
}

// LoginResponse contains the response data after successful login
type LoginResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Token  string `json:"token"`
}

// AdminLoginResponse contains the response data after successful admin login
type AdminLoginResponse struct {
	AdminID string `json:"admin_id"`
	Email   string `json:"email"`
	Token   string `json:"token"`
}

// LoginRequest contains the login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminRegisterRequest contains the data for admin registration
type AdminRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// UserRegisterRequest contains the data for user registration by admin
type UserRegisterRequest struct {
	Email           string   `json:"email"`
	Role            string   `json:"role"` // "pentester" or "stakeholder"
	FirstName       string   `json:"first_name"`
	LastName        string   `json:"last_name"`
	Specialization  string   `json:"specialization,omitempty"`   // For pentesters only
	ExperienceYears int      `json:"experience_years,omitempty"` // For pentesters only
	Skills          []string `json:"skills,omitempty"`           // For pentesters only - array of skills
	Certifications  []string `json:"certifications,omitempty"`   // For pentesters only - array of certifications
	Company         string   `json:"company,omitempty"`          // For stakeholders only
	Address         string   `json:"address,omitempty"`          // For stakeholders only
}

// UserRegisterResponse contains the response after user registration
type UserRegisterResponse struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Password string `json:"temporary_password"` // Temporary password to send to user
}

// LoginUser handles user login (pentester or stakeholder)
func (au *AuthUseCase) LoginUser(req *LoginRequest) (*LoginResponse, error) {
	// Check if user exists
	user, err := au.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is active
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is not active")
	}

	// Verify password
	valid, err := utils.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Update last login
	if err := au.userRepo.UpdateUserLastLogin(user.ID); err != nil {
		return nil, fmt.Errorf("failed to update login time: %w", err)
	}

	// Generate PASETO token
	token, err := au.tokenManager.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Token:  token,
	}, nil
}

// LoginAdmin handles admin login
func (au *AuthUseCase) LoginAdmin(req *LoginRequest) (*AdminLoginResponse, error) {
	// Check if admin exists
	admin, err := au.adminRepo.GetAdminByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	valid, err := utils.VerifyPassword(req.Password, admin.PasswordHash)
	if err != nil || !valid {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Generate PASETO token
	token, err := au.tokenManager.GenerateToken(admin.ID, admin.Email, "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AdminLoginResponse{
		AdminID: admin.ID,
		Email:   admin.Email,
		Token:   token,
	}, nil
}

// RegisterAdmin creates a new admin account
func (au *AuthUseCase) RegisterAdmin(req *AdminRegisterRequest) (*AdminLoginResponse, error) {
	// Check if admin already exists
	_, err := au.adminRepo.GetAdminByEmail(req.Email)
	if err == nil {
		return nil, fmt.Errorf("admin with this email already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin
	admin := &domain.Admin{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := au.adminRepo.CreateAdmin(admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}

	// Generate PASETO token
	token, err := au.tokenManager.GenerateToken(admin.ID, admin.Email, "admin")
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &AdminLoginResponse{
		AdminID: admin.ID,
		Email:   admin.Email,
		Token:   token,
	}, nil
}

// RegisterUser creates a new user (pentester or stakeholder) - called by admin only
func (au *AuthUseCase) RegisterUser(req *UserRegisterRequest) (*UserRegisterResponse, error) {
	// Validate role
	if req.Role != "pentester" && req.Role != "stakeholder" {
		return nil, fmt.Errorf("invalid role: must be 'pentester' or 'stakeholder'")
	}

	// Check if user already exists
	_, err := au.userRepo.GetUserByEmail(req.Email)
	if err == nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Generate temporary password
	tempPassword, err := utils.GenerateRandomPassword(12)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	// Hash password
	passwordHash, err := utils.HashPassword(tempPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID := uuid.New().String()
	user := &domain.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
		Status:       "active",
	}

	if err := au.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create role-specific profile
	if req.Role == "pentester" {
		// Convert skills and certifications arrays to JSON strings
		skillsJSON, _ := json.Marshal(req.Skills)
		certificationsJSON, _ := json.Marshal(req.Certifications)

		pentester := &domain.Pentester{
			ID:              uuid.New().String(),
			UserID:          userID,
			FirstName:       req.FirstName,
			LastName:        req.LastName,
			Email:           req.Email,
			Specialization:  req.Specialization,
			ExperienceYears: req.ExperienceYears,
			Status:          "active",
			Skills:          string(skillsJSON),
			Certifications:  string(certificationsJSON),
		}

		if err := au.pentesterRepo.CreatePentester(pentester); err != nil {
			return nil, fmt.Errorf("failed to create pentester profile: %w", err)
		}
	} else if req.Role == "stakeholder" {
		stakeholder := &domain.Stakeholder{
			ID:        uuid.New().String(),
			UserID:    userID,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Company:   req.Company,
			Address:   req.Address,
		}

		if err := au.stakeholderRepo.CreateStakeholder(stakeholder); err != nil {
			return nil, fmt.Errorf("failed to create stakeholder profile: %w", err)
		}
	}

	// Send email notification with temporary password if email service is available
	if au.emailService != nil {
		if err := au.emailService.SendUserRegistrationEmail(req.Email, req.FirstName, tempPassword); err != nil {
			// Log the error but don't fail the registration
			fmt.Printf("Warning: failed to send registration email to %s: %v\n", req.Email, err)
		}
	}

	return &UserRegisterResponse{
		UserID:   userID,
		Email:    req.Email,
		Role:     req.Role,
		Password: tempPassword,
	}, nil
}

// ValidateToken validates a PASETO token
func (au *AuthUseCase) ValidateToken(token string) (*middleware.PasetoTokenClaims, error) {
	claims, err := au.tokenManager.VerifyToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	return claims, nil
}
