package user

import (
	"github.com/gofiber/fiber/v2"
	idgenerator "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/id_generator"
)

func SignInWithGuestHandler(c *fiber.Ctx) error {
	uuid := idgenerator.GenerateUUID()

	return c.JSON(guestSignInResponse{
		UID: uuid,
	})
}
