package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/raksitnongbua/planning-poker-service/internal/repository"
)

type RoomRequest struct {
	RoomName   string `json:"room_name"`
	HostingID  string `json:"hosting_id"`
	DeskConfig string `json:"desk_config"`
}
type RoomResponse struct {
	RoomID    string `json:"room_id"`
	CreatedAt string `json:"created_at"`
}
type GuestSignInResponse struct {
	UID string `json:"uuid"`
}

type MessageAction struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

type JoinRoomPayload struct {
	Name string `json:"name"`
}

type EstimatedPointPayload struct {
	Value string `json:"value"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func HealthCheckHandler(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).SendString("Healthy")
}

func findMemberIndex(members []Member, targetId string) int {
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

func getTimeNow() string {
	return time.Now().Format(time.RFC3339)
}

func CreateNewRoomHandler(c *fiber.Ctx) error {
	var request RoomRequest

	if err := json.Unmarshal(c.Body(), &request); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	fmt.Printf("Request payload %s %s %s\n", request.RoomName, request.HostingID, request.DeskConfig)

	if request.RoomName == "" || request.HostingID == "" || request.DeskConfig == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	roomID := generateUniqueRoomID()
	room := Room{Name: request.RoomName, Status: "VOTING", CreatedAt: getTimeNow(), UpdatedAt: getTimeNow(), DeskConfig: request.DeskConfig}

	docRef := repository.RoomsColRef.Doc(roomID)
	docRef.Set(context.TODO(), room)

	fmt.Printf("Room created: %s (%s)\n", request.RoomName, roomID)

	createdAt := getTimeNow()

	return c.JSON(RoomResponse{
		RoomID:    roomID,
		CreatedAt: createdAt,
	})
}

func SignInWithGuestHandler(c *fiber.Ctx) error {
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

func roomExists(roomId string) bool {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Get(context.TODO())
	return err == nil
}

func foundUser(uid string, members []Member) bool {
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
	room.Members = append(room.Members, Member{
		ID: uid, Name: name, LastActiveAt: getTimeNow(), EstimatedValue: ""})
	room.MemberIDs = append(room.MemberIDs, uid)

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

	Value := updatePointData.Value
	room.Members[index].EstimatedValue = Value
	room.Members[index].LastActiveAt = getTimeNow()

	return true
}

func getResult(members []Member) map[string]int {
	result := make(map[string]int)
	for _, member := range members {
		if member.EstimatedValue != "" {
			result[member.EstimatedValue] = result[member.EstimatedValue] + 1
		}
	}
	return result
}

func SocketRoomHandler(c *websocket.Conn) {
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
	roomInfo := getRoomInfo(roomId)

	c.WriteJSON(MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
	if !foundUser(uid, roomInfo.Members) {
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
		room := getRoomInfo(roomId)

		switch receivedMessage.Action {
		case "JOIN_ROOM":
			if !foundUser(uid, room.Members) && joinRoom(receivedMessage.Payload, uid, &room) {
				room.UpdatedAt = getTimeNow()
				log.Println("room updated:", room)
				docRef := repository.RoomsColRef.Doc(roomId)
				docRef.Set(context.TODO(), room)
				broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
			} else {
				c.WriteJSON(fiber.Map{"error": "JOIN_ROOM_FAILED"})
			}
		case "UPDATE_ACTIVE_USER":
			if foundUser(uid, room.Members) {
				index := findMemberIndex(room.Members, uid)

				if index != -1 {
					room.Members[index].LastActiveAt = getTimeNow()
					docRef := repository.RoomsColRef.Doc(roomId)
					docRef.Set(context.TODO(), room)
					broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
				} else {
					c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
				}
			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}
		case "UPDATE_ESTIMATED_VALUE":
			if foundUser(uid, room.Members) {
				index := findMemberIndex(room.Members, uid)
				if index != -1 && updateEstimatedPoint(receivedMessage.Payload, index, &room) {
					room.Result = getResult(room.Members)
					room.UpdatedAt = getTimeNow()
					docRef := repository.RoomsColRef.Doc(roomId)
					docRef.Set(context.TODO(), room)
					broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
				} else {
					c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
				}
			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}
		case "REVEAL_CARDS":
			if foundUser(uid, room.Members) {
				index := findMemberIndex(room.Members, uid)
				if index != -1 {
					room.Members[index].LastActiveAt = getTimeNow()
				}
			}
			room.UpdatedAt = getTimeNow()
			room.Status = "REVEALED_CARDS"
			docRef := repository.RoomsColRef.Doc(roomId)
			docRef.Set(context.TODO(), room)
			broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})

		case "RESET_ROOM":
			for index := range room.Members {
				room.Members[index].EstimatedValue = ""
			}
			room.Status = "VOTING"
			room.UpdatedAt = getTimeNow()
			room.Result = make(map[string]int)
			docRef := repository.RoomsColRef.Doc(roomId)

			docRef.Update(context.TODO(), []firestore.Update{
				{Path: "Status", Value: room.Status},
				{Path: "UpdatedAt", Value: room.UpdatedAt},
				{Path: "Members", Value: room.Members},
				{Path: "Result", Value: room.Result}})

			broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: room})
		}
	}
}

func getRoomInfo(roomId string) Room {
	docRef := repository.RoomsColRef.Doc(roomId)
	docSnapshot, err := docRef.Get(context.TODO())
	if err != nil {
		log.Fatalf("Failed to get document: %v", err)
	}
	var roomInfo Room
	if err := docSnapshot.DataTo(&roomInfo); err != nil {
		log.Fatalf("Failed to map Firestore document data: %v", err)
	}
	return roomInfo
}

func structToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	val := reflect.ValueOf(obj)
	typ := reflect.TypeOf(obj)

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		tag := typ.Field(i).Tag.Get("json")
		result[tag] = field.Interface()
	}

	return result
}

func GetRecentRoomsHandler(c *fiber.Ctx) error {
	id := c.Params("id")
	query := repository.RoomsColRef.Where("MemberIDs", "array-contains", id).OrderBy("UpdatedAt", firestore.Desc)

	docs, err := query.Documents(c.Context()).GetAll()
	if err != nil {
		log.Fatalf("error get recent rooms: %v", err)
	}

	var rooms []map[string]interface{}
	for _, doc := range docs {
		var room Room
		if err := doc.DataTo(&room); err != nil {
			log.Fatalf("Failed to map Firestore document data: %v", err)
		}
		var newRoom map[string]interface{}
		newRoom = structToMap(room)
		newRoom["id"] = doc.Ref.ID

		rooms = append(rooms, newRoom)
	}

	return c.JSON(fiber.Map{"data": rooms})
}
