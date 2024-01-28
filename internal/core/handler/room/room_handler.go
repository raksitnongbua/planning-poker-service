package room

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	idgenerator "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/id_generator"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
	"github.com/raksitnongbua/planning-poker-service/internal/repository"
	"github.com/raksitnongbua/planning-poker-service/internal/repository/room"
)

func CreateNewRoomHandler(c *fiber.Ctx) error {
	var request RoomRequest

	if err := json.Unmarshal(c.Body(), &request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	fmt.Printf("Request payload %s %s %s\n", request.RoomName, request.HostingID, request.DeskConfig)

	if request.RoomName == "" || request.HostingID == "" || request.DeskConfig == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}
	now := timer.GetTimeNow()
	roomID := idgenerator.GenerateUniqueRoomID()
	room := domain.Room{Name: request.RoomName, Status: "VOTING", CreatedAt: now, UpdatedAt: now, DeskConfig: request.DeskConfig}

	docRef := repository.RoomsColRef.Doc(roomID)
	docRef.Set(context.TODO(), room)

	fmt.Printf("Room created: %s (%s)\n", request.RoomName, roomID)

	createdAt := now

	return c.JSON(RoomResponse{
		RoomID:    roomID,
		CreatedAt: createdAt,
	})
}

func GetRecentRoomsHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	rooms, err := room.QueryRecentRooms(context.TODO(), id)
	if err != nil {
		return c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": rooms})
}
