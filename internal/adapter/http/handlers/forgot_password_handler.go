package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"

	"github.com/pentsecops/backend/internal/core/usecases"
	"github.com/pentsecops/backend/pkg/middleware"
)

type ForgotPasswordHandler struct {
	forgotPasswordUseCase *usecases.ForgotPasswordUseCase
	tokenManager          *middleware.TokenManager
}

func NewForgotPasswordHandler(
	forgotPasswordUseCase *usecases.ForgotPasswordUseCase,
	tokenManager *middleware.TokenManager,
) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		forgotPasswordUseCase: forgotPasswordUseCase,
		tokenManager:          tokenManager,
	}
}

// RequestPasswordReset handles POST /api/v1/forgot-password
func (fph *ForgotPasswordHandler) RequestPasswordReset(c *fiber.Ctx) error {
	fmt.Println("[API] POST /api/v1/forgot-password")

	type Request struct {
		Email string `json:"email"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("[ERROR] Failed to parse request body: %v\n", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Email == "" {
		fmt.Println("[INFO] Email is required")
		return c.Status(400).JSON(fiber.Map{
			"error": "Email is required",
		})
	}

	err := fph.forgotPasswordUseCase.RequestPasswordReset(req.Email)
	if err != nil {
		fmt.Printf("[ERROR] Failed to process password reset request: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to process request",
		})
	}

	fmt.Println("[SUCCESS] Password reset request submitted")
	return c.Status(200).JSON(fiber.Map{
		"message": "Request Escalated to the Administration ,If email exists, you will receive a password reset request",
	})
}

// GetPendingRequests handles GET /api/v1/forgot-password-requests (admin only)
func (fph *ForgotPasswordHandler) GetPendingRequests(c *fiber.Ctx) error {
	fmt.Println("[API] GET /api/v1/forgot-password-requests")

	// Verify admin token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		fmt.Println("[ERROR] Authorization header is missing")
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Extract token from "Bearer <token>" format
	const bearerSchema string = "Bearer "
	token := authHeader
	if len(authHeader) > len(bearerSchema) && authHeader[:len(bearerSchema)] == bearerSchema {
		token = authHeader[len(bearerSchema):]
	}

	claims, err := fph.tokenManager.VerifyToken(token)
	if err != nil {
		fmt.Printf("[ERROR] Failed to verify token: %v\n", err)
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if claims.Role != "admin" {
		fmt.Println("[INFO] Non-admin user attempted to access admin endpoint")
		return c.Status(403).JSON(fiber.Map{
			"error": "Forbidden - Admin access required",
		})
	}

	requests, err := fph.forgotPasswordUseCase.GetPendingRequests()
	if err != nil {
		fmt.Printf("[ERROR] Failed to fetch pending requests: %v\n", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch requests",
		})
	}

	fmt.Printf("[SUCCESS] Retrieved %d pending password reset requests\n", len(requests))
	return c.Status(200).JSON(fiber.Map{
		"requests": requests,
	})
}

// ApproveRejectRequest handles POST /api/v1/forgot-password-request/:request_id (admin only)
func (fph *ForgotPasswordHandler) ApproveRejectRequest(c *fiber.Ctx) error {
	fmt.Println("[API] POST /api/v1/forgot-password-request/:request_id")

	// Verify admin token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		fmt.Println("[ERROR] Authorization header is missing")
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Extract token from "Bearer <token>" format
	const bearerSchema string = "Bearer "
	token := authHeader
	if len(authHeader) > len(bearerSchema) && authHeader[:len(bearerSchema)] == bearerSchema {
		token = authHeader[len(bearerSchema):]
	}

	claims, err := fph.tokenManager.VerifyToken(token)
	if err != nil {
		fmt.Printf("[ERROR] Failed to verify token: %v\n", err)
		return c.Status(401).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	if claims.Role != "admin" {
		fmt.Println("[INFO] Non-admin user attempted to access admin endpoint")
		return c.Status(403).JSON(fiber.Map{
			"error": "Forbidden - Admin access required",
		})
	}

	requestID := c.Params("request_id")
	if requestID == "" {
		fmt.Println("[INFO] Request ID is required")
		return c.Status(400).JSON(fiber.Map{
			"error": "Request ID is required",
		})
	}

	type Request struct {
		Status string `json:"status"`
	}

	var req Request
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("[ERROR] Failed to parse request body: %v\n", err)
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Status != "accepted" && req.Status != "rejected" {
		fmt.Println("[INFO] Invalid status provided")
		return c.Status(400).JSON(fiber.Map{
			"error": "Status must be 'accepted' or 'rejected'",
		})
	}

	if req.Status == "accepted" {
		// Approve the request and generate password
		result, err := fph.forgotPasswordUseCase.ApproveRequest(requestID)
		if err != nil {
			fmt.Printf("[ERROR] Failed to approve request: %v\n", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		fmt.Printf("[SUCCESS] Password reset request %s approved\n", requestID)
		return c.Status(200).JSON(fiber.Map{
			"message": "Request approved",
			"request": result,
		})
	} else {
		// Reject the request
		err := fph.forgotPasswordUseCase.RejectRequest(requestID)
		if err != nil {
			fmt.Printf("[ERROR] Failed to reject request: %v\n", err)
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		fmt.Printf("[SUCCESS] Password reset request %s rejected\n", requestID)
		return c.Status(200).JSON(fiber.Map{
			"message": "Request rejected",
		})
	}
}
