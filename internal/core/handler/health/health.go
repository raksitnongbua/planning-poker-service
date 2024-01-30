package health

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func HealthCheckHandler(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("Healthy")
}
