package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pentsecops/backend/internal/adapter/http/handlers"
	"github.com/pentsecops/backend/pkg/middleware"
)

// Router defines all API routes for the application
type Router struct {
	authHandler           *handlers.AuthHandler
	profileHandler        *handlers.ProfileHandler
	userManagementHandler *handlers.UserManagementHandler
	passwordHandler       *handlers.PasswordHandler
	forgotPasswordHandler *handlers.ForgotPasswordHandler
	chatHandler           *handlers.ChatHandler
	logoutHandler         *handlers.LogoutHandler
	projectHandler        *handlers.ProjectHandler
	taskHandler           *handlers.TaskHandler
	taskSubmissionHandler *handlers.TaskSubmissionHandler
	vulnerabilityHandler  *handlers.VulnerabilityHandler
	messageHandler        *handlers.MessageHandler
	notificationHandler   *handlers.NotificationHandler
	activityHandler       *handlers.ActivityHandler
	websocketHandler      *handlers.WebSocketHandler
	tokenManager          *middleware.TokenManager
}

// NewRouter creates a new router with all handlers
func NewRouter(
	authHandler *handlers.AuthHandler,
	profileHandler *handlers.ProfileHandler,
	userManagementHandler *handlers.UserManagementHandler,
	passwordHandler *handlers.PasswordHandler,
	forgotPasswordHandler *handlers.ForgotPasswordHandler,
	chatHandler *handlers.ChatHandler,
	logoutHandler *handlers.LogoutHandler,
	projectHandler *handlers.ProjectHandler,
	taskHandler *handlers.TaskHandler,
	taskSubmissionHandler *handlers.TaskSubmissionHandler,
	vulnerabilityHandler *handlers.VulnerabilityHandler,

	messageHandler *handlers.MessageHandler,
	notificationHandler *handlers.NotificationHandler,
	activityHandler *handlers.ActivityHandler,
	websocketHandler *handlers.WebSocketHandler,
	tokenManager *middleware.TokenManager,
) *Router {
	return &Router{
		authHandler:           authHandler,
		profileHandler:        profileHandler,
		userManagementHandler: userManagementHandler,
		passwordHandler:       passwordHandler,
		forgotPasswordHandler: forgotPasswordHandler,
		chatHandler:           chatHandler,
		logoutHandler:         logoutHandler,
		projectHandler:        projectHandler,
		taskHandler:           taskHandler,
		taskSubmissionHandler: taskSubmissionHandler,
		vulnerabilityHandler:  vulnerabilityHandler,
		messageHandler:        messageHandler,
		notificationHandler:   notificationHandler,
		activityHandler:       activityHandler,
		websocketHandler:      websocketHandler,
		tokenManager:          tokenManager,
	}
}

// SetupRoutes configures all API routes
func (r *Router) SetupRoutes(app *fiber.App) {
	// Apply CORS middleware globally
	app.Use(middleware.CORSMiddleware())

	// Public routes (no authentication required)
	// User login
	app.Post("/api/v1/login", r.authHandler.LoginUser)

	// Admin routes
	app.Post("/api/v1/admin/register", r.authHandler.RegisterAdmin)
	app.Post("/api/v1/admin/login", r.authHandler.LoginAdmin)
	app.Post("/api/v1/admin/register/user", r.authHandler.RegisterUser)

	// Admin User Management routes
	app.Get("/api/v1/admin/users", r.userManagementHandler.ListAllUsers)
	app.Get("/api/v1/admin/user/:user_id", r.userManagementHandler.GetUser)
	app.Post("/api/v1/admin/user/:user_id", r.userManagementHandler.UpdateUser)
	app.Post("/api/v1/admin/user/:user_id/profile", r.userManagementHandler.AdminUpdateUserProfile)
	app.Delete("/api/v1/admin/user/:user_id", r.userManagementHandler.DeleteUser)

	// Create authentication middleware
	authMiddleware := r.tokenManager.AuthMiddleware()

	// Protected routes (authentication required)
	// User profile
	app.Get("/api/v1/profile", authMiddleware, r.profileHandler.GetProfile)
	app.Post("/api/v1/profile", authMiddleware, r.userManagementHandler.UpdateProfile)
	app.Get("/api/v1/profile/resume", authMiddleware, r.userManagementHandler.GetUserResume)

	// Password management
	app.Post("/api/v1/change-password", authMiddleware, r.passwordHandler.ChangePassword)

	// Forgot Password management
	app.Post("/api/v1/forgot-password", authMiddleware, r.forgotPasswordHandler.RequestPasswordReset)
	app.Get("/api/v1/forgot-password-requests", authMiddleware, r.forgotPasswordHandler.GetPendingRequests)
	app.Post("/api/v1/forgot-password-request/:request_id", authMiddleware, r.forgotPasswordHandler.ApproveRejectRequest)

	// AI Chat management (authenticated users only)
	app.Post("/api/v1/chats", authMiddleware, r.chatHandler.ProcessChat)
	app.Get("/api/v1/chats/history", authMiddleware, r.chatHandler.GetConversationHistory)
	app.Delete("/api/v1/chats/history", authMiddleware, r.chatHandler.ClearConversationHistory)

	// Logout endpoint (authenticated users only)
	app.Post("/api/v1/logout", authMiddleware, r.logoutHandler.Logout)

	// Project Management routes
	app.Post("/api/v1/projects", authMiddleware, r.projectHandler.CreateProject)
	app.Get("/api/v1/projects/:id", authMiddleware, r.projectHandler.GetProject)
	app.Get("/api/v1/admin/projects", authMiddleware, r.projectHandler.GetProjectsByAdmin)
	app.Post("/api/v1/projects/:id", authMiddleware, r.projectHandler.UpdateProject)
	app.Delete("/api/v1/projects/:id", authMiddleware, r.projectHandler.DeleteProject)
	app.Post("/api/v1/projects/:id/members", authMiddleware, r.projectHandler.AddMemberToProject)
	app.Get("/api/v1/projects/:id/members", authMiddleware, r.projectHandler.GetProjectMembers)
	app.Delete("/api/v1/projects/:id/members/:userID", authMiddleware, r.projectHandler.RemoveMemberFromProject)
	app.Get("/api/v1/user/projects", authMiddleware, r.projectHandler.GetUserProjects)

	// Task Management routes
	app.Post("/api/v1/tasks", authMiddleware, r.taskHandler.CreateTask)
	app.Get("/api/v1/tasks/:id", authMiddleware, r.taskHandler.GetTask)
	app.Get("/api/v1/projects/:projectID/tasks", authMiddleware, r.taskHandler.GetTasksByProject)
	app.Get("/api/v1/pentester/tasks", authMiddleware, r.taskHandler.GetTasksByPentester)
	app.Post("/api/v1/tasks/:id", authMiddleware, r.taskHandler.UpdateTask)
	app.Post("/api/v1/tasks/:id/assign", authMiddleware, r.taskHandler.AssignTask)
	app.Delete("/api/v1/tasks/:id", authMiddleware, r.taskHandler.DeleteTask)
	app.Post("/api/v1/tasks/:id/close", authMiddleware, r.taskHandler.CloseTask)

	// Task Submission routes (specific routes first!)
	app.Post("/api/v1/task-submission/:task_id", authMiddleware, r.taskSubmissionHandler.SubmitTask)
	app.Get("/api/v1/task-submission/attachments/:attachment_id", authMiddleware, r.taskSubmissionHandler.DownloadTaskSubmissionAttachment)
	app.Post("/api/v1/task-submission/:submission_id/review", authMiddleware, r.taskSubmissionHandler.ReviewSubmission)
	app.Get("/api/v1/task-submission/:task_id", authMiddleware, r.taskSubmissionHandler.GetTaskSubmissions)

	// Vulnerability Management routes
	app.Post("/api/v1/vulnerabilities", authMiddleware, r.vulnerabilityHandler.CreateVulnerability)
	app.Get("/api/v1/vulnerabilities/:id", authMiddleware, r.vulnerabilityHandler.GetVulnerability)
	app.Get("/api/v1/projects/:projectID/vulnerabilities", authMiddleware, r.vulnerabilityHandler.GetVulnerabilitiesByProject)
	app.Get("/api/v1/tasks/:taskID/vulnerabilities", authMiddleware, r.vulnerabilityHandler.GetVulnerabilitiesByTask)
	app.Get("/api/v1/pentester/vulnerabilities", authMiddleware, r.vulnerabilityHandler.GetVulnerabilitiesByPentester)
	app.Get("/api/v1/vulnerabilities/severity/:severity", authMiddleware, r.vulnerabilityHandler.GetVulnerabilitiesBySeverity)
	app.Get("/api/v1/vulnerabilities/status/:status", authMiddleware, r.vulnerabilityHandler.GetVulnerabilitiesByStatus)
	app.Post("/api/v1/vulnerabilities/:id", authMiddleware, r.vulnerabilityHandler.UpdateVulnerability)
	app.Post("/api/v1/vulnerabilities/:id/status", authMiddleware, r.vulnerabilityHandler.UpdateVulnerabilityStatus)
	app.Delete("/api/v1/vulnerabilities/:id", authMiddleware, r.vulnerabilityHandler.DeleteVulnerability)

	// Message Management routes (authenticated users only)
	app.Get("/api/v1/messages", authMiddleware, r.messageHandler.GetChats)
	app.Post("/api/v1/messages/send", authMiddleware, r.messageHandler.SendMessage)
	app.Get("/api/v1/messages/conversation/:user_id", authMiddleware, r.messageHandler.GetConversation)
	app.Get("/api/v1/messages/attachments/:attachment_id", authMiddleware, r.messageHandler.DownloadMessageAttachment)
	app.Post("/api/v1/messages/:id/delivered", authMiddleware, r.messageHandler.MarkMessageAsDelivered)
	app.Post("/api/v1/messages/:id/read", authMiddleware, r.messageHandler.MarkMessageAsRead)
	app.Get("/api/v1/messages/unread/count", authMiddleware, r.messageHandler.GetUnreadCount)
	app.Get("/api/v1/messages/users", authMiddleware, r.messageHandler.GetUsers)
	app.Delete("/api/v1/messages/:id", authMiddleware, r.messageHandler.DeleteMessage)

	// Notification Management routes (authenticated users only)
	app.Post("/api/v1/notifications", authMiddleware, r.notificationHandler.CreateNotification)
	app.Get("/api/v1/notifications", authMiddleware, r.notificationHandler.GetUserNotifications)
	app.Get("/api/v1/notifications/:id/stats", authMiddleware, r.notificationHandler.GetNotificationStats)
	app.Delete("/api/v1/notifications/:id", authMiddleware, r.notificationHandler.DeleteNotification)

	// Activity/Audit Log routes (authenticated users only)
	app.Get("/api/v1/activities/user", authMiddleware, r.activityHandler.GetUserActivities)
	app.Get("/api/v1/activities", authMiddleware, r.activityHandler.GetAllActivities)
	app.Get("/api/v1/activities/recent", authMiddleware, r.activityHandler.GetRecentActivities)
	app.Get("/api/v1/activities/type/:type", authMiddleware, r.activityHandler.GetActivitiesByType)

	// WebSocket route for real-time messaging and notifications
	app.Get("/ws", r.websocketHandler.HandleWebSocket)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})
}
