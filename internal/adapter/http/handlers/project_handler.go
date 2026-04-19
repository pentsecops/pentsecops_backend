package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/usecases"
)

// ProjectHandler handles project HTTP requests
type ProjectHandler struct {
	projectUseCase *usecases.ProjectUseCase
	authUseCase    *usecases.AuthUseCase
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(projectUseCase *usecases.ProjectUseCase, authUseCase *usecases.AuthUseCase) *ProjectHandler {
	return &ProjectHandler{
		projectUseCase: projectUseCase,
		authUseCase:    authUseCase,
	}
}

// CreateProject handles project creation endpoint POST /api/v1/projects
func (ph *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/projects - Project creation attempt")

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		log.Printf("[ERROR] CreateProject - Missing authentication token")
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		log.Printf("[ERROR] CreateProject - Token validation failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	var req usecases.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] CreateProject - Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	project, err := ph.projectUseCase.CreateProject(&req, claims.UserID)
	if err != nil {
		log.Printf("[ERROR] CreateProject - Failed to create project: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to create project",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] CreateProject - Project created successfully: %s", project.ID)
	return c.Status(fiber.StatusCreated).JSON(response{
		Success: true,
		Message: "Project created successfully",
		Data:    project,
	})
}

// GetProject handles project retrieval endpoint GET /api/v1/projects/:id
func (ph *ProjectHandler) GetProject(c *fiber.Ctx) error {
	projectID := c.Params("id")
	log.Printf("[API] GET /api/v1/projects/%s - Project retrieval attempt", projectID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	project, err := ph.projectUseCase.GetProject(projectID)
	if err != nil {
		log.Printf("[ERROR] GetProject - Project not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "Project not found",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetProject - Project retrieved: %s", projectID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Project retrieved successfully",
		Data:    project,
	})
}

// GetProjectsByAdmin handles admin projects retrieval endpoint GET /api/v1/admin/projects
func (ph *ProjectHandler) GetProjectsByAdmin(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/admin/projects - Admin projects retrieval attempt")

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	projects, err := ph.projectUseCase.GetProjectsByAdmin(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] GetProjectsByAdmin - Failed to retrieve projects: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to retrieve projects",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetProjectsByAdmin - Retrieved %d projects", len(projects))
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Projects retrieved successfully",
		Data:    projects,
	})
}

// UpdateProject handles project update endpoint PUT /api/v1/projects/:id
func (ph *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	projectID := c.Params("id")
	log.Printf("[API] PUT /api/v1/projects/%s - Project update attempt", projectID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	var req usecases.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] UpdateProject - Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	project, err := ph.projectUseCase.UpdateProject(projectID, &req)
	if err != nil {
		log.Printf("[ERROR] UpdateProject - Failed to update project: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to update project",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] UpdateProject - Project updated: %s", projectID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Project updated successfully",
		Data:    project,
	})
}

// DeleteProject handles project deletion endpoint DELETE /api/v1/projects/:id
func (ph *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	projectID := c.Params("id")
	log.Printf("[API] DELETE /api/v1/projects/%s - Project deletion attempt", projectID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	if err := ph.projectUseCase.DeleteProject(projectID); err != nil {
		log.Printf("[ERROR] DeleteProject - Failed to delete project: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to delete project",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] DeleteProject - Project deleted: %s", projectID)
	return c.Status(fiber.StatusNoContent).JSON(response{
		Success: true,
		Message: "Project deleted successfully",
	})
}

// AddMemberToProject handles adding member to project endpoint POST /api/v1/projects/:id/members
func (ph *ProjectHandler) AddMemberToProject(c *fiber.Ctx) error {
	projectID := c.Params("id")
	log.Printf("[API] POST /api/v1/projects/%s/members - Add member attempt", projectID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	var req usecases.AddMemberToProjectRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] AddMemberToProject - Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	member, err := ph.projectUseCase.AddMemberToProject(projectID, &req)
	if err != nil {
		log.Printf("[ERROR] AddMemberToProject - Failed to add member: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to add member",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] AddMemberToProject - Member added to project: %s", projectID)
	return c.Status(fiber.StatusCreated).JSON(response{
		Success: true,
		Message: "Member added successfully",
		Data:    member,
	})
}

// GetProjectMembers handles project members retrieval endpoint GET /api/v1/projects/:id/members
func (ph *ProjectHandler) GetProjectMembers(c *fiber.Ctx) error {
	projectID := c.Params("id")
	log.Printf("[API] GET /api/v1/projects/%s/members - Project members retrieval attempt", projectID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	members, err := ph.projectUseCase.GetProjectMembers(projectID)
	if err != nil {
		log.Printf("[ERROR] GetProjectMembers - Failed to retrieve members: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to retrieve members",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetProjectMembers - Retrieved %d members for project: %s", len(members), projectID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Members retrieved successfully",
		Data:    members,
	})
}

// RemoveMemberFromProject handles member removal endpoint DELETE /api/v1/projects/:id/members/:userID
func (ph *ProjectHandler) RemoveMemberFromProject(c *fiber.Ctx) error {
	projectID := c.Params("id")
	userID := c.Params("userID")
	log.Printf("[API] DELETE /api/v1/projects/%s/members/%s - Remove member attempt", projectID, userID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	if err := ph.projectUseCase.RemoveMemberFromProject(projectID, userID); err != nil {
		log.Printf("[ERROR] RemoveMemberFromProject - Failed to remove member: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to remove member",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] RemoveMemberFromProject - Member removed from project: %s", projectID)
	return c.Status(fiber.StatusNoContent).JSON(response{
		Success: true,
		Message: "Member removed successfully",
	})
}

// GetUserProjects handles user projects retrieval endpoint GET /api/v1/user/projects
func (ph *ProjectHandler) GetUserProjects(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/user/projects - User projects retrieval attempt")

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := ph.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	projects, err := ph.projectUseCase.GetUserProjects(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] GetUserProjects - Failed to retrieve user projects: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to retrieve projects",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetUserProjects - Retrieved %d projects for user: %s", len(projects), claims.UserID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Projects retrieved successfully",
		Data:    projects,
	})
}
