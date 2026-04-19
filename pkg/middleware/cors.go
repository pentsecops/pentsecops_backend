package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware returns a Fiber CORS middleware
func CORSMiddleware() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowHeaders:     "Content-Type, Authorization, Accept, X-Requested-With",
		ExposeHeaders:    "Content-Type, Content-Disposition, Content-Length",
		AllowCredentials: false,
		MaxAge:           3600,
	})
}
