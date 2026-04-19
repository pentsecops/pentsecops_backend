package usecases

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pentsecops/backend/internal/core/domain"
)

// TaskSubmissionUseCase handles task submission business logic
type TaskSubmissionUseCase struct {
	submissionRepo domain.TaskSubmissionRepository
	attachmentRepo domain.TaskSubmissionAttachmentRepository
	reviewRepo     domain.TaskSubmissionReviewRepository
	taskRepo       domain.TaskRepository
	pentesterRepo  domain.PentesterRepository
}

// NewTaskSubmissionUseCase creates a new task submission use case
func NewTaskSubmissionUseCase(
	submissionRepo domain.TaskSubmissionRepository,
	attachmentRepo domain.TaskSubmissionAttachmentRepository,
	reviewRepo domain.TaskSubmissionReviewRepository,
	taskRepo domain.TaskRepository,
	pentesterRepo domain.PentesterRepository,
) *TaskSubmissionUseCase {
	return &TaskSubmissionUseCase{
		submissionRepo: submissionRepo,
		attachmentRepo: attachmentRepo,
		reviewRepo:     reviewRepo,
		taskRepo:       taskRepo,
		pentesterRepo:  pentesterRepo,
	}
}

// SubmitTaskRequest contains data for task submission
type SubmitTaskRequest struct {
	Notes       string           `json:"notes"`
	Attachments []AttachmentData `json:"attachments,omitempty"`
}

// AttachmentData represents file attachment data
type AttachmentData struct {
	FileName string
	FilePath string
	FileSize int64
	Notes    string
}

// SubmitTask creates a new task submission
func (uc *TaskSubmissionUseCase) SubmitTask(taskID, userID string, req *SubmitTaskRequest) (*domain.TaskSubmissionResponse, error) {
	// Verify pentester
	pentester, err := uc.pentesterRepo.GetPentesterByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("pentester not found: %w", err)
	}

	// Verify task exists and pentester is assigned
	task, err := uc.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	if task.AssignedTo == nil {
		return nil, fmt.Errorf("this task has not been assigned to anyone yet")
	}

	if *task.AssignedTo != pentester.ID {
		return nil, fmt.Errorf("this task is assigned to another pentester, not you")
	}

	// Create submission
	submission := &domain.TaskSubmission{
		ID:          uuid.New().String(),
		TaskID:      taskID,
		PentesterID: pentester.ID,
		Notes:       req.Notes,
		SubmittedAt: time.Now().UTC(),
	}

	if err := uc.submissionRepo.CreateSubmission(submission); err != nil {
		return nil, fmt.Errorf("failed to create submission: %w", err)
	}

	// Add attachments
	var attachments []*domain.TaskSubmissionAttachment
	if len(req.Attachments) > 0 {
		for _, att := range req.Attachments {
			attachment := &domain.TaskSubmissionAttachment{
				ID:           uuid.New().String(),
				SubmissionID: submission.ID,
				FileName:     att.FileName,
				FilePath:     att.FilePath,
				FileSize:     att.FileSize,
				Notes:        att.Notes,
				UploadedAt:   time.Now().UTC(),
			}

			if err := uc.attachmentRepo.CreateAttachment(attachment); err != nil {
				// Log but don't fail the submission
				fmt.Printf("[WARN] Failed to save attachment: %v\n", err)
				continue
			}

			attachments = append(attachments, attachment)
		}
	}

	return &domain.TaskSubmissionResponse{
		Submission:  submission,
		Attachments: attachments,
		Reviews:     []*domain.TaskSubmissionReview{},
	}, nil
}

// GetTaskSubmissions retrieves all submissions with their attachments and reviews
func (uc *TaskSubmissionUseCase) GetTaskSubmissions(taskID string) ([]*domain.TaskSubmissionResponse, error) {
	// Verify task exists
	_, err := uc.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, fmt.Errorf("task not found: %w", err)
	}

	// Get submissions
	submissions, err := uc.submissionRepo.GetSubmissionsByTask(taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get submissions: %w", err)
	}

	var responses []*domain.TaskSubmissionResponse
	for _, submission := range submissions {
		// Get attachments for this submission
		attachments, err := uc.attachmentRepo.GetAttachmentsBySubmission(submission.ID)
		if err != nil {
			attachments = []*domain.TaskSubmissionAttachment{}
		}

		// Get reviews for this submission
		reviews, err := uc.reviewRepo.GetReviewsBySubmission(submission.ID)
		if err != nil {
			reviews = []*domain.TaskSubmissionReview{}
		}

		responses = append(responses, &domain.TaskSubmissionResponse{
			Submission:  submission,
			Attachments: attachments,
			Reviews:     reviews,
		})
	}

	return responses, nil
}

// ReviewTaskSubmissionRequest contains data for reviewing a submission
type ReviewTaskSubmissionRequest struct {
	Status      string `json:"status"` // in_progress, completed
	ReviewNotes string `json:"review_notes"`
}

// ReviewTaskSubmission creates a review for a task submission
func (uc *TaskSubmissionUseCase) ReviewTaskSubmission(submissionID, reviewerID, reviewerRole string, req *ReviewTaskSubmissionRequest) (*domain.TaskSubmissionReview, error) {
	// Verify submission exists
	submission, err := uc.submissionRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// Verify reviewer role is valid
	if reviewerRole != "admin" && reviewerRole != "stakeholder" {
		return nil, fmt.Errorf("invalid reviewer role: %s", reviewerRole)
	}

	// Verify status is valid
	if req.Status != "in_progress" && req.Status != "completed" {
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Create review
	review := &domain.TaskSubmissionReview{
		ID:           uuid.New().String(),
		SubmissionID: submission.ID,
		ReviewerID:   reviewerID,
		ReviewerRole: reviewerRole,
		Status:       req.Status,
		ReviewNotes:  req.ReviewNotes,
		ReviewedAt:   time.Now().UTC(),
	}

	if err := uc.reviewRepo.CreateReview(review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return review, nil
}

// GetSubmission retrieves a single submission with all its data
func (uc *TaskSubmissionUseCase) GetSubmission(submissionID string) (*domain.TaskSubmissionResponse, error) {
	submission, err := uc.submissionRepo.GetSubmissionByID(submissionID)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}

	// Get attachments
	attachments, err := uc.attachmentRepo.GetAttachmentsBySubmission(submissionID)
	if err != nil {
		attachments = []*domain.TaskSubmissionAttachment{}
	}

	// Get reviews
	reviews, err := uc.reviewRepo.GetReviewsBySubmission(submissionID)
	if err != nil {
		reviews = []*domain.TaskSubmissionReview{}
	}

	return &domain.TaskSubmissionResponse{
		Submission:  submission,
		Attachments: attachments,
		Reviews:     reviews,
	}, nil
}
