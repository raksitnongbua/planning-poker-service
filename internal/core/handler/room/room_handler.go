package room

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/constants"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/profile"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
)

func CreateNewRoomHandler(c *fiber.Ctx) error {
	var req RoomRequest

	if err := json.Unmarshal(c.Body(), &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.RoomName == "" || req.HostingID == "" || req.DeskConfig == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}
	now := timer.GetTimeNow()

	roomID, err := room.CreateNewRoom(req.RoomName, req.DeskConfig)
	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(RoomResponse{
		RoomID:    roomID,
		CreatedAt: now,
	})
}

func GetRecentRoomsHandler(c *fiber.Ctx) error {
	var id string
	id = c.Params("id") // Guest Id fallback
	session := c.Cookies(constants.NextAuthSessionCookie)
	if session != "" {
		p, err := profile.GetProfile(session)
		if err == nil {
			id = p.UID
		}
	}
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	rooms, err := room.GetResendRooms(id)
	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": rooms})
}
