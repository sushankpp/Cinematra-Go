package middleware

import (
	"cinematra-backend/utils"

	"github.com/gofiber/fiber/v3"
)

func JWTProtected() fiber.Handler {
	return func(c fiber.Ctx) error {
		// read token from cookie
		token := c.Cookies("token")

		if token == "" {
			return c.Status(401).JSON(fiber.Map{"error": "unauthorized - no token"})
		}

		// reuse  existing ValidateJWT
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			return c.Status(401).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		// store userID in context for handlers
		userID, ok := claims["user_id"].(string)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "user_id not found in token"})
		}
		c.Locals("userID", userID)

		return c.Next()
	}
}
