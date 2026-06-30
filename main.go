package main

import (
	"cinematra-backend/config"
	"cinematra-backend/middleware"
	"cinematra-backend/routes"
	"cinematra-backend/utils/moderation"
	"cinematra-backend/utils/recommend"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()
	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024, // 10 MB
	})
	middleware.SetupCors(app)
	app.Get("/media/*", static.New("./storage"))
	routes.SetupRoutes(app)

	if err := moderation.LoadGlobalModel("utils/moderation/content_filter_model.json"); err != nil {
		log.Fatal("Error loading content filter model:", err)
	}
	recommend.StartBackgroundRefresh(10 * time.Minute)

	log.Fatal(app.Listen(":8080"))
}
