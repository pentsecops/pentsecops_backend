package domain

import "time"

// Project defines a security engagement
type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"` // active, paused, completed, archived
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ProjectResponse is the response model for project
type ProjectResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts Project to ProjectResponse
func (p *Project) ToResponse() *ProjectResponse {
	return &ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
		StartDate:   p.StartDate,
		EndDate:     p.EndDate,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ProjectMember defines a member assigned to a project
type ProjectMember struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"` // pentester, stakeholder
	AddedAt   time.Time `json:"added_at"`
	Project   *Project  `json:"project,omitempty"` // Full project details
}

// ProjectMemberResponse is the response model for project member
type ProjectMemberResponse struct {
	ID        string           `json:"id"`
	ProjectID string           `json:"project_id"`
	UserID    string           `json:"user_id"`
	Role      string           `json:"role"`
	AddedAt   time.Time        `json:"added_at"`
	Project   *ProjectResponse `json:"project,omitempty"` // Full project details
}

// ToResponse converts ProjectMember to ProjectMemberResponse
func (pm *ProjectMember) ToResponse() *ProjectMemberResponse {
	var projectResp *ProjectResponse
	if pm.Project != nil {
		projectResp = pm.Project.ToResponse()
	}
	return &ProjectMemberResponse{
		ID:        pm.ID,
		ProjectID: pm.ProjectID,
		UserID:    pm.UserID,
		Role:      pm.Role,
		AddedAt:   pm.AddedAt,
		Project:   projectResp,
	}
}
