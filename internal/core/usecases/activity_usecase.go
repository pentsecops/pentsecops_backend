package usecases

import (
	"fmt"
	"log"
	"time"

	"github.com/pentsecops/backend/internal/core/domain"
)

// ActivityUseCase handles business logic for activities/audit logs
type ActivityUseCase struct {
	activityRepo domain.ActivityRepository
	userRepo     domain.UserRepository
}

// NewActivityUseCase creates a new activity use case
func NewActivityUseCase(
	activityRepo domain.ActivityRepository,
	userRepo domain.UserRepository,
) *ActivityUseCase {
	return &ActivityUseCase{
		activityRepo: activityRepo,
		userRepo:     userRepo,
	}
}

// LogActivity creates a new activity log entry
func (uc *ActivityUseCase) LogActivity(
	userID string,
	activityType domain.ActivityType,
	description string,
	metadata map[string]interface{},
	ipAddress string,
	userAgent string,
) (*domain.Activity, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Get user email
	user, err := uc.userRepo.GetUserByID(userID)
	userEmail := ""
	if err == nil && user != nil {
		userEmail = user.Email
	}

	activity := &domain.Activity{
		UserID:       userID,
		Email:        userEmail,
		ActivityType: activityType,
		Description:  description,
		Metadata:     metadata,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}

	savedActivity, err := uc.activityRepo.CreateActivity(activity)
	if err != nil {
		log.Printf("[ERROR] LogActivity - Failed to save activity: %v", err)
		return nil, fmt.Errorf("failed to log activity: %w", err)
	}

	log.Printf("[SUCCESS] Activity logged: type=%s, user=%s", activityType, userID)
	return savedActivity, nil
}

// GetUserActivities retrieves activities for a specific user
func (uc *ActivityUseCase) GetUserActivities(userID string, limit, offset int) ([]*domain.Activity, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	activities, err := uc.activityRepo.GetUserActivities(userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetUserActivities - Failed: %v", err)
		return nil, fmt.Errorf("failed to get user activities: %w", err)
	}

	return activities, nil
}

// GetAllActivities retrieves all activities (admin only)
func (uc *ActivityUseCase) GetAllActivities(limit, offset int) ([]*domain.Activity, error) {
	activities, err := uc.activityRepo.GetAllActivities(limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetAllActivities - Failed: %v", err)
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, nil
}

// GetActivitiesByType retrieves activities by type
func (uc *ActivityUseCase) GetActivitiesByType(activityType domain.ActivityType, limit, offset int) ([]*domain.Activity, error) {
	activities, err := uc.activityRepo.GetActivitiesByType(activityType, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetActivitiesByType - Failed: %v", err)
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, nil
}

// GetRecentActivities retrieves recent activities (last 24 hours)
func (uc *ActivityUseCase) GetRecentActivities(limit, offset int) ([]*domain.Activity, error) {
	now := time.Now()
	startDate := now.Add(-24 * time.Hour)

	activities, err := uc.activityRepo.GetActivitiesByDateRange(startDate, now, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetRecentActivities - Failed: %v", err)
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}

	return activities, nil
}

// GetActivitiesByDateRange retrieves activities within a date range
func (uc *ActivityUseCase) GetActivitiesByDateRange(startDate, endDate time.Time, limit, offset int) ([]*domain.Activity, error) {
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start_date must be before end_date")
	}

	activities, err := uc.activityRepo.GetActivitiesByDateRange(startDate, endDate, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetActivitiesByDateRange - Failed: %v", err)
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return activities, nil
}
