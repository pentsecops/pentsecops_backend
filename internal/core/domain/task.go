package domain

import "time"

// Task defines a testing activity within a project
type Task struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	AssignedTo  *string    `json:"assigned_to"` // pentester id
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"` // low, medium, high, critical
	Status      string     `json:"status"`   // assigned, in_progress, submitted, review, closed
	Deadline    *time.Time `json:"deadline"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TaskResponse is the response model for task
type TaskResponse struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	AssignedTo  *string    `json:"assigned_to"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Status      string     `json:"status"`
	Deadline    *time.Time `json:"deadline"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ToResponse converts Task to TaskResponse
func (t *Task) ToResponse() *TaskResponse {
	return &TaskResponse{
		ID:          t.ID,
		ProjectID:   t.ProjectID,
		AssignedTo:  t.AssignedTo,
		Title:       t.Title,
		Description: t.Description,
		Priority:    t.Priority,
		Status:      t.Status,
		Deadline:    t.Deadline,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// TaskSubmission defines a submission of a task by a pentester
type TaskSubmission struct {
	ID          string    `json:"id"`
	TaskID      string    `json:"task_id"`
	PentesterID string    `json:"pentester_id"`
	Notes       string    `json:"notes"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// TaskSubmissionAttachment defines an attachment for a task submission
type TaskSubmissionAttachment struct {
	ID           string    `json:"id"`
	SubmissionID string    `json:"submission_id"`
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`
	FileSize     int64     `json:"file_size"`
	Notes        string    `json:"notes,omitempty"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

// TaskSubmissionReview defines a review of a task submission by admin or stakeholder
type TaskSubmissionReview struct {
	ID           string    `json:"id"`
	SubmissionID string    `json:"submission_id"`
	ReviewerID   string    `json:"reviewer_id"`
	ReviewerRole string    `json:"reviewer_role"` // admin or stakeholder
	Status       string    `json:"status"`        // in_progress or completed
	ReviewNotes  string    `json:"review_notes,omitempty"`
	ReviewedAt   time.Time `json:"reviewed_at"`
}

// TaskSubmissionResponse is the complete response including submission, attachments, and reviews
type TaskSubmissionResponse struct {
	Submission  *TaskSubmission             `json:"submission"`
	Attachments []*TaskSubmissionAttachment `json:"attachments,omitempty"`
	Reviews     []*TaskSubmissionReview     `json:"reviews,omitempty"`
}
