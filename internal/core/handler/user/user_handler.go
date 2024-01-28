package user

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	idgenerator "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/id_generator"
)

func SignInWithGuestHandler(c *fiber.Ctx) error {
	uuid := idgenerator.GenerateUUID()

	fmt.Printf("Guest created: %s\n", uuid)
	return c.JSON(GuestSignInResponse{
		UID: uuid,
	})
}