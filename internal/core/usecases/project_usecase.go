package usecases

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pentsecops/backend/internal/core/domain"
)

// ProjectUseCase handles project business logic
type ProjectUseCase struct {
	projectRepo       domain.ProjectRepository
	projectMemberRepo domain.ProjectMemberRepository
	userRepo          domain.UserRepository
}

// NewProjectUseCase creates a new project use case
func NewProjectUseCase(
	projectRepo domain.ProjectRepository,
	projectMemberRepo domain.ProjectMemberRepository,
	userRepo domain.UserRepository,
) *ProjectUseCase {
	return &ProjectUseCase{
		projectRepo:       projectRepo,
		projectMemberRepo: projectMemberRepo,
		userRepo:          userRepo,
	}
}

// CreateProjectRequest contains the data for creating a project
type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

// CreateProject creates a new project
func (uc *ProjectUseCase) CreateProject(req *CreateProjectRequest, adminID string) (*domain.ProjectResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("project name is required")
	}

	now := time.Now().UTC()

	project := &domain.Project{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		CreatedBy:   adminID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Parse start_date if provided
	if req.StartDate != nil && *req.StartDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format (use YYYY-MM-DD): %w", err)
		}
		project.StartDate = &parsedDate
	}

	// Parse end_date if provided
	if req.EndDate != nil && *req.EndDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format (use YYYY-MM-DD): %w", err)
		}
		project.EndDate = &parsedDate
	}

	if err := uc.projectRepo.CreateProject(project); err != nil {
		return nil, err
	}

	return project.ToResponse(), nil
}

// UpdateProjectRequest contains the data for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// UpdateProject updates an existing project
func (uc *ProjectUseCase) UpdateProject(projectID string, req *UpdateProjectRequest) (*domain.ProjectResponse, error) {
	project, err := uc.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.Status != "" {
		project.Status = req.Status
	}

	if err := uc.projectRepo.UpdateProject(project); err != nil {
		return nil, err
	}

	return project.ToResponse(), nil
}

// GetProject retrieves a project by ID
func (uc *ProjectUseCase) GetProject(projectID string) (*domain.ProjectResponse, error) {
	project, err := uc.projectRepo.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}
	return project.ToResponse(), nil
}

// GetProjectsByAdmin retrieves all projects created by an admin
func (uc *ProjectUseCase) GetProjectsByAdmin(adminID string) ([]*domain.ProjectResponse, error) {
	projects, err := uc.projectRepo.GetProjectsByAdmin(adminID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.ProjectResponse
	for _, p := range projects {
		responses = append(responses, p.ToResponse())
	}
	return responses, nil
}

// DeleteProject deletes a project
func (uc *ProjectUseCase) DeleteProject(projectID string) error {
	return uc.projectRepo.DeleteProject(projectID)
}

// AddMemberToProjectRequest contains the data for adding a member to a project
type AddMemberToProjectRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // pentester or stakeholder
}

// AddMemberToProject adds a member to a project
func (uc *ProjectUseCase) AddMemberToProject(projectID string, req *AddMemberToProjectRequest) (*domain.ProjectMemberResponse, error) {
	// Verify project exists
	if _, err := uc.projectRepo.GetProjectByID(projectID); err != nil {
		return nil, err
	}

	// Verify user exists
	if _, err := uc.userRepo.GetUserByID(req.UserID); err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify role is valid
	if req.Role != "pentester" && req.Role != "stakeholder" {
		return nil, fmt.Errorf("invalid role: must be 'pentester' or 'stakeholder'")
	}

	member := &domain.ProjectMember{
		ID:        uuid.New().String(),
		ProjectID: projectID,
		UserID:    req.UserID,
		Role:      req.Role,
	}

	if err := uc.projectMemberRepo.AddMember(member); err != nil {
		return nil, err
	}

	return member.ToResponse(), nil
}

// GetProjectMembers retrieves all members of a project
func (uc *ProjectUseCase) GetProjectMembers(projectID string) ([]*domain.ProjectMemberResponse, error) {
	members, err := uc.projectMemberRepo.GetProjectMembers(projectID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.ProjectMemberResponse
	for _, m := range members {
		responses = append(responses, m.ToResponse())
	}
	return responses, nil
}

// RemoveMemberFromProject removes a member from a project
func (uc *ProjectUseCase) RemoveMemberFromProject(projectID, userID string) error {
	return uc.projectMemberRepo.RemoveMember(projectID, userID)
}

// GetUserProjects retrieves all projects a user is a member of
func (uc *ProjectUseCase) GetUserProjects(userID string) ([]*domain.ProjectMemberResponse, error) {
	members, err := uc.projectMemberRepo.GetUserProjects(userID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.ProjectMemberResponse
	for _, m := range members {
		responses = append(responses, m.ToResponse())
	}
	return responses, nil
}
