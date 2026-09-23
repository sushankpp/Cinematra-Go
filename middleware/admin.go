package middleware

import (
	"github.com/gofiber/fiber/v3"
)

// RequireSuperAdmin blocks requests from non-superadmin users.
// Must be used after JWTProtected() so that c.Locals("userRole") is set.
func RequireSuperAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		role, _ := c.Locals("userRole").(string)
		if role != "superadmin" {
			return c.Status(403).JSON(fiber.Map{"error": "forbidden - superadmin access required"})
		}
		return c.Next()
	}
}
