package main

import (
	"cinematra-backend/config"
	"cinematra-backend/middleware"
	"cinematra-backend/routes"
	"log"

	"github.com/gofiber/fiber/v3"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	config.ConnectDB()
	app := fiber.New()
	middleware.SetupCors(app)
	routes.SetupRoutes(app)

	log.Fatal(app.Listen(":8080"))
}
