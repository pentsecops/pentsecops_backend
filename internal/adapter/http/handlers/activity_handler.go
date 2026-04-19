package handlers

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/core/domain"
	"github.com/pentsecops/backend/internal/core/usecases"
)

// ActivityHandler handles activity-related HTTP requests
type ActivityHandler struct {
	activityUseCase *usecases.ActivityUseCase
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(activityUseCase *usecases.ActivityUseCase) *ActivityHandler {
	return &ActivityHandler{
		activityUseCase: activityUseCase,
	}
}

// ActivityResponse represents an activity in the response
type ActivityResponse struct {
	ID           string            `json:"id"`
	UserID       string            `json:"user_id"`
	Email        string            `json:"email,omitempty"`
	ActivityType string            `json:"activity_type"`
	Description  string            `json:"description"`
	Metadata     map[string]string `json:"metadata,omitempty"`
	IPAddress    string            `json:"ip_address,omitempty"`
	UserAgent    string            `json:"user_agent,omitempty"`
	CreatedAt    string            `json:"created_at"`
}

// GetUserActivities handles GET /api/v1/activities/user
// Retrieves activities for the authenticated user
func (h *ActivityHandler) GetUserActivities(c *fiber.Ctx) error {
	// Get authenticated user ID
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse pagination parameters
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get user activities
	activities, err := h.activityUseCase.GetUserActivities(userID, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetUserActivities - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to responses
	responses := make([]ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = h.activityToResponse(activity)
	}

	return c.JSON(fiber.Map{
		"activities": responses,
		"count":      len(responses),
		"limit":      limit,
		"offset":     offset,
	})
}

// GetAllActivities handles GET /api/v1/activities
// Retrieves all activities (admin only)
func (h *ActivityHandler) GetAllActivities(c *fiber.Ctx) error {
	// Get authenticated user - should be admin
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse pagination parameters
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get all activities
	activities, err := h.activityUseCase.GetAllActivities(limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetAllActivities - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to responses
	responses := make([]ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = h.activityToResponse(activity)
	}

	return c.JSON(fiber.Map{
		"activities": responses,
		"count":      len(responses),
		"limit":      limit,
		"offset":     offset,
	})
}

// GetRecentActivities handles GET /api/v1/activities/recent
// Retrieves recent activities (last 24 hours)
func (h *ActivityHandler) GetRecentActivities(c *fiber.Ctx) error {
	// Get authenticated user - should be admin
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// Parse pagination parameters
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get recent activities
	activities, err := h.activityUseCase.GetRecentActivities(limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetRecentActivities - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to responses
	responses := make([]ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = h.activityToResponse(activity)
	}

	return c.JSON(fiber.Map{
		"activities": responses,
		"count":      len(responses),
		"limit":      limit,
		"offset":     offset,
	})
}

// GetActivitiesByType handles GET /api/v1/activities/type/:type
// Retrieves activities by type
func (h *ActivityHandler) GetActivitiesByType(c *fiber.Ctx) error {
	// Get authenticated user - should be admin
	userID, ok := c.Locals("userID").(string)
	if !ok || userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	activityTypeStr := c.Params("type")
	if activityTypeStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "activity type is required",
		})
	}

	activityType := domain.ActivityType(activityTypeStr)

	// Parse pagination parameters
	limit := 50
	offset := 0
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Get activities by type
	activities, err := h.activityUseCase.GetActivitiesByType(activityType, limit, offset)
	if err != nil {
		log.Printf("[ERROR] GetActivitiesByType - Failed: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Convert to responses
	responses := make([]ActivityResponse, len(activities))
	for i, activity := range activities {
		responses[i] = h.activityToResponse(activity)
	}

	return c.JSON(fiber.Map{
		"activities": responses,
		"count":      len(responses),
		"limit":      limit,
		"offset":     offset,
	})
}

// Helper function to convert domain.Activity to ActivityResponse
func (h *ActivityHandler) activityToResponse(activity *domain.Activity) ActivityResponse {
	metadata := make(map[string]string)
	if activity.Metadata != nil {
		for k, v := range activity.Metadata {
			if val, ok := v.(string); ok {
				metadata[k] = val
			}
		}
	}

	return ActivityResponse{
		ID:           activity.ID,
		UserID:       activity.UserID,
		Email:        activity.Email,
		ActivityType: string(activity.ActivityType),
		Description:  activity.Description,
		Metadata:     metadata,
		IPAddress:    activity.IPAddress,
		UserAgent:    activity.UserAgent,
		CreatedAt:    activity.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
