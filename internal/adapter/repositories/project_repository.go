package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// ProjectRepositoryImpl implements the ProjectRepository interface
type ProjectRepositoryImpl struct {
	db *sql.DB
}

// NewProjectRepository creates a new project repository
func NewProjectRepository(db *sql.DB) domain.ProjectRepository {
	return &ProjectRepositoryImpl{db: db}
}

// CreateProject creates a new project in the database
func (r *ProjectRepositoryImpl) CreateProject(project *domain.Project) error {
	query := `
		INSERT INTO projects (id, name, description, status, start_date, end_date, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Exec(query,
		project.ID,
		project.Name,
		project.Description,
		project.Status,
		project.StartDate,
		project.EndDate,
		project.CreatedBy,
		project.CreatedAt,
		project.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

// GetProjectByID retrieves a project by ID
func (r *ProjectRepositoryImpl) GetProjectByID(id string) (*domain.Project, error) {
	query := `
		SELECT id, name, description, status, start_date, end_date, created_by, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	project := &domain.Project{}
	err := r.db.QueryRow(query, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Status,
		&project.StartDate,
		&project.EndDate,
		&project.CreatedBy,
		&project.CreatedAt,
		&project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// GetAllProjects retrieves all projects
func (r *ProjectRepositoryImpl) GetAllProjects() ([]*domain.Project, error) {
	query := `
		SELECT id, name, description, status, start_date, end_date, created_by, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		project := &domain.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Status,
			&project.StartDate,
			&project.EndDate,
			&project.CreatedBy,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// GetProjectsByAdmin retrieves all projects created by an admin
func (r *ProjectRepositoryImpl) GetProjectsByAdmin(adminID string) ([]*domain.Project, error) {
	query := `
		SELECT id, name, description, status, start_date, end_date, created_by, created_at, updated_at
		FROM projects
		WHERE created_by = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, adminID)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		project := &domain.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Status,
			&project.StartDate,
			&project.EndDate,
			&project.CreatedBy,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	return projects, nil
}

// UpdateProject updates a project
func (r *ProjectRepositoryImpl) UpdateProject(project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = $1, description = $2, status = $3, start_date = $4, end_date = $5, updated_at = $6
		WHERE id = $7
	`

	_, err := r.db.Exec(query,
		project.Name,
		project.Description,
		project.Status,
		project.StartDate,
		project.EndDate,
		time.Now().UTC(),
		project.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	return nil
}

// DeleteProject deletes a project
func (r *ProjectRepositoryImpl) DeleteProject(id string) error {
	query := `DELETE FROM projects WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	return nil
}

// ProjectMemberRepositoryImpl implements the ProjectMemberRepository interface
type ProjectMemberRepositoryImpl struct {
	db *sql.DB
}

// NewProjectMemberRepository creates a new project member repository
func NewProjectMemberRepository(db *sql.DB) domain.ProjectMemberRepository {
	return &ProjectMemberRepositoryImpl{db: db}
}

// AddMember adds a member to a project
func (r *ProjectMemberRepositoryImpl) AddMember(member *domain.ProjectMember) error {
	query := `
		INSERT INTO project_members (id, project_id, user_id, role, added_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(query,
		member.ID,
		member.ProjectID,
		member.UserID,
		member.Role,
		time.Now().UTC(),
	)

	if err != nil {
		return fmt.Errorf("failed to add project member: %w", err)
	}

	return nil
}

// GetProjectMembers retrieves all members of a project
func (r *ProjectMemberRepositoryImpl) GetProjectMembers(projectID string) ([]*domain.ProjectMember, error) {
	query := `
		SELECT id, project_id, user_id, role, added_at
		FROM project_members
		WHERE project_id = $1
		ORDER BY added_at DESC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project members: %w", err)
	}
	defer rows.Close()

	var members []*domain.ProjectMember
	for rows.Next() {
		member := &domain.ProjectMember{}
		err := rows.Scan(
			&member.ID,
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.AddedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

// GetUserProjects retrieves all projects a user is a member of with full project details
func (r *ProjectMemberRepositoryImpl) GetUserProjects(userID string) ([]*domain.ProjectMember, error) {
	query := `
		SELECT 
			pm.id, pm.project_id, pm.user_id, pm.role, pm.added_at,
			p.id, p.name, p.description, p.status, p.start_date, p.end_date, p.created_by, p.created_at, p.updated_at
		FROM project_members pm
		JOIN projects p ON pm.project_id = p.id
		WHERE pm.user_id = $1
		ORDER BY pm.added_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user projects: %w", err)
	}
	defer rows.Close()

	var members []*domain.ProjectMember
	for rows.Next() {
		member := &domain.ProjectMember{}
		project := &domain.Project{}
		err := rows.Scan(
			&member.ID,
			&member.ProjectID,
			&member.UserID,
			&member.Role,
			&member.AddedAt,
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Status,
			&project.StartDate,
			&project.EndDate,
			&project.CreatedBy,
			&project.CreatedAt,
			&project.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan project member: %w", err)
		}
		member.Project = project
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return members, nil
}

// GetMemberRole retrieves the role of a user in a project
func (r *ProjectMemberRepositoryImpl) GetMemberRole(projectID, userID string) (string, error) {
	query := `
		SELECT role FROM project_members
		WHERE project_id = $1 AND user_id = $2
	`

	var role string
	err := r.db.QueryRow(query, projectID, userID).Scan(&role)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("member not found in project")
		}
		return "", fmt.Errorf("failed to get member role: %w", err)
	}

	return role, nil
}

// RemoveMember removes a member from a project
func (r *ProjectMemberRepositoryImpl) RemoveMember(projectID, userID string) error {
	query := `DELETE FROM project_members WHERE project_id = $1 AND user_id = $2`

	_, err := r.db.Exec(query, projectID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove project member: %w", err)
	}

	return nil
}

// IsMemberOfProject checks if a user is a member of a project
func (r *ProjectMemberRepositoryImpl) IsMemberOfProject(projectID, userID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM project_members
			WHERE project_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(query, projectID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check project membership: %w", err)
	}

	return exists, nil
}
