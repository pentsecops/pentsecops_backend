package handlers

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/services"
)

// TaskHandler handles task HTTP requests
type TaskHandler struct {
	taskUseCase     *usecases.TaskUseCase
	authUseCase     *usecases.AuthUseCase
	fileService     *services.FileService
	activityUseCase *usecases.ActivityUseCase
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskUseCase *usecases.TaskUseCase, authUseCase *usecases.AuthUseCase) *TaskHandler {
	return &TaskHandler{
		taskUseCase: taskUseCase,
		authUseCase: authUseCase,
	}
}

// NewTaskHandlerWithServices creates a new task handler with file service and activity use case
func NewTaskHandlerWithServices(
	taskUseCase *usecases.TaskUseCase,
	authUseCase *usecases.AuthUseCase,
	fileService *services.FileService,
	activityUseCase *usecases.ActivityUseCase,
) *TaskHandler {
	return &TaskHandler{
		taskUseCase:     taskUseCase,
		authUseCase:     authUseCase,
		fileService:     fileService,
		activityUseCase: activityUseCase,
	}
}

// CreateTask handles task creation endpoint POST /api/v1/tasks
func (th *TaskHandler) CreateTask(c *fiber.Ctx) error {
	log.Printf("[API] POST /api/v1/tasks - Task creation attempt")

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	var req usecases.CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] CreateTask - Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	task, err := th.taskUseCase.CreateTask(&req)
	if err != nil {
		log.Printf("[ERROR] CreateTask - Failed to create task: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to create task",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] CreateTask - Task created: %s", task.ID)
	return c.Status(fiber.StatusCreated).JSON(response{
		Success: true,
		Message: "Task created successfully",
		Data:    task,
	})
}

// GetTask handles task retrieval endpoint GET /api/v1/tasks/:id
func (th *TaskHandler) GetTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	log.Printf("[API] GET /api/v1/tasks/%s - Task retrieval attempt", taskID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	task, err := th.taskUseCase.GetTask(taskID)
	if err != nil {
		log.Printf("[ERROR] GetTask - Task not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(response{
			Success: false,
			Message: "Task not found",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetTask - Task retrieved: %s", taskID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Task retrieved successfully",
		Data:    task,
	})
}

// GetTasksByProject handles project tasks retrieval endpoint GET /api/v1/projects/:projectID/tasks
func (th *TaskHandler) GetTasksByProject(c *fiber.Ctx) error {
	projectID := c.Params("projectID")
	log.Printf("[API] GET /api/v1/projects/%s/tasks - Project tasks retrieval attempt", projectID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	tasks, err := th.taskUseCase.GetTasksByProject(projectID)
	if err != nil {
		log.Printf("[ERROR] GetTasksByProject - Failed to retrieve tasks: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to retrieve tasks",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetTasksByProject - Retrieved %d tasks for project: %s", len(tasks), projectID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Tasks retrieved successfully",
		Data:    tasks,
	})
}

// GetTasksByPentester handles pentester tasks retrieval endpoint GET /api/v1/pentester/tasks
func (th *TaskHandler) GetTasksByPentester(c *fiber.Ctx) error {
	log.Printf("[API] GET /api/v1/pentester/tasks - Pentester tasks retrieval attempt")

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	claims, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	tasks, err := th.taskUseCase.GetTasksByPentester(claims.UserID)
	if err != nil {
		log.Printf("[ERROR] GetTasksByPentester - Failed to retrieve tasks: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to retrieve tasks",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetTasksByPentester - Retrieved %d tasks for pentester: %s", len(tasks), claims.UserID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Tasks retrieved successfully",
		Data:    tasks,
	})
}

// UpdateTask handles task update endpoint PUT /api/v1/tasks/:id
func (th *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	log.Printf("[API] PUT /api/v1/tasks/%s - Task update attempt", taskID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	var req usecases.UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] UpdateTask - Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	task, err := th.taskUseCase.UpdateTask(taskID, &req)
	if err != nil {
		log.Printf("[ERROR] UpdateTask - Failed to update task: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to update task",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] UpdateTask - Task updated: %s", taskID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Task updated successfully",
		Data:    task,
	})
}

// AssignTask handles task assignment endpoint POST /api/v1/tasks/:id/assign
func (th *TaskHandler) AssignTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	log.Printf("[API] POST /api/v1/tasks/%s/assign - Task assignment attempt", taskID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	var req usecases.AssignTaskRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] AssignTask - Invalid request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	task, err := th.taskUseCase.AssignTask(taskID, &req)
	if err != nil {
		log.Printf("[ERROR] AssignTask - Failed to assign task: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to assign task",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] AssignTask - Task assigned: %s", taskID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Task assigned successfully",
		Data:    task,
	})
}

// DeleteTask handles task deletion endpoint DELETE /api/v1/tasks/:id
func (th *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	log.Printf("[API] DELETE /api/v1/tasks/%s - Task deletion attempt", taskID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	if err := th.taskUseCase.DeleteTask(taskID); err != nil {
		log.Printf("[ERROR] DeleteTask - Failed to delete task: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to delete task",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] DeleteTask - Task deleted: %s", taskID)
	return c.Status(fiber.StatusNoContent).JSON(response{
		Success: true,
		Message: "Task deleted successfully",
	})
}

// CloseTask handles task closing endpoint POST /api/v1/tasks/:id/close
func (th *TaskHandler) CloseTask(c *fiber.Ctx) error {
	taskID := c.Params("id")
	log.Printf("[API] POST /api/v1/tasks/%s/close - Task closing attempt", taskID)

	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	task, err := th.taskUseCase.CloseTask(taskID)
	if err != nil {
		log.Printf("[ERROR] CloseTask - Failed to close task: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to close task",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] CloseTask - Task closed: %s", taskID)
	return c.Status(fiber.StatusOK).JSON(response{
		Success: true,
		Message: "Task closed successfully",
		Data:    task,
	})
}
