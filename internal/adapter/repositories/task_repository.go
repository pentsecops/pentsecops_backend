package repositories

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// TaskRepositoryImpl implements the TaskRepository interface
type TaskRepositoryImpl struct {
	db *sql.DB
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *sql.DB) domain.TaskRepository {
	return &TaskRepositoryImpl{db: db}
}

// CreateTask creates a new task in the database
func (r *TaskRepositoryImpl) CreateTask(task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, project_id, assigned_to, title, description, priority, status, deadline, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.Exec(query,
		task.ID,
		task.ProjectID,
		task.AssignedTo,
		task.Title,
		task.Description,
		task.Priority,
		task.Status,
		task.Deadline,
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// GetTaskByID retrieves a task by ID
func (r *TaskRepositoryImpl) GetTaskByID(id string) (*domain.Task, error) {
	query := `
		SELECT id, project_id, assigned_to, title, description, priority, status, deadline, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`

	task := &domain.Task{}
	err := r.db.QueryRow(query, id).Scan(
		&task.ID,
		&task.ProjectID,
		&task.AssignedTo,
		&task.Title,
		&task.Description,
		&task.Priority,
		&task.Status,
		&task.Deadline,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return task, nil
}

// GetTasksByProject retrieves all tasks for a project
func (r *TaskRepositoryImpl) GetTasksByProject(projectID string) ([]*domain.Task, error) {
	query := `
		SELECT id, project_id, assigned_to, title, description, priority, status, deadline, created_at, updated_at
		FROM tasks
		WHERE project_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssignedTo,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.Status,
			&task.Deadline,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByPentester retrieves all tasks assigned to a pentester
func (r *TaskRepositoryImpl) GetTasksByPentester(pentesterID string) ([]*domain.Task, error) {
	query := `
		SELECT id, project_id, assigned_to, title, description, priority, status, deadline, created_at, updated_at
		FROM tasks
		WHERE assigned_to = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, pentesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssignedTo,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.Status,
			&task.Deadline,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksByStatus retrieves all tasks with a specific status
func (r *TaskRepositoryImpl) GetTasksByStatus(status string) ([]*domain.Task, error) {
	query := `
		SELECT id, project_id, assigned_to, title, description, priority, status, deadline, created_at, updated_at
		FROM tasks
		WHERE status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssignedTo,
			&task.Title,
			&task.Description,
			&task.Priority,
			&task.Status,
			&task.Deadline,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// UpdateTask updates a task
func (r *TaskRepositoryImpl) UpdateTask(task *domain.Task) error {
	query := `
		UPDATE tasks
		SET project_id = $1, assigned_to = $2, title = $3, description = $4, priority = $5, status = $6, deadline = $7, updated_at = $8
		WHERE id = $9
	`

	_, err := r.db.Exec(query,
		task.ProjectID,
		task.AssignedTo,
		task.Title,
		task.Description,
		task.Priority,
		task.Status,
		task.Deadline,
		time.Now().UTC(),
		task.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

// DeleteTask deletes a task
func (r *TaskRepositoryImpl) DeleteTask(id string) error {
	query := `DELETE FROM tasks WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// UpdateTaskStatus updates the status of a task
func (r *TaskRepositoryImpl) UpdateTaskStatus(taskID, status string) error {
	query := `
		UPDATE tasks
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Exec(query, status, time.Now().UTC(), taskID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}

	return nil
}

// TaskSubmissionRepositoryImpl implements TaskSubmissionRepository
type TaskSubmissionRepositoryImpl struct {
	db *sql.DB
}

// NewTaskSubmissionRepository creates a new task submission repository
func NewTaskSubmissionRepository(db *sql.DB) domain.TaskSubmissionRepository {
	return &TaskSubmissionRepositoryImpl{db: db}
}

func (r *TaskSubmissionRepositoryImpl) CreateSubmission(submission *domain.TaskSubmission) error {
	query := `
		INSERT INTO task_submissions (id, task_id, pentester_id, notes, submitted_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(query,
		submission.ID,
		submission.TaskID,
		submission.PentesterID,
		submission.Notes,
		submission.SubmittedAt,
	)

	return err
}

func (r *TaskSubmissionRepositoryImpl) GetSubmissionByID(id string) (*domain.TaskSubmission, error) {
	query := `
		SELECT id, task_id, pentester_id, notes, submitted_at
		FROM task_submissions
		WHERE id = $1
	`

	submission := &domain.TaskSubmission{}
	err := r.db.QueryRow(query, id).Scan(
		&submission.ID,
		&submission.TaskID,
		&submission.PentesterID,
		&submission.Notes,
		&submission.SubmittedAt,
	)

	if err != nil {
		return nil, err
	}

	return submission, nil
}

func (r *TaskSubmissionRepositoryImpl) GetSubmissionsByTask(taskID string) ([]*domain.TaskSubmission, error) {
	query := `
		SELECT id, task_id, pentester_id, notes, submitted_at
		FROM task_submissions
		WHERE task_id = $1
		ORDER BY submitted_at DESC
	`

	rows, err := r.db.Query(query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*domain.TaskSubmission
	for rows.Next() {
		submission := &domain.TaskSubmission{}
		err := rows.Scan(
			&submission.ID,
			&submission.TaskID,
			&submission.PentesterID,
			&submission.Notes,
			&submission.SubmittedAt,
		)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}

	return submissions, rows.Err()
}

func (r *TaskSubmissionRepositoryImpl) DeleteSubmission(id string) error {
	query := `DELETE FROM task_submissions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// TaskSubmissionAttachmentRepositoryImpl implements TaskSubmissionAttachmentRepository
type TaskSubmissionAttachmentRepositoryImpl struct {
	db *sql.DB
}

// NewTaskSubmissionAttachmentRepository creates a new task submission attachment repository
func NewTaskSubmissionAttachmentRepository(db *sql.DB) domain.TaskSubmissionAttachmentRepository {
	return &TaskSubmissionAttachmentRepositoryImpl{db: db}
}

func (r *TaskSubmissionAttachmentRepositoryImpl) CreateAttachment(attachment *domain.TaskSubmissionAttachment) error {
	query := `
		INSERT INTO task_submission_attachments (id, submission_id, file_name, file_path, file_size, notes, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query,
		attachment.ID,
		attachment.SubmissionID,
		attachment.FileName,
		attachment.FilePath,
		attachment.FileSize,
		attachment.Notes,
		attachment.UploadedAt,
	)

	return err
}

func (r *TaskSubmissionAttachmentRepositoryImpl) GetAttachmentsBySubmission(submissionID string) ([]*domain.TaskSubmissionAttachment, error) {
	query := `
		SELECT id, submission_id, file_name, file_path, file_size, notes, uploaded_at
		FROM task_submission_attachments
		WHERE submission_id = $1
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Query(query, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attachments []*domain.TaskSubmissionAttachment
	for rows.Next() {
		attachment := &domain.TaskSubmissionAttachment{}
		err := rows.Scan(
			&attachment.ID,
			&attachment.SubmissionID,
			&attachment.FileName,
			&attachment.FilePath,
			&attachment.FileSize,
			&attachment.Notes,
			&attachment.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}

	return attachments, rows.Err()
}

func (r *TaskSubmissionAttachmentRepositoryImpl) GetAttachmentByID(id string) (*domain.TaskSubmissionAttachment, error) {
	query := `
		SELECT id, submission_id, file_name, file_path, file_size, notes, uploaded_at
		FROM task_submission_attachments
		WHERE id = $1
	`

	attachment := &domain.TaskSubmissionAttachment{}
	err := r.db.QueryRow(query, id).Scan(
		&attachment.ID,
		&attachment.SubmissionID,
		&attachment.FileName,
		&attachment.FilePath,
		&attachment.FileSize,
		&attachment.Notes,
		&attachment.UploadedAt,
	)

	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (r *TaskSubmissionAttachmentRepositoryImpl) DeleteAttachment(id string) error {
	query := `DELETE FROM task_submission_attachments WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// TaskSubmissionReviewRepositoryImpl implements TaskSubmissionReviewRepository
type TaskSubmissionReviewRepositoryImpl struct {
	db *sql.DB
}

// NewTaskSubmissionReviewRepository creates a new task submission review repository
func NewTaskSubmissionReviewRepository(db *sql.DB) domain.TaskSubmissionReviewRepository {
	return &TaskSubmissionReviewRepositoryImpl{db: db}
}

func (r *TaskSubmissionReviewRepositoryImpl) CreateReview(review *domain.TaskSubmissionReview) error {
	query := `
		INSERT INTO task_submission_reviews (id, submission_id, reviewer_id, reviewer_role, status, review_notes, reviewed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(query,
		review.ID,
		review.SubmissionID,
		review.ReviewerID,
		review.ReviewerRole,
		review.Status,
		review.ReviewNotes,
		review.ReviewedAt,
	)

	return err
}

func (r *TaskSubmissionReviewRepositoryImpl) GetReviewsBySubmission(submissionID string) ([]*domain.TaskSubmissionReview, error) {
	query := `
		SELECT id, submission_id, reviewer_id, reviewer_role, status, review_notes, reviewed_at
		FROM task_submission_reviews
		WHERE submission_id = $1
		ORDER BY reviewed_at DESC
	`

	rows, err := r.db.Query(query, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []*domain.TaskSubmissionReview
	for rows.Next() {
		review := &domain.TaskSubmissionReview{}
		err := rows.Scan(
			&review.ID,
			&review.SubmissionID,
			&review.ReviewerID,
			&review.ReviewerRole,
			&review.Status,
			&review.ReviewNotes,
			&review.ReviewedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

func (r *TaskSubmissionReviewRepositoryImpl) GetReviewByID(id string) (*domain.TaskSubmissionReview, error) {
	query := `
		SELECT id, submission_id, reviewer_id, reviewer_role, status, review_notes, reviewed_at
		FROM task_submission_reviews
		WHERE id = $1
	`

	review := &domain.TaskSubmissionReview{}
	err := r.db.QueryRow(query, id).Scan(
		&review.ID,
		&review.SubmissionID,
		&review.ReviewerID,
		&review.ReviewerRole,
		&review.Status,
		&review.ReviewNotes,
		&review.ReviewedAt,
	)

	if err != nil {
		return nil, err
	}

	return review, nil
}

func (r *TaskSubmissionReviewRepositoryImpl) UpdateReview(review *domain.TaskSubmissionReview) error {
	query := `
		UPDATE task_submission_reviews
		SET status = $1, review_notes = $2, reviewed_at = $3
		WHERE id = $4
	`

	_, err := r.db.Exec(query, review.Status, review.ReviewNotes, review.ReviewedAt, review.ID)
	return err
}

func (r *TaskSubmissionReviewRepositoryImpl) DeleteReview(id string) error {
	query := `DELETE FROM task_submission_reviews WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
