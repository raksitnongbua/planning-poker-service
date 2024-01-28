package roomsocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"cloud.google.com/go/firestore"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
	"github.com/raksitnongbua/planning-poker-service/internal/repository"
	repo "github.com/raksitnongbua/planning-poker-service/internal/repository/room"
)

type MessageAction struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func joinRoom(payload interface{}, uid string, room *domain.Room) bool {

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
		fmt.Println("Error unmarshal payload:", err)
		return false
	}
	name := joinRoomData.Name
	room.Members = append(room.Members, domain.Member{
		ID: uid, Name: name, LastActiveAt: timer.GetTimeNow(), EstimatedValue: ""})
	room.MemberIDs = append(room.MemberIDs, uid)

	return true
}
func broadcastMessage(roomId string, message interface{}) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		clientRoomID := client.Locals("roomId")
		if clientRoomID != roomId {
			continue
		}

		err := client.WriteJSON(message)
		if err != nil {
			log.Printf("Error sending message to client: %v", err)
		}
	}
}

func SocketRoomHandler(c *websocket.Conn) {
	roomId := c.Params("id")

	if !repo.RoomExists(roomId) {
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
	roomInfo := repo.GetRoomInfo(roomId)

	c.WriteJSON(MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
	if !room.IsUserInRoom(uid, roomInfo.Members) {
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
		roomInfo := repo.GetRoomInfo(roomId)
		now := timer.GetTimeNow()
		isUserInRoom := room.IsUserInRoom(uid, roomInfo.Members)
		switch receivedMessage.Action {
		case "JOIN_ROOM":
			if !isUserInRoom && joinRoom(receivedMessage.Payload, uid, &roomInfo) {
				roomInfo.UpdatedAt = now
				log.Println("room updated:", roomInfo)
				docRef := repository.RoomsColRef.Doc(roomId)
				docRef.Set(context.TODO(), roomInfo)
				broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
			} else {
				c.WriteJSON(fiber.Map{"error": "JOIN_ROOM_FAILED"})
			}
		case "UPDATE_ACTIVE_USER":
			if isUserInRoom {
				index := room.FindMemberIndex(roomInfo.Members, uid)

				if index != -1 {
					roomInfo.Members[index].LastActiveAt = now
					docRef := repository.RoomsColRef.Doc(roomId)
					docRef.Set(context.TODO(), roomInfo)
					broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
				} else {
					c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
				}
			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}
		case "UPDATE_ESTIMATED_VALUE":
			if isUserInRoom {
				index := room.FindMemberIndex(roomInfo.Members, uid)
				if index != -1 {
					estimatedPayload, err := TransformPayloadToEstimatedPoint(receivedMessage.Payload)
					if err != nil {
						c.WriteJSON(fiber.Map{"error": "INVALID_PAYLOAD"})
						return
					}
					updatedMembers := room.UpdateEstimatedValue(roomInfo.Members, index, estimatedPayload.Value)
					calculatedResult := room.CalculateResult(roomInfo.Members)

					if repo.UpdateEstimatedValue(roomId, updatedMembers, calculatedResult) != nil {
						c.WriteJSON(fiber.Map{"error": "UPDATE_ESTIMATED_VALUE_FAILED"})
						return
					}

					broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
				} else {
					c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
				}
			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}
		case "REVEAL_CARDS":
			if isUserInRoom {
				index := room.FindMemberIndex(roomInfo.Members, uid)
				if index != -1 {
					roomInfo.Members[index].LastActiveAt = now
				}
			}
			roomInfo.UpdatedAt = now
			roomInfo.Status = "REVEALED_CARDS"
			docRef := repository.RoomsColRef.Doc(roomId)
			docRef.Set(context.TODO(), roomInfo)
			broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})

		case "RESET_ROOM":
			for index := range roomInfo.Members {
				roomInfo.Members[index].EstimatedValue = ""
			}
			roomInfo.Status = "VOTING"
			roomInfo.UpdatedAt = now
			roomInfo.Result = make(map[string]int)
			docRef := repository.RoomsColRef.Doc(roomId)

			docRef.Update(context.TODO(), []firestore.Update{
				{Path: "Status", Value: roomInfo.Status},
				{Path: "UpdatedAt", Value: roomInfo.UpdatedAt},
				{Path: "Members", Value: roomInfo.Members},
				{Path: "Result", Value: roomInfo.Result}})

			broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
		}
	}
}
