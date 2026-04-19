package usecases

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pentsecops/backend/internal/core/domain"
)

// TaskUseCase handles task business logic
type TaskUseCase struct {
	taskRepo      domain.TaskRepository
	projectRepo   domain.ProjectRepository
	pentesterRepo domain.PentesterRepository
}

// NewTaskUseCase creates a new task use case
func NewTaskUseCase(
	taskRepo domain.TaskRepository,
	projectRepo domain.ProjectRepository,
	pentesterRepo domain.PentesterRepository,
) *TaskUseCase {
	return &TaskUseCase{
		taskRepo:      taskRepo,
		projectRepo:   projectRepo,
		pentesterRepo: pentesterRepo,
	}
}

// CreateTaskRequest contains the data for creating a task
type CreateTaskRequest struct {
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Deadline    *time.Time `json:"deadline"`
}

// CreateTask creates a new task
func (uc *TaskUseCase) CreateTask(req *CreateTaskRequest) (*domain.TaskResponse, error) {
	if req.ProjectID == "" {
		return nil, fmt.Errorf("project_id is required")
	}
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if req.Priority == "" {
		req.Priority = "medium"
	}

	// Verify project exists
	if _, err := uc.projectRepo.GetProjectByID(req.ProjectID); err != nil {
		return nil, fmt.Errorf("project not found: %w", err)
	}

	now := time.Now().UTC()

	task := &domain.Task{
		ID:          uuid.New().String(),
		ProjectID:   req.ProjectID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      "assigned",
		Deadline:    req.Deadline,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.taskRepo.CreateTask(task); err != nil {
		return nil, err
	}

	return task.ToResponse(), nil
}

// UpdateTaskRequest contains the data for updating a task
type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	Deadline    *time.Time `json:"deadline"`
	AssignedTo  *string    `json:"assigned_to"`
}

// UpdateTask updates an existing task
func (uc *TaskUseCase) UpdateTask(taskID string, req *UpdateTaskRequest) (*domain.TaskResponse, error) {
	task, err := uc.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		task.Title = req.Title
	}
	if req.Description != "" {
		task.Description = req.Description
	}
	if req.Priority != "" {
		task.Priority = req.Priority
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Deadline != nil {
		task.Deadline = req.Deadline
	}
	if req.AssignedTo != nil {
		task.AssignedTo = req.AssignedTo
	}

	if err := uc.taskRepo.UpdateTask(task); err != nil {
		return nil, err
	}

	return task.ToResponse(), nil
}

// GetTask retrieves a task by ID
func (uc *TaskUseCase) GetTask(taskID string) (*domain.TaskResponse, error) {
	task, err := uc.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	return task.ToResponse(), nil
}

// GetTasksByProject retrieves all tasks for a project
func (uc *TaskUseCase) GetTasksByProject(projectID string) ([]*domain.TaskResponse, error) {
	tasks, err := uc.taskRepo.GetTasksByProject(projectID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.TaskResponse
	for _, t := range tasks {
		responses = append(responses, t.ToResponse())
	}
	return responses, nil
}

// GetTasksByPentester retrieves all tasks assigned to a pentester
func (uc *TaskUseCase) GetTasksByPentester(pentesterID string) ([]*domain.TaskResponse, error) {
	tasks, err := uc.taskRepo.GetTasksByPentester(pentesterID)
	if err != nil {
		return nil, err
	}

	var responses []*domain.TaskResponse
	for _, t := range tasks {
		responses = append(responses, t.ToResponse())
	}
	return responses, nil
}

// AssignTaskRequest contains the data for assigning a task
type AssignTaskRequest struct {
	PentesterID string `json:"pentester_id"`
}

// AssignTask assigns a task to a pentester
func (uc *TaskUseCase) AssignTask(taskID string, req *AssignTaskRequest) (*domain.TaskResponse, error) {
	task, err := uc.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	// Verify pentester exists
	if _, err := uc.pentesterRepo.GetPentesterByID(req.PentesterID); err != nil {
		return nil, fmt.Errorf("pentester not found: %w", err)
	}

	task.AssignedTo = &req.PentesterID
	task.Status = "in_progress"

	if err := uc.taskRepo.UpdateTask(task); err != nil {
		return nil, err
	}

	return task.ToResponse(), nil
}

// DeleteTask deletes a task
func (uc *TaskUseCase) DeleteTask(taskID string) error {
	return uc.taskRepo.DeleteTask(taskID)
}

// CloseTask closes a task
func (uc *TaskUseCase) CloseTask(taskID string) (*domain.TaskResponse, error) {
	task, err := uc.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	task.Status = "closed"
	if err := uc.taskRepo.UpdateTask(task); err != nil {
		return nil, err
	}

	return task.ToResponse(), nil
}
