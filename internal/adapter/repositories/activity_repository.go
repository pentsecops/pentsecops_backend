package repositories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// ActivityRepository handles activity persistence
type ActivityRepository struct {
	db *sql.DB
}

// NewActivityRepository creates a new activity repository
func NewActivityRepository(db *sql.DB) *ActivityRepository {
	return &ActivityRepository{db: db}
}

// CreateActivity creates a new activity log entry
func (r *ActivityRepository) CreateActivity(activity *domain.Activity) (*domain.Activity, error) {
	if activity.UserID == "" || activity.ActivityType == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	// Convert metadata to JSON
	var metadataJSON *string
	if activity.Metadata != nil && len(activity.Metadata) > 0 {
		data, err := json.Marshal(activity.Metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metadata: %w", err)
		}
		jsonStr := string(data)
		metadataJSON = &jsonStr
	}

	query := `
		INSERT INTO activities (user_id, email, activity_type, description, metadata, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, user_id, email, activity_type, description, metadata, ip_address, user_agent, created_at
	`

	var metadata *string
	err := r.db.QueryRow(
		query,
		activity.UserID,
		activity.Email,
		activity.ActivityType,
		activity.Description,
		metadataJSON,
		activity.IPAddress,
		activity.UserAgent,
		time.Now(),
	).Scan(
		&activity.ID,
		&activity.UserID,
		&activity.Email,
		&activity.ActivityType,
		&activity.Description,
		&metadata,
		&activity.IPAddress,
		&activity.UserAgent,
		&activity.CreatedAt,
	)

	if err != nil {
		log.Printf("[ERROR] Failed to create activity: %v", err)
		return nil, fmt.Errorf("failed to create activity: %w", err)
	}

	// Parse metadata if present
	if metadata != nil {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(*metadata), &m); err == nil {
			activity.Metadata = m
		}
	}

	log.Printf("[SUCCESS] Activity created: id=%s, user=%s, type=%s", activity.ID, activity.UserID, activity.ActivityType)
	return activity, nil
}

// GetActivityByID retrieves an activity by ID
func (r *ActivityRepository) GetActivityByID(activityID string) (*domain.Activity, error) {
	query := `
		SELECT id, user_id, email, activity_type, description, metadata, ip_address, user_agent, created_at
		FROM activities
		WHERE id = $1
	`

	return r.scanActivity(r.db.QueryRow(query, activityID))
}

// GetUserActivities retrieves activities for a specific user
func (r *ActivityRepository) GetUserActivities(userID string, limit, offset int) ([]*domain.Activity, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, user_id, email, activity_type, description, metadata, ip_address, user_agent, created_at
		FROM activities
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.scanActivities(r.db.Query(query, userID, limit, offset))
}

// GetAllActivities retrieves all activities with pagination
func (r *ActivityRepository) GetAllActivities(limit, offset int) ([]*domain.Activity, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, user_id, email, activity_type, description, metadata, ip_address, user_agent, created_at
		FROM activities
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	return r.scanActivities(r.db.Query(query, limit, offset))
}

// GetActivitiesByType retrieves activities by type
func (r *ActivityRepository) GetActivitiesByType(activityType domain.ActivityType, limit, offset int) ([]*domain.Activity, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, user_id, email, activity_type, description, metadata, ip_address, user_agent, created_at
		FROM activities
		WHERE activity_type = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	return r.scanActivities(r.db.Query(query, activityType, limit, offset))
}

// GetActivitiesByDateRange retrieves activities within a date range
func (r *ActivityRepository) GetActivitiesByDateRange(startDate, endDate time.Time, limit, offset int) ([]*domain.Activity, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, user_id, email, activity_type, description, metadata, ip_address, user_agent, created_at
		FROM activities
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	return r.scanActivities(r.db.Query(query, startDate, endDate, limit, offset))
}

// Helper function to scan a single activity
func (r *ActivityRepository) scanActivity(row *sql.Row) (*domain.Activity, error) {
	var activity domain.Activity
	var metadata *string

	err := row.Scan(
		&activity.ID,
		&activity.UserID,
		&activity.Email,
		&activity.ActivityType,
		&activity.Description,
		&metadata,
		&activity.IPAddress,
		&activity.UserAgent,
		&activity.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("activity not found")
		}
		return nil, fmt.Errorf("failed to get activity: %w", err)
	}

	// Parse metadata if present
	if metadata != nil {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(*metadata), &m); err == nil {
			activity.Metadata = m
		}
	}

	return &activity, nil
}

// Helper function to scan multiple activities
func (r *ActivityRepository) scanActivities(rows *sql.Rows, err error) ([]*domain.Activity, error) {
	if err != nil {
		log.Printf("[ERROR] Failed to query activities: %v", err)
		return nil, fmt.Errorf("failed to query activities: %w", err)
	}
	defer rows.Close()

	var activities []*domain.Activity
	for rows.Next() {
		var activity domain.Activity
		var metadata *string

		if err := rows.Scan(
			&activity.ID,
			&activity.UserID,
			&activity.Email,
			&activity.ActivityType,
			&activity.Description,
			&metadata,
			&activity.IPAddress,
			&activity.UserAgent,
			&activity.CreatedAt,
		); err != nil {
			log.Printf("[ERROR] Failed to scan activity: %v", err)
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}

		// Parse metadata if present
		if metadata != nil {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(*metadata), &m); err == nil {
				activity.Metadata = m
			}
		}

		activities = append(activities, &activity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating activities: %w", err)
	}

	return activities, nil
}
