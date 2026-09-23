package middleware

import (
	"cinematra-backend/utils"

	"github.com/gofiber/fiber/v3"
)

func JWTProtected() fiber.Handler {
	return func(c fiber.Ctx) error {
		// read token from cookie or Authorization header
		token := c.Cookies("token")
		if token == "" {
			authHeader := c.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}
		}

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

		// store role in context for admin checks
		role, _ := claims["role"].(string)
		if role == "" {
			role = "user"
		}
		c.Locals("userRole", role)

		return c.Next()
	}
}

func OptionalJWT() fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Cookies("token")
		if token == "" {
			authHeader := c.Get("Authorization")
			if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token = authHeader[7:]
			}
		}
		if token == "" {
			return c.Next()
		}

		claims, err := utils.ValidateJWT(token)
		if err != nil {
			// fmt.Println("OptionalJWT: token invalid:", err)
			return c.Next()
		}
		// fmt.Println("OptionalJWT claims:", claims)

		if userID, ok := claims["user_id"].(string); ok {
			c.Locals("userID", userID)
		}
		return c.Next()
	}
}
