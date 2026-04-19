package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/pentsecops/backend/internal/adapter/http/handlers"
	"github.com/pentsecops/backend/internal/adapter/http/routes"
	"github.com/pentsecops/backend/internal/adapter/repositories"
	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/database"
	"github.com/pentsecops/backend/pkg/middleware"
	"github.com/pentsecops/backend/pkg/services"
	"github.com/pentsecops/backend/pkg/websocket"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("[WARN] .env file not found, using system environment variables")
	}

	// Load environment variables or use defaults for development
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "postgres"
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "postgres"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "pentsecops"
	}

	// Connect to database
	// For Supabase: disable prepared statement caching entirely to avoid connection pool issues
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require prepared_statement_cache_mode=off binary_parameters=no",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure connection pool for Supabase
	// Smaller pool to reduce connection issues with Railways/Supabase connection limits
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetConnMaxIdleTime(time.Minute * 1)

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("✓ Database connection established")

	// Run database migrations
	migrator := database.NewMigrator(db)
	if err := migrator.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize token manager with PASETO
	publicKeyStr := os.Getenv("PASETO_PUBLIC_KEY")
	privateKeyStr := os.Getenv("PASETO_PRIVATE_KEY")

	// If keys don't exist, generate new ones
	if publicKeyStr == "" || privateKeyStr == "" {
		log.Println("⚠ PASETO keys not found in environment. Generating new keys...")
		pub, priv, err := middleware.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate PASETO keys: %v", err)
		}
		publicKeyStr = pub
		privateKeyStr = priv
		log.Println("⚠ Generated PASETO keys (store these in your .env file):")
		log.Printf("PASETO_PUBLIC_KEY=%s\n", publicKeyStr)
		log.Printf("PASETO_PRIVATE_KEY=%s\n", privateKeyStr)
	}

	tokenManager, err := middleware.NewTokenManager(publicKeyStr, privateKeyStr, 24*time.Hour)
	if err != nil {
		log.Fatalf("Failed to initialize token manager: %v", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	adminRepo := repositories.NewAdminRepository(db)
	pentesterRepo := repositories.NewPentesterRepository(db)
	stakeholderRepo := repositories.NewStakeholderRepository(db)
	forgotPasswordRepo := repositories.NewForgotPasswordRequestRepository(db)
	chatRepo := repositories.NewChatRepository(db)
	logoutRepo := repositories.NewLogoutRepository(db)
	projectRepo := repositories.NewProjectRepository(db)
	projectMemberRepo := repositories.NewProjectMemberRepository(db)
	taskRepo := repositories.NewTaskRepository(db)
	taskSubmissionRepo := repositories.NewTaskSubmissionRepository(db)
	taskSubmissionAttachmentRepo := repositories.NewTaskSubmissionAttachmentRepository(db)
	taskSubmissionReviewRepo := repositories.NewTaskSubmissionReviewRepository(db)
	vulnerabilityRepo := repositories.NewVulnerabilityRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	messageAttachmentRepo := repositories.NewMessageAttachmentRepository(db)
	notificationRepo := repositories.NewNotificationRepository(db)
	notificationDeliveryLogRepo := repositories.NewNotificationDeliveryLogRepository(db)
	activityRepo := repositories.NewActivityRepository(db)

	// Initialize email service
	emailService, _ := services.NewEmailService()
	if emailService != nil && emailService.IsEnabled() {
		log.Println("✓ Email service initialized and enabled")
	}

	// Initialize Groq AI service
	groqService, _ := services.NewGroqService()
	if groqService != nil && groqService.IsEnabled() {
		log.Println("✓ Groq AI service initialized and enabled")
	}

	// Initialize file service for uploads
	fileService := services.NewFileService("storage/resumes", "storage/attachments")
	log.Println("✓ File service initialized")

	// Initialize use cases
	authUseCase := usecases.NewAuthUseCase(
		userRepo,
		adminRepo,
		pentesterRepo,
		stakeholderRepo,
		tokenManager,
		emailService,
	)

	profileUseCase := usecases.NewProfileUseCase(
		userRepo,
		pentesterRepo,
		stakeholderRepo,
	)

	userManagementUseCase := usecases.NewUserManagementUseCase(
		userRepo,
		pentesterRepo,
		stakeholderRepo,
	)

	passwordUseCase := usecases.NewPasswordUseCase(
		userRepo,
		adminRepo,
	)

	forgotPasswordUseCase := usecases.NewForgotPasswordUseCase(
		forgotPasswordRepo,
		userRepo,
		pentesterRepo,
		stakeholderRepo,
		emailService,
	)

	chatUseCase := usecases.NewChatUseCase(chatRepo, groqService, userRepo)
	logoutUseCase := usecases.NewLogoutUseCase(logoutRepo)
	projectUseCase := usecases.NewProjectUseCase(projectRepo, projectMemberRepo, userRepo)
	taskUseCase := usecases.NewTaskUseCase(taskRepo, projectRepo, pentesterRepo)
	taskSubmissionUseCase := usecases.NewTaskSubmissionUseCase(taskSubmissionRepo, taskSubmissionAttachmentRepo, taskSubmissionReviewRepo, taskRepo, pentesterRepo)
	vulnerabilityUseCase := usecases.NewVulnerabilityUseCase(vulnerabilityRepo, projectRepo, taskRepo, pentesterRepo)
	pentesterUseCase := usecases.NewPentesterUseCase(pentesterRepo)
	messageUseCase := usecases.NewMessageUseCase(messageRepo, messageAttachmentRepo, userRepo)
	activityUseCase := usecases.NewActivityUseCase(activityRepo, userRepo)

	// Create notification use case with email service if available
	var notificationUseCase *usecases.NotificationUseCase
	if emailService != nil && emailService.IsEnabled() {
		notificationUseCase = usecases.NewNotificationUseCaseWithEmail(notificationRepo, notificationDeliveryLogRepo, userRepo, adminRepo, emailService)
		log.Println("✓ Notification use case initialized with email service")
	} else {
		notificationUseCase = usecases.NewNotificationUseCase(notificationRepo, notificationDeliveryLogRepo, userRepo, adminRepo)
		log.Println("⚠ Notification use case initialized without email service")
	}

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authUseCase, activityUseCase)
	profileHandler := handlers.NewProfileHandler(profileUseCase, authUseCase)
	userManagementHandler := handlers.NewUserManagementHandler(userManagementUseCase, tokenManager, fileService, pentesterUseCase)
	passwordHandler := handlers.NewPasswordHandler(passwordUseCase, tokenManager)
	forgotPasswordHandler := handlers.NewForgotPasswordHandler(forgotPasswordUseCase, tokenManager)
	chatHandler := handlers.NewChatHandler(chatUseCase, tokenManager)
	logoutHandler := handlers.NewLogoutHandler(logoutUseCase, tokenManager, activityUseCase)
	projectHandler := handlers.NewProjectHandler(projectUseCase, authUseCase)
	taskHandler := handlers.NewTaskHandlerWithServices(taskUseCase, authUseCase, fileService, activityUseCase)
	taskSubmissionHandler := handlers.NewTaskSubmissionHandler(taskSubmissionUseCase, authUseCase, fileService)
	taskSubmissionHandler.SetAttachmentRepository(taskSubmissionAttachmentRepo)
	vulnerabilityHandler := handlers.NewVulnerabilityHandler(vulnerabilityUseCase, authUseCase)
	messageHandler := handlers.NewMessageHandlerWithAttachments(messageUseCase, userRepo, messageRepo, messageAttachmentRepo, fileService)

	// Create notification handler with email service if available
	var notificationHandler *handlers.NotificationHandler
	if emailService != nil && emailService.IsEnabled() {
		notificationHandler = handlers.NewNotificationHandlerWithEmail(notificationUseCase, emailService)
	} else {
		notificationHandler = handlers.NewNotificationHandler(notificationUseCase)
	}

	activityHandler := handlers.NewActivityHandler(activityUseCase)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "PentSecOps Backend",
	})

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub(30*time.Second, 60*time.Second)
	go wsHub.Run()
	log.Println("✓ WebSocket Hub initialized and started")

	// Initialize WebSocket Service
	wsService := services.NewWebSocketService(wsHub, messageUseCase, notificationUseCase)
	log.Println("✓ WebSocket Service initialized")

	// Initialize WebSocket Handler
	websocketHandler := handlers.NewWebSocketHandler(wsHub, tokenManager, wsService)
	log.Println("✓ WebSocket Handler initialized")

	// Initialize router
	router := routes.NewRouter(authHandler, profileHandler, userManagementHandler, passwordHandler, forgotPasswordHandler, chatHandler, logoutHandler, projectHandler, taskHandler, taskSubmissionHandler, vulnerabilityHandler, messageHandler, notificationHandler, activityHandler, websocketHandler, tokenManager)
	router.SetupRoutes(app)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("✓ Server starting on http://localhost%s\n", addr)

	if err := app.Listen(addr); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
