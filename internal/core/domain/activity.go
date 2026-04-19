package domain

import "time"

// ActivityType represents the type of activity
type ActivityType string

const (
	ActivityTypeUserLogin            ActivityType = "user_login"
	ActivityTypeUserLogout           ActivityType = "user_logout"
	ActivityTypePasswordChanged      ActivityType = "password_changed"
	ActivityTypePasswordReset        ActivityType = "password_reset"
	ActivityTypeUserCreated          ActivityType = "user_created"
	ActivityTypeUserUpdated          ActivityType = "user_updated"
	ActivityTypeUserDeleted          ActivityType = "user_deleted"
	ActivityTypeReportSubmitted      ActivityType = "report_submitted"
	ActivityTypeReportViewed         ActivityType = "report_viewed"
	ActivityTypeAIChatUsed           ActivityType = "ai_chat_used"
	ActivityTypeProjectCreated       ActivityType = "project_created"
	ActivityTypeProjectUpdated       ActivityType = "project_updated"
	ActivityTypeProjectDeleted       ActivityType = "project_deleted"
	ActivityTypeTaskCreated          ActivityType = "task_created"
	ActivityTypeTaskUpdated          ActivityType = "task_updated"
	ActivityTypeTaskSubmitted        ActivityType = "task_submitted"
	ActivityTypeTaskAssigned         ActivityType = "task_assigned"
	ActivityTypeVulnerabilityCreated ActivityType = "vulnerability_created"
	ActivityTypeVulnerabilityUpdated ActivityType = "vulnerability_updated"
	ActivityTypeMessageSent          ActivityType = "message_sent"
	ActivityTypeNotificationCreated  ActivityType = "notification_created"
	ActivityTypeProfileUpdated       ActivityType = "profile_updated"
	ActivityTypeFileUploaded         ActivityType = "file_uploaded"
	ActivityTypeFileDownloaded       ActivityType = "file_downloaded"
)

// Activity represents a user action/activity log
type Activity struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	Email        string                 `json:"email,omitempty"`
	ActivityType ActivityType           `json:"activity_type"`
	Description  string                 `json:"description"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ActivityRepository defines methods for activity persistence
type ActivityRepository interface {
	CreateActivity(activity *Activity) (*Activity, error)
	GetActivityByID(activityID string) (*Activity, error)
	GetUserActivities(userID string, limit, offset int) ([]*Activity, error)
	GetAllActivities(limit, offset int) ([]*Activity, error)
	GetActivitiesByType(activityType ActivityType, limit, offset int) ([]*Activity, error)
	GetActivitiesByDateRange(startDate, endDate time.Time, limit, offset int) ([]*Activity, error)
}
