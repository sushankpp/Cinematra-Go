package routes

import (
	"cinematra-backend/handlers"
	"cinematra-backend/middleware"

	"github.com/gofiber/fiber/v3"
)

func SetupRoutes(app *fiber.App) {
	//auth routes
	auth := app.Group("/auth")
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Post("/logout", handlers.Logout)
	auth.Get("/google", handlers.GoogleLogin)
	auth.Get("/google/callback", handlers.GoogleCallback)
	auth.Get("/me", handlers.Me)

	// video routes - upload needs auth , viewing is public
	videos := app.Group("/videos")
	videos.Post("/upload", middleware.JWTProtected(), handlers.UploadVideo)
	app.Get("/videos/:id", handlers.GetVideo)
	app.Get("/videos", handlers.GetVideos)
}
