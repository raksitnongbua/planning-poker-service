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
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	LastActiveAt   string  `json:"last_active_at"`
	EstimatedPoint float32 `json:"estimated_point"`
}
type Room struct {
	Name    string `json:"name"`
	Members []User `json:"members"`
}
type MessageAction struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

type JoinRoomPayload struct {
	Name string `json:"name"`
}

type EstimatedPointPayload struct {
	Point float32 `json:"point"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
	//TODO: sync with database for solved problems serverless when inactivated state
	rooms = make(map[string]Room)
)

func healthCheckHandler(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("Healthy")
}

func findMemberIndex(members []User, targetId string) int {
	for i, user := range members {
		if user.ID == targetId {
			return i
		}
	}
	return -1
}

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
		fmt.Println("Invalid payload format for JOIN_ROOM action.")
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
	room.Members = append(room.Members, User{
		ID: uid, Name: name, LastActiveAt: time.Now().Format(time.RFC3339), EstimatedPoint: -1})

	return true
}

func updateEstimatedPoint(payload interface{}, index int, room *Room) bool {
	updatePointPayload, ok := payload.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid payload format for UPDATE_POINT action.")
		return false
	}

	var updatePointData EstimatedPointPayload
	payloadBytes, err := json.Marshal(updatePointPayload)
	if err != nil {
		fmt.Println("Error marshaling payload:", err)
		return false
	}

	err = json.Unmarshal(payloadBytes, &updatePointData)
	if err != nil {
		fmt.Println("Error unmarshaling payload:", err)
		return false
	}

	point := updatePointData.Point
	room.Members[index].EstimatedPoint = point
	room.Members[index].LastActiveAt = time.Now().Format(time.RFC3339)

	return true
}

func handleRoomSocket(c *websocket.Conn) {
	roomId := c.Params("id")

	if !roomExists(roomId) {
		c.WriteJSON(fiber.Map{"error": "Room not found"})
		log.Printf("Room with ID %s not found", roomId)
		c.Close()
		return
	}
	uid := c.Params("uid")

	c.Locals("roomId", roomId)

	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, c)
		clientsMu.Unlock()

		_ = c.Close()
	}()

	c.WriteJSON(MessageAction{Action: "UPDATE_ROOM", Payload: rooms[roomId]})
	if !foundUser(uid, rooms[roomId].Members) {
		c.WriteJSON(MessageAction{Action: "NEED_TO_JOIN"})
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
				broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
			} else {
				c.WriteJSON(fiber.Map{"error": "JOIN_ROOM_FAILED"})
			}
		case "UPDATE_ACTIVE_USER":
			if foundUser(uid, rooms[roomId].Members) {
				room := rooms[roomId]
				index := findMemberIndex(room.Members, uid)

				if index != -1 {
					room.Members[index].LastActiveAt = time.Now().Format(time.RFC3339)
					rooms[roomId] = room

					broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
				} else {
					c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
				}
			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}

		case "UPDATE_ESTIMATED_POINT":
			if foundUser(uid, rooms[roomId].Members) {
				room := rooms[roomId]
				index := findMemberIndex(room.Members, uid)

				if index != -1 && updateEstimatedPoint(receivedMessage.Payload, index, &room) {
					rooms[roomId] = room
					broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
				} else {
					c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
				}
			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
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
