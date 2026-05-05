package routes

import (
	"cinematra-backend/handlers"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App) {
	auth := app.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Get("/google", handlers.GoogleLogin)
	auth.Get("/google/callback", handlers.GoogleCallback)
	auth.Get("/me", handlers.Me)
}
