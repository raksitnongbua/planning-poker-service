package room

import (
	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/constants"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/profile"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
)

func CreateNewRoomHandler(c *fiber.Ctx) error {
	req, err := unmarshalRoomRequest(c.Body())
	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"error": err.Error()})
	}

	if req.RoomName == "" || req.HostingID == "" || req.DeskConfig == "" {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"error": "Missing required fields"})
	}
	roomID, err := room.CreateNewRoom(req.RoomName, req.DeskConfig)
	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(roomResponse{
		RoomID:    roomID,
		CreatedAt: timer.GetTimeNow(),
	})
}

func CleanupExpiredRoomsHandler(c *fiber.Ctx) error {
	result, err := room.CleanupExpiredRooms()
	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}

func KickMemberHandler(c *fiber.Ctx) error {
	roomId := c.Params("roomId")
	memberID := c.Params("memberId")
	if roomId == "" || memberID == "" {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{"error": "Missing required fields"})
	}

	roomInfo, err := room.KickMember(roomId, memberID)
	if err != nil {
		if err.Error() == "member not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": roomInfo})
}

func GetRecentRoomsHandler(c *fiber.Ctx) error {
	var id string
	id = c.Params("id") // Guest Id fallback

	var cookieKey string
	if c.Secure() {
		cookieKey = constants.SecureSessionCookie
	} else {
		cookieKey = constants.SessionCookie
	}
	sessionCookie := c.Cookies(cookieKey)
	if sessionCookie != "" {
		p, err := profile.GetProfile(sessionCookie)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}
		id = p.UID
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
