package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
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

type User struct {
	ID   string
	Name string
}
type Room struct {
	Name    string
	Members []User
}
type MessageAction struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

type JoinRoomPayload struct {
	Name string `json:"name"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	rooms     = make(map[string]Room)
)

func generateUniqueRoomID() string {
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
	randomString := generateRandomString(5)
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
	rooms[roomID] = Room{Name: request.RoomName}

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

func broadcastMessage(roomId string, message any) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		cRoomId := client.Locals("roomId")
		if cRoomId != roomId {
			continue
		}

		err := client.WriteJSON(message)

		if err != nil {
			log.Printf("Error sending message to client: %v", err)
		}
	}
}

func broadcastRoomInfo(roomId string) {
	broadcastMessage(roomId, rooms[roomId])
}

func roomExists(roomId string) bool {
	room, exists := rooms[roomId]
	log.Println(room)
	return exists
}

func foundUser(uid string, members []User) bool {
	found := false
	for _, user := range members {
		if user.ID == uid {
			found = true
			break
		}
	}
	return found
}
func joinRoom(payload interface{}, uid string, room *Room) bool {
	joinRoomPayload, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid payload format for REGISTER action.")
		return false
	}

	var joinRoomData JoinRoomPayload
	payloadBytes, err := json.Marshal(joinRoomPayload)
	if err != nil {
		fmt.Println("Error marshaling payload:", err)
		return false
	}

	err = json.Unmarshal(payloadBytes, &joinRoomData)
	if err != nil {
		fmt.Println("Error unmarshaling payload:", err)
		return false
	}

	name := joinRoomData.Name
	room.Members = append(room.Members, User{ID: uid, Name: name})

	return true
}

func leaveRoom(uid string, roomId string) bool {
	var filteredMembers []User
	for _, user := range rooms[roomId].Members {
		if user.ID != uid {
			filteredMembers = append(filteredMembers, user)
		}
	}

	room := rooms[roomId]
	room.Members = filteredMembers
	rooms[roomId] = room

	log.Printf("(%s) leaved room(%s) %s", uid, roomId, filteredMembers)
	return true
}

func handleRoomSocket(c *websocket.Conn) {
	roomId := c.Params("id")

	if !roomExists(roomId) {
		c.WriteJSON("Room not found")
		log.Printf("Room with ID %s not found", roomId)
		c.Close()
		return
	}
	uid := c.Params("uid")

	c.Locals("roomId", roomId)
	c.WriteJSON(MessageAction{Action: "UPDATE_ROOM", Payload: rooms[roomId]})
	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, c)
		clientsMu.Unlock()
		if foundUser(uid, rooms[roomId].Members) {
			leaveRoom(uid, roomId)
		} else {
			log.Printf("user unregistered leaved room(%s)", roomId)
		}
		_ = c.Close()
	}()

	if !foundUser(uid, rooms[roomId].Members) {
		c.WriteJSON(fiber.Map{"action": "NEED_TO_JOIN"})
	}

	var (
		msg []byte
		err error
	)
	for {
		if _, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("recv: %s", msg)
		var receivedMessage MessageAction
		if err := json.Unmarshal(msg, &receivedMessage); err != nil {
			log.Println("json unmarshal:", err)
			break
		}

		switch receivedMessage.Action {
		case "JOIN_ROOM":
			room := rooms[roomId]

			if !foundUser(uid, rooms[roomId].Members) && joinRoom(receivedMessage.Payload, uid, &room) {
				rooms[roomId] = room
				c.WriteJSON(room)

				broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
			} else {
				c.WriteJSON(fiber.Map{"error": "JOIN_ROOM_FAILED"})
			}
		}

	}
}

func main() {
	app := fiber.New()
	app.Use(cors.New())
	app.Get("/health", healthCheckHandler)
	api := app.Group("/api")

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/room/:uid/:id", websocket.New(handleRoomSocket))

	v1 := api.Group("v1")
	v1.Get("/", func(c *fiber.Ctx) error {
		return c.Status(http.StatusOK).SendString("Api v1 is ready!")
	})
	v1.Get("/guest/sign-in", signInWithGuestHandler)
	v1.Post("/new-room", createNewRoomHandler)

	app.Listen(":3001")
}
