package handlers

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/services"
)

// TaskSubmissionHandler handles task submission HTTP requests
type TaskSubmissionHandler struct {
	submissionUseCase *usecases.TaskSubmissionUseCase
	authUseCase       *usecases.AuthUseCase
	fileService       *services.FileService
	attachmentRepo    domain.TaskSubmissionAttachmentRepository
}

// NewTaskSubmissionHandler creates a new task submission handler
func NewTaskSubmissionHandler(
	submissionUseCase *usecases.TaskSubmissionUseCase,
	authUseCase *usecases.AuthUseCase,
	fileService *services.FileService,
) *TaskSubmissionHandler {
	return &TaskSubmissionHandler{
		submissionUseCase: submissionUseCase,
		authUseCase:       authUseCase,
		fileService:       fileService,
	}
}

// SetAttachmentRepository sets the attachment repository for the handler
func (th *TaskSubmissionHandler) SetAttachmentRepository(repo domain.TaskSubmissionAttachmentRepository) {
	th.attachmentRepo = repo
}

// SubmitTask handles POST /api/v1/task-submission/:task_id
func (th *TaskSubmissionHandler) SubmitTask(c *fiber.Ctx) error {
	taskID := c.Params("task_id")
	log.Printf("[API] POST /api/v1/task-submission/%s - Task submission attempt", taskID)

	// Extract token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	// Validate token
	claims, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	// Only pentesters can submit tasks
	if claims.Role != "pentester" {
		log.Printf("[ERROR] SubmitTask - Invalid role: %s", claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"success": false,
			"message": "Forbidden",
			"error":   "only pentesters can submit tasks",
		})
	}

	// Get notes from form
	notes := c.FormValue("notes", "")

	// Get file attachments from multipart form
	form, err := c.MultipartForm()
	if err != nil && err.Error() != "request Content-Type is not multipart/form-data" {
		log.Printf("[ERROR] SubmitTask - Failed to parse multipart form: %v", err)
	}

	var attachmentList []usecases.AttachmentData
	var uploadedPaths []string

	if form != nil && form.File != nil && len(form.File["attachments"]) > 0 {
		files := form.File["attachments"]
		attachmentNotes := form.Value["attachment_notes"]

		log.Printf("[INFO] SubmitTask - Processing %d attachments", len(files))

		for idx, file := range files {
			// Validate file size (max 50MB)
			if file.Size > 50*1024*1024 {
				// Clean up already uploaded files
				for _, path := range uploadedPaths {
					os.Remove(path)
				}
				log.Printf("[ERROR] SubmitTask - File too large: %s", file.Filename)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "File too large",
					"error":   fmt.Sprintf("attachment %d must be less than 50MB", idx+1),
				})
			}

			// Open uploaded file
			src, err := file.Open()
			if err != nil {
				log.Printf("[ERROR] SubmitTask - Failed to open file: %v", err)
				for _, path := range uploadedPaths {
					os.Remove(path)
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Internal server error",
					"error":   "failed to process file",
				})
			}
			defer src.Close()

			// Create temp file
			tmpFile, err := os.CreateTemp("", "submission-*")
			if err != nil {
				log.Printf("[ERROR] SubmitTask - Failed to create temp file: %v", err)
				for _, path := range uploadedPaths {
					os.Remove(path)
				}
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Internal server error",
					"error":   "failed to create temp file",
				})
			}

			// Copy file
			bytesRead, err := io.Copy(tmpFile, src)
			if err != nil {
				log.Printf("[ERROR] SubmitTask - Failed to copy file: %v", err)
				tmpFile.Close()
				os.Remove(tmpFile.Name())
				for _, path := range uploadedPaths {
					os.Remove(path)
				}
				return c.Status(fiber.StatusInternalServerError).JSON(response{
					Success: false,
					Message: "Internal server error",
					Error:   "failed to upload file",
				})
			}
			log.Printf("[DEBUG] SubmitTask - Copied %d bytes for %s", bytesRead, file.Filename)

			// Reset pointer for upload
			tmpFile.Seek(0, 0)

			// Save to task-submissions folder
			filePath, err := th.fileService.UploadTaskSubmission(tmpFile, file.Filename)
			tmpFile.Close()
			os.Remove(tmpFile.Name())
			log.Printf("[DEBUG] SubmitTask - File saved to: %s", filePath)

			if err != nil {
				log.Printf("[ERROR] SubmitTask - Failed to save file: %v", err)
				for _, path := range uploadedPaths {
					os.Remove(path)
				}
				return c.Status(fiber.StatusInternalServerError).JSON(response{
					Success: false,
					Message: "Failed to save file",
					Error:   fmt.Sprintf("failed to save attachment %d", idx+1),
				})
			}

			// Get attachment notes if provided
			var attNotes string
			if idx < len(attachmentNotes) {
				attNotes = attachmentNotes[idx]
			}

			attachmentList = append(attachmentList, usecases.AttachmentData{
				FileName: file.Filename,
				FilePath: filePath,
				FileSize: file.Size,
				Notes:    attNotes,
			})

			uploadedPaths = append(uploadedPaths, filePath)
		}
	}

	// Create submission request
	submitReq := &usecases.SubmitTaskRequest{
		Notes:       notes,
		Attachments: attachmentList,
	}

	// Submit task
	result, err := th.submissionUseCase.SubmitTask(taskID, claims.UserID, submitReq)
	if err != nil {
		log.Printf("[ERROR] SubmitTask - Failed to submit: %v", err)
		// Clean up files on error
		for _, path := range uploadedPaths {
			os.Remove(path)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Failed to submit task",
			"error":   err.Error(),
		})
	}

	log.Printf("[SUCCESS] SubmitTask - Task submitted: %s by user: %s", taskID, claims.UserID)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Task submitted successfully",
		"data":    result,
	})
}

// GetTaskSubmissions handles GET /api/v1/task-submission/:task_id
func (th *TaskSubmissionHandler) GetTaskSubmissions(c *fiber.Ctx) error {
	taskID := c.Params("task_id")
	log.Printf("[API] GET /api/v1/task-submission/%s - Retrieve submissions", taskID)

	// Extract token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	// Validate token
	_, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	// Get all submissions for the task
	submissions, err := th.submissionUseCase.GetTaskSubmissions(taskID)
	if err != nil {
		log.Printf("[ERROR] GetTaskSubmissions - Failed to retrieve: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to retrieve submissions",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] GetTaskSubmissions - Retrieved %d submissions for task: %s", len(submissions), taskID)
	return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
		"success": true,
		"message": "Submissions retrieved successfully",
		"data":    submissions,
	})
}

// ReviewSubmission handles POST /api/v1/task-submission/:submission_id/review
func (th *TaskSubmissionHandler) ReviewSubmission(c *fiber.Ctx) error {
	submissionID := c.Params("submission_id")
	log.Printf("[API] POST /api/v1/task-submission/%s/review - Review attempt", submissionID)

	// Extract token
	tokenString := extractTokenFromHeader(c)
	if tokenString == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   "missing token",
		})
	}

	// Validate token
	claims, err := th.authUseCase.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response{
			Success: false,
			Message: "Unauthorized",
			Error:   err.Error(),
		})
	}

	// Only admins and stakeholders can review
	if claims.Role != "admin" && claims.Role != "stakeholder" {
		log.Printf("[ERROR] ReviewSubmission - Invalid role: %s", claims.Role)
		return c.Status(fiber.StatusForbidden).JSON(response{
			Success: false,
			Message: "Forbidden",
			Error:   "only admins and stakeholders can review submissions",
		})
	}

	// Parse request body
	var req usecases.ReviewTaskSubmissionRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("[ERROR] ReviewSubmission - Failed to parse body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Invalid request",
			Error:   err.Error(),
		})
	}

	// Create review
	review, err := th.submissionUseCase.ReviewTaskSubmission(submissionID, claims.UserID, claims.Role, &req)
	if err != nil {
		log.Printf("[ERROR] ReviewSubmission - Failed to review: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(response{
			Success: false,
			Message: "Failed to review submission",
			Error:   err.Error(),
		})
	}

	log.Printf("[SUCCESS] ReviewSubmission - Reviewed by %s: %s", claims.Role, submissionID)
	return c.Status(fiber.StatusCreated).JSON(map[string]interface{}{
		"success": true,
		"message": "Review created successfully",
		"data":    review,
	})
}

// DownloadTaskSubmissionAttachment handles GET /api/v1/task-submission/attachments/:attachment_id
// Allows downloading an attachment from a task submission
func (th *TaskSubmissionHandler) DownloadTaskSubmissionAttachment(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		log.Printf("[ERROR] DownloadTaskSubmissionAttachment - User not authenticated")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	attachmentID := c.Params("attachment_id")
	if attachmentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "attachment_id is required",
		})
	}

	// Check if attachment repository is available
	if th.attachmentRepo == nil || th.fileService == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "attachment service not available",
		})
	}

	// Get attachment from database
	attachment, err := th.attachmentRepo.GetAttachmentByID(attachmentID)
	if err != nil {
		log.Printf("[ERROR] DownloadTaskSubmissionAttachment - Attachment not found: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "attachment not found",
		})
	}

	// Verify file exists
	if _, err := os.Stat(attachment.FilePath); err != nil {
		log.Printf("[ERROR] DownloadTaskSubmissionAttachment - File not found on disk: %v", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "file not found on disk",
		})
	}

	log.Printf("[DEBUG] DownloadTaskSubmissionAttachment - Sending file: %s", attachment.FilePath)
	log.Printf("[SUCCESS] DownloadTaskSubmissionAttachment - Downloaded: %s by user %s", attachment.FileName, userID)

	// Send file with proper headers
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", attachment.FileName))
	return c.SendFile(attachment.FilePath)
}
