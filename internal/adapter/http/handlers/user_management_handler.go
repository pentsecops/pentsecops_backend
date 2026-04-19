package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/middleware"
	"github.com/pentsecops/backend/pkg/services"
)

// UserManagementHandler handles user management HTTP requests
type UserManagementHandler struct {
	userManagementUseCase *usecases.UserManagementUseCase
	tokenManager          *middleware.TokenManager
	fileService           *services.FileService
	pentesterUseCase      *usecases.PentesterUseCase
}

// NewUserManagementHandler creates a new user management handler
func NewUserManagementHandler(
	userManagementUseCase *usecases.UserManagementUseCase,
	tokenManager *middleware.TokenManager,
	fileService *services.FileService,
	pentesterUseCase *usecases.PentesterUseCase,
) *UserManagementHandler {
	return &UserManagementHandler{
		userManagementUseCase: userManagementUseCase,
		tokenManager:          tokenManager,
		fileService:           fileService,
		pentesterUseCase:      pentesterUseCase,
	}
}

// GetUser retrieves a specific user by ID (admin only)
// GET /api/v1/admin/user/:user_id
func (umh *UserManagementHandler) GetUser(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/admin/user/:user_id - Admin retrieving user details")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] GetUser - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := umh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] GetUser - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	if claims.Role != "admin" {
		log.Printf("[ERROR] GetUser - Non-admin user attempted access, user_id: %s, role: %s", claims.UserID, claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Success: false,
			Message: "Forbidden",
			Error:   "admin access required",
		})
	}

	userID := c.Params("user_id")
	if userID == "" {
		log.Printf("[ERROR] GetUser - Missing user_id parameter")
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   "missing user_id",
		})
	}

	userDetails, err := umh.userManagementUseCase.GetUserByID(userID)
	if err != nil {
		log.Printf("[ERROR] GetUser - User not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "Not found",
			Error:   "user not found",
		})
	}

	log.Printf("[SUCCESS] GetUser - Retrieved user: user_id=%s, role=%s, admin_id=%s", userID, userDetails.User.Role, claims.UserID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "User retrieved successfully",
		Data:    userDetails,
	})
}

// UpdateUser updates a specific user (admin only)
// POST /api/v1/admin/user/:user_id
func (umh *UserManagementHandler) UpdateUser(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/admin/user/:user_id - Admin updating user")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] UpdateUser - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := umh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] UpdateUser - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	if claims.Role != "admin" {
		log.Printf("[ERROR] UpdateUser - Non-admin user attempted access, user_id: %s, role: %s", claims.UserID, claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Success: false,
			Message: "Forbidden",
			Error:   "admin access required",
		})
	}

	userID := c.Params("user_id")
	if userID == "" {
		log.Printf("[ERROR] UpdateUser - Missing user_id parameter")
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   "missing user_id",
		})
	}

	var req struct {
		Status          string `json:"status"`
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		Email           string `json:"email"`
		Company         string `json:"company"`          // For stakeholders
		Address         string `json:"address"`          // For stakeholders
		Specialization  string `json:"specialization"`   // For pentesters
		ExperienceYears int    `json:"experience_years"` // For pentesters
	}
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] UpdateUser - Invalid request format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   fmt.Sprintf("invalid request: %v", err),
		})
	}

	if req.Status != "" {
		// Update user status
		err := umh.userManagementUseCase.UpdateUserStatus(userID, req.Status)
		if err != nil {
			log.Printf("[ERROR] UpdateUser - Failed to update status: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Success: false,
				Message: "Bad request",
				Error:   err.Error(),
			})
		}
		log.Printf("[INFO] UpdateUser - User status updated: user_id=%s, new_status=%s", userID, req.Status)
	}

	// Update profile if any profile fields are provided
	if req.FirstName != "" || req.LastName != "" || req.Email != "" || req.Company != "" || req.Address != "" || req.Specialization != "" || req.ExperienceYears > 0 {
		profileReq := &usecases.UpdateProfileRequest{
			FirstName:       req.FirstName,
			LastName:        req.LastName,
			Email:           req.Email,
			Company:         req.Company,
			Address:         req.Address,
			Specialization:  req.Specialization,
			ExperienceYears: req.ExperienceYears,
		}

		err := umh.userManagementUseCase.UpdateUserProfile(userID, profileReq)
		if err != nil {
			log.Printf("[ERROR] UpdateUser - Failed to update profile: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Success: false,
				Message: "Bad request",
				Error:   err.Error(),
			})
		}
		log.Printf("[INFO] UpdateUser - User profile updated: user_id=%s", userID)
	}

	userDetails, err := umh.userManagementUseCase.GetUserByID(userID)
	if err != nil {
		log.Printf("[ERROR] UpdateUser - Failed to retrieve updated user: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Success: false,
			Message: "Internal server error",
			Error:   "failed to retrieve updated user",
		})
	}

	log.Printf("[SUCCESS] UpdateUser - User updated: user_id=%s, admin_id=%s", userID, claims.UserID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "User updated successfully",
		Data:    userDetails,
	})
}

// DeleteUser deletes a specific user (admin only)
// DELETE /api/v1/admin/user/:user_id
func (umh *UserManagementHandler) DeleteUser(c *fiber.Ctx) error {
	log.Printf("[API] DELETE /api/v1/admin/user/:user_id - Admin deleting user")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] DeleteUser - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := umh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] DeleteUser - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	if claims.Role != "admin" {
		log.Printf("[ERROR] DeleteUser - Non-admin user attempted access, user_id: %s, role: %s", claims.UserID, claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Success: false,
			Message: "Forbidden",
			Error:   "admin access required",
		})
	}

	userID := c.Params("user_id")
	if userID == "" {
		log.Printf("[ERROR] DeleteUser - Missing user_id parameter")
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   "missing user_id",
		})
	}

	err = umh.userManagementUseCase.DeleteUser(userID)
	if err != nil {
		log.Printf("[ERROR] DeleteUser - Failed to delete user: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "Not found",
			Error:   "user not found",
		})
	}

	log.Printf("[SUCCESS] DeleteUser - User deleted: user_id=%s, admin_id=%s", userID, claims.UserID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ListAllUsers retrieves all users (admin only)
// GET /api/v1/admin/users
func (umh *UserManagementHandler) ListAllUsers(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/admin/users - Admin listing all users")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] ListAllUsers - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := umh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] ListAllUsers - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	if claims.Role != "admin" {
		log.Printf("[ERROR] ListAllUsers - Non-admin user attempted access, user_id: %s, role: %s", claims.UserID, claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Success: false,
			Message: "Forbidden",
			Error:   "admin access required",
		})
	}

	// Extract pagination parameters
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

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

	users, err := umh.userManagementUseCase.GetAllUsersPaginated(page, limit)
	if err != nil {
		log.Printf("[ERROR] ListAllUsers - Failed to retrieve users: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Success: false,
			Message: "Internal server error",
			Error:   "failed to retrieve users",
		})
	}

	log.Printf("[SUCCESS] ListAllUsers - Retrieved %d/%d users (page=%d, limit=%d), admin_id=%s", len(users.Users), users.Total, page, limit, claims.UserID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

// UpdateProfile updates the user's own profile (supports resume upload for pentesters)
// POST /api/v1/profile
// Content-Type: multipart/form-data or application/json
func (umh *UserManagementHandler) UpdateProfile(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/profile - User updating own profile")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] UpdateProfile - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := umh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] UpdateProfile - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	// Parse form data (supports both JSON and multipart)
	var req usecases.UpdateProfileRequest

	// Try to parse as form data first (for file uploads)
	firstName := c.FormValue("first_name")
	lastName := c.FormValue("last_name")
	email := c.FormValue("email")
	company := c.FormValue("company")
	address := c.FormValue("address")
	specialization := c.FormValue("specialization")
	experienceYears := c.FormValue("experience_years")
	about := c.FormValue("about")
	skillsStr := c.FormValue("skills")
	certificationsStr := c.FormValue("certifications")

	// If form values exist, use them; otherwise try JSON parsing
	if firstName != "" || lastName != "" || email != "" || company != "" || address != "" || specialization != "" || experienceYears != "" || about != "" || skillsStr != "" || certificationsStr != "" {
		// Parse form data
		req.FirstName = firstName
		req.LastName = lastName
		req.Email = email
		req.Company = company
		req.Address = address
		req.Specialization = specialization
		req.About = about
		if experienceYears != "" {
			fmt.Sscanf(experienceYears, "%d", &req.ExperienceYears)
		}

		// Parse skills array (JSON format: ["Python","Bash"])
		if skillsStr != "" {
			var skills []string
			if err := json.Unmarshal([]byte(skillsStr), &skills); err == nil {
				req.Skills = skills
			} else {
				log.Printf("[WARN] UpdateProfile - Failed to parse skills: %v", err)
			}
		}

		// Parse certifications array (JSON format: ["OSCP","CEH"])
		if certificationsStr != "" {
			var certifications []string
			if err := json.Unmarshal([]byte(certificationsStr), &certifications); err == nil {
				req.Certifications = certifications
			} else {
				log.Printf("[WARN] UpdateProfile - Failed to parse certifications: %v", err)
			}
		}
	} else {
		// Fallback to JSON parsing
		if err := c.BodyParser(&req); err != nil {
			log.Printf("[ERROR] UpdateProfile - Invalid request format: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Success: false,
				Message: "Bad request",
				Error:   fmt.Sprintf("invalid request: %v", err),
			})
		}
	}

	// Handle resume file upload for pentesters
	file, err := c.FormFile("resume")
	if err == nil && file != nil {
		// File provided - validate and upload
		if file.Size > 5*1024*1024 {
			return c.Status(fiber.StatusBadRequest).JSON(response{
				Success: false,
				Message: "File too large",
				Error:   "resume file must be less than 5MB",
			})
		}

		// Open uploaded file
		src, err := file.Open()
		if err != nil {
			log.Printf("[ERROR] UpdateProfile - Failed to open file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Success: false,
				Message: "Internal server error",
				Error:   err.Error(),
			})
		}
		defer src.Close()

		// Create temporary file
		tmpFile, err := os.CreateTemp("", "resume-*")
		if err != nil {
			log.Printf("[ERROR] UpdateProfile - Failed to create temp file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Success: false,
				Message: "Internal server error",
				Error:   err.Error(),
			})
		}
		defer os.Remove(tmpFile.Name())

		// Copy uploaded file to temp file
		if _, err := io.Copy(tmpFile, src); err != nil {
			log.Printf("[ERROR] UpdateProfile - Failed to copy file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Success: false,
				Message: "Internal server error",
				Error:   err.Error(),
			})
		}

		// Reset file pointer
		tmpFile.Seek(0, 0)

		// Save file to storage
		filePath, err := umh.fileService.UploadResume(tmpFile, file.Filename)
		if err != nil {
			log.Printf("[ERROR] UpdateProfile - Failed to save file: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Success: false,
				Message: "Failed to save file",
				Error:   err.Error(),
			})
		}

		// Update pentester profile with file path
		_, err = umh.pentesterUseCase.UpdateResumeFilePath(claims.UserID, filePath)
		if err != nil {
			log.Printf("[ERROR] UpdateProfile - Failed to update resume path: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(response{
				Success: false,
				Message: "Failed to update resume",
				Error:   err.Error(),
			})
		}
		log.Printf("[INFO] UpdateProfile - Resume uploaded successfully for user: %s", claims.UserID)
	}

	// Update profile fields
	err = umh.userManagementUseCase.UpdateUserProfile(claims.UserID, &req)
	if err != nil {
		log.Printf("[ERROR] UpdateProfile - Failed to update profile: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   err.Error(),
		})
	}

	userDetails, err := umh.userManagementUseCase.GetUserProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] UpdateProfile - Failed to retrieve updated profile: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Success: false,
			Message: "Internal server error",
			Error:   "failed to retrieve updated profile",
		})
	}

	log.Printf("[SUCCESS] UpdateProfile - User profile updated: user_id=%s, role=%s", claims.UserID, claims.Role)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Profile updated successfully",
		Data:    userDetails,
	})
}

// AdminUpdateUserProfile updates a user's profile (admin only)
// POST /api/v1/admin/user/:user_id/profile
func (umh *UserManagementHandler) AdminUpdateUserProfile(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/admin/user/:user_id/profile - Admin updating user profile")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] AdminUpdateUserProfile - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := umh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] AdminUpdateUserProfile - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	if claims.Role != "admin" {
		log.Printf("[ERROR] AdminUpdateUserProfile - Non-admin user attempted access, user_id: %s, role: %s", claims.UserID, claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Success: false,
			Message: "Forbidden",
			Error:   "admin access required",
		})
	}

	userID := c.Params("user_id")
	if userID == "" {
		log.Printf("[ERROR] AdminUpdateUserProfile - Missing user_id parameter")
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   "missing user_id",
		})
	}

	var req usecases.UpdateProfileRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] AdminUpdateUserProfile - Invalid request format: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   fmt.Sprintf("invalid request: %v", err),
		})
	}

	err = umh.userManagementUseCase.AdminUpdateUserProfile(userID, &req)
	if err != nil {
		log.Printf("[ERROR] AdminUpdateUserProfile - Failed to update profile: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Bad request",
			Error:   err.Error(),
		})
	}

	userDetails, err := umh.userManagementUseCase.GetUserByID(userID)
	if err != nil {
		log.Printf("[ERROR] AdminUpdateUserProfile - Failed to retrieve updated profile: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Success: false,
			Message: "Internal server error",
			Error:   "failed to retrieve updated profile",
		})
	}

	log.Printf("[SUCCESS] AdminUpdateUserProfile - User profile updated by admin: user_id=%s, admin_id=%s, role=%s", userID, claims.UserID, userDetails.User.Role)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "User profile updated successfully",
		Data:    userDetails,
	})
}

// GetUserResume retrieves the authenticated user's resume file
// GET /api/v1/profile/resume
func (umh *UserManagementHandler) GetUserResume(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/profile/resume - User retrieving their resume")

	// Extract and validate token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] GetUserResume - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := umh.tokenManager.VerifyToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] GetUserResume - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "invalid token",
		})
	}

	// Get user's profile to fetch resume file path
	userDetails, err := umh.userManagementUseCase.GetUserProfile(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] GetUserResume - Failed to get user profile: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "User profile not found",
			Error:   err.Error(),
		})
	}

	// Get the resume file path based on user role
	var resumeFilePath string
	if userDetails.User.Role == "pentester" && userDetails.Profile != nil {
		if pentesterProfile, ok := userDetails.Profile.(*domain.Pentester); ok {
			resumeFilePath = pentesterProfile.ResumeFilePath
		}
	}

	if resumeFilePath == "" {
		log.Printf("[ERROR] GetUserResume - No resume found for user: %s", claims.UserID)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "Resume not found",
			Error:   "user has not uploaded a resume",
		})
	}

	// Check if file exists
	if !umh.fileService.FileExists(resumeFilePath) {
		log.Printf("[ERROR] GetUserResume - File not found at path: %s", resumeFilePath)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "File not found",
			Error:   "resume file does not exist on server",
		})
	}

	// Open and read the file
	file, err := os.Open(resumeFilePath)
	if err != nil {
		log.Printf("[ERROR] GetUserResume - Failed to open file: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Success: false,
			Message: "Internal server error",
			Error:   "failed to read resume file",
		})
	}
	defer file.Close()

	// Read entire file into memory
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Printf("[ERROR] GetUserResume - Failed to read file content: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(response{
			Success: false,
			Message: "Internal server error",
			Error:   "failed to read file content",
		})
	}

	// Set response headers
	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="resume_%s.pdf"`, claims.UserID))
	c.Set("Content-Length", fmt.Sprintf("%d", len(fileBytes)))
	c.Set("Cache-Control", "no-cache, must-revalidate")
	c.Set("Pragma", "no-cache")

	log.Printf("[SUCCESS] GetUserResume - Resume retrieved successfully for user: %s, size: %d bytes", claims.UserID, len(fileBytes))

	// Send file content
	return c.Status(fiber.StatusOK).Send(fileBytes)
}
