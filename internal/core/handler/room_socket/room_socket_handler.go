package roomsocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
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

func noticeUpdateRoom(roomId string, roomInfo domain.Room) {
	broadcastMessage(roomId, MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
}

func SocketRoomHandler(c *websocket.Conn) {
	roomId := c.Params("id")

	if !room.IsRoomExists(roomId) {
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
	roomInfo := room.GetRoomInfo(roomId)

	c.WriteJSON(MessageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
	if !room.IsUserInRoomWithId(uid, roomId) {
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
		var receivedMessage MessageAction
		if err := json.Unmarshal(msg, &receivedMessage); err != nil {
			log.Println("json unmarshal:", err)
			break
		}

		switch receivedMessage.Action {
		case "JOIN_ROOM":
			joinRoomPayload, err := TransformPayloadToJoinRoom(receivedMessage.Payload)
			if err != nil {
				c.WriteJSON(fiber.Map{"error": "INVALID_PAYLOAD"})
				return
			}
			roomInfo, err := room.JoinRoom(joinRoomPayload.Name, uid, roomId)
			if err != nil {
				c.WriteJSON(fiber.Map{"error": "JOIN_ROOM_FAILED"})
				return
			}
			noticeUpdateRoom(roomId, roomInfo)

		case "UPDATE_ESTIMATED_VALUE":
			index := room.FindMemberIndex(roomInfo.Members, uid)
			if index != -1 {
				estimatedPayload, err := TransformPayloadToEstimatedPoint(receivedMessage.Payload)
				if err != nil {
					c.WriteJSON(fiber.Map{"error": "INVALID_PAYLOAD"})
					return
				}
				roomInfo, err := room.UpdateEstimatedValue(index, estimatedPayload.Value, roomId)

				if err != nil {
					c.WriteJSON(fiber.Map{"error": "UPDATE_ESTIMATED_VALUE_FAILED"})
					return
				}
				noticeUpdateRoom(roomId, roomInfo)

			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}
		case "REVEAL_CARDS":
			index := room.FindMemberIndex(roomInfo.Members, uid)
			if index != -1 {
				roomInfo, err := room.RevealCards(index, roomId)
				if err != nil {
					c.WriteJSON(fiber.Map{"error": "REVEAL_CARDS_FAILED"})
					return
				}

				noticeUpdateRoom(roomId, roomInfo)
			}

		case "RESET_ROOM":
			roomInfo, err := room.ResetRoom(roomId)
			if err != nil {
				c.WriteJSON(fiber.Map{"error": "RESET_ROOM_FAILED"})
				return
			}
			noticeUpdateRoom(roomId, roomInfo)

		}
	}
}
