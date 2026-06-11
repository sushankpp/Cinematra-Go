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
	auth.Get("/me", middleware.JWTProtected(), handlers.Me)
	auth.Patch("/update", middleware.JWTProtected(), handlers.UpdateProfile)

	// video routes - upload needs auth , viewing is public
	videos := app.Group("/videos")
	videos.Get("/", handlers.GetVideos)
	videos.Get("/mine", middleware.JWTProtected(), handlers.GetUserVideo)
	videos.Post("/upload", middleware.JWTProtected(), handlers.UploadVideo)
	videos.Get("/:id", handlers.GetVideo)
	videos.Get("/:id/stream-info", handlers.StreamInfo)
	videos.Put("/:id", middleware.JWTProtected(), handlers.UpdateVideo)
	videos.Post("/:id/view", handlers.IncrementVideoViews)

	//comment routes - posting and deleting comments needs auth, viewing is public
	app.Get("/videos/:id/comments", handlers.GetComments)
	app.Post("/comments", middleware.JWTProtected(), handlers.PostComment)
	app.Delete("/comments/:id", middleware.JWTProtected(), handlers.DeleteComment)

	//subcribers routes - subscribing and unsubscribing needs auth, viewing is public
	subscriptions := app.Group("/subscriptions")
	subscriptions.Post("/", middleware.JWTProtected(), handlers.Subscribe)
	subscriptions.Delete("/:channelId", middleware.JWTProtected(), handlers.UnSubscribe)
	subscriptions.Get("/:channelId/count", handlers.GetSubscriberCount)
	subscriptions.Get("/:channelId", middleware.JWTProtected(), handlers.GetSubscriptions)

}
