package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

func healthCheckHandler(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("Healthy")
}

type RoomRequest struct {
	RoomName  string `json:"room_name"`
	HostingID string `json:"hosting_id"`
}

type RoomResponse struct {
	RoomID    string `json:"room_id"`
	CreatedAt string `json:"created_at"`
}
type GuestSignInResponse struct {
	UID string `json:"uuid"`
}

func generateUniqueRoomID() string {
	// Combine random string, and UUID for better uniqueness
	randomString := generateRandomString(5)
	uuid := uuid.New().String()

	return strings.Join([]string{randomString, uuid}, "-")
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	runes := []rune(charset)
	b := make([]rune, length)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}

func generateUUID() string {
	// Combine random string, UUID and timestamp for better uniqueness
	randomString := generateRandomString(5) // Generate a 5-character random string
	uuid := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")

	return strings.Join([]string{randomString, uuid, timestamp}, "-")
}

func createNewRoomHandler(c *fiber.Ctx) error {
	var request RoomRequest

	if err := json.Unmarshal(c.Body(), &request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if request.RoomName == "" || request.HostingID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	roomID := generateUniqueRoomID()

	fmt.Printf("Room created: %s (%s)\n", request.RoomName, roomID)

	createdAt := time.Now().Format(time.RFC3339)

	return c.JSON(RoomResponse{
		RoomID:    roomID,
		CreatedAt: createdAt,
	})
}

func signInWithGuestHandler(c *fiber.Ctx) error {
	uuid := generateUUID()

	fmt.Printf("Guest created: %s\n", uuid)
	return c.JSON(GuestSignInResponse{
		UID: uuid,
	})
}

func main() {
	app := fiber.New()
	app.Use(cors.New())
	app.Get("/health", healthCheckHandler)
	api := app.Group("/api")

	v1 := api.Group("v1")
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).SendString("Api v1 is ready!")
	})
	v1.Get("/guest/sign-in", signInWithGuestHandler)
	v1.Post("/new-room", createNewRoomHandler)

	app.Listen(":8000")
}
