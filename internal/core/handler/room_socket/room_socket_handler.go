package roomsocket

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	roomService "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room"
	socketService "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room_socket"
)

type messageAction struct {
	Action  string      `json:"action"`
	Payload interface{} `json:"payload"`
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

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
	broadcastMessage(roomId, messageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
}

func SocketRoomHandler(c *websocket.Conn) {
	roomId := c.Params("id")

	if !roomService.IsRoomExists(roomId) {
		c.WriteJSON(fiber.Map{"error": "Room not found"})
		log.Printf("Room with ID %s not found", roomId)
		c.Close()
		return
	}

	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, c)
		clientsMu.Unlock()

		_ = c.Close()
	}()

	uid := c.Params("uid")
	c.Locals("roomId", roomId)

	roomInfo := roomService.GetRoomInfo(roomId)

	c.WriteJSON(messageAction{Action: "UPDATE_ROOM", Payload: roomInfo})
	if !roomService.IsUserInRoomWithId(uid, roomId) {
		c.WriteJSON(messageAction{Action: "NEED_TO_JOIN"})
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
		var receivedMessage messageAction
		if err := json.Unmarshal(msg, &receivedMessage); err != nil {
			log.Println("json unmarshal:", err)
			break
		}

		switch receivedMessage.Action {
		case "JOIN_ROOM":
			joinRoomPayload, err := transformPayloadToJoinRoom(receivedMessage.Payload)
			if err != nil {
				c.WriteJSON(fiber.Map{"error": "INVALID_PAYLOAD"})
				return
			}
			roomInfo, err := socketService.JoinRoom(uid, joinRoomPayload.Name, joinRoomPayload.Profile, roomId)
			if err != nil {
				log.Printf(err.Error())
				c.WriteJSON(fiber.Map{"error": "JOIN_ROOM_FAILED"})
				return
			}
			noticeUpdateRoom(roomId, roomInfo)

		case "UPDATE_ESTIMATED_VALUE":
			roomInfo = roomService.GetRoomInfo(roomId)
			index := socketService.FindMemberIndex(roomInfo.Members, uid)
			if index != -1 {
				estimatedPayload, err := transformPayloadToEstimatedPoint(receivedMessage.Payload)
				if err != nil {
					c.WriteJSON(fiber.Map{"error": "INVALID_PAYLOAD"})
					return
				}
				roomInfo, err := socketService.UpdateEstimatedValue(index, estimatedPayload.Value, roomId)

				if err != nil {
					log.Printf(err.Error())
					c.WriteJSON(fiber.Map{"error": "UPDATE_ESTIMATED_VALUE_FAILED"})
					return
				}
				noticeUpdateRoom(roomId, roomInfo)

			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}
		case "REVEAL_CARDS":
			roomInfo = roomService.GetRoomInfo(roomId)
			index := socketService.FindMemberIndex(roomInfo.Members, uid)
			if index != -1 {
				roomInfo, err := socketService.RevealCards(index, roomId)
				if err != nil {
					c.WriteJSON(fiber.Map{"error": "REVEAL_CARDS_FAILED"})
					return
				}

				noticeUpdateRoom(roomId, roomInfo)
			} else {
				c.WriteJSON(fiber.Map{"error": "NOT_FOUND_USER"})
			}

		case "RESET_ROOM":
			roomInfo, err := socketService.ResetRoom(roomId)
			if err != nil {
				c.WriteJSON(fiber.Map{"error": "RESET_ROOM_FAILED"})
				return
			}
			noticeUpdateRoom(roomId, roomInfo)

		}
	}
}
