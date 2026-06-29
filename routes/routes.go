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

	videos.Get("/recommendations", middleware.JWTProtected(), handlers.GetRecommendations)
	videos.Get("/trending", handlers.GetGlobalTrending)
	videos.Get("/trending/me", middleware.JWTProtected(), handlers.GetPersonalizedTrending)

	// Wildcard :id routes — these match anything, so they go last
	videos.Get("/:id", handlers.GetVideo)
	videos.Get("/:id/stream-info", handlers.StreamInfo)
	videos.Get("/:id/comments", handlers.GetComments)
	videos.Put("/:id", middleware.JWTProtected(), handlers.UpdateVideo)
	videos.Post("/:id/view", handlers.IncrementVideoViews)
	videos.Delete("/:id", middleware.JWTProtected(), handlers.DeleteVideo)

	//comment routes - posting and deleting comments needs auth, viewing is public
	app.Post("/comments", middleware.JWTProtected(), handlers.PostComment)
	app.Delete("/comments/:id", middleware.JWTProtected(), handlers.DeleteComment)

	//channel route
	app.Get("/channels", middleware.OptionalJWT(), handlers.GetChannels)

	//subcribers routes - subscribing and unsubscribing needs auth, viewing is public
	subscriptions := app.Group("/subscriptions")
	subscriptions.Post("/", middleware.JWTProtected(), handlers.Subscribe)
	subscriptions.Delete("/:channelId", middleware.JWTProtected(), handlers.UnSubscribe)
	subscriptions.Get("/:channelId/count", handlers.GetSubscriberCount)
	subscriptions.Get("/:channelId", middleware.JWTProtected(), handlers.GetSubscriptions)

	//user preference route
	app.Post("/preferences", middleware.JWTProtected(), handlers.SaveUserPreference)
	app.Get("/preferences", middleware.JWTProtected(), handlers.GetUserPreference)

	//watch history route
	app.Get("/watch/history", middleware.JWTProtected(), handlers.GetWatchHistory)
	app.Post("/watch/:id", middleware.JWTProtected(), handlers.WatchVideo)
}
