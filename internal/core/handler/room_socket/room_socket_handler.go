package roomsocket

import (
	"encoding/json"
	"sync"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	roomService "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room"
	socketService "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room_socket"
	"github.com/raksitnongbua/planning-poker-service/pkg/logger"
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
			logger.Error("error sending message to client", "error", err)
		}
	}
}

func broadcastToOthers(sender *websocket.Conn, roomId string, message interface{}) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		if client == sender {
			continue
		}
		if client.Locals("roomId") != roomId {
			continue
		}
		if err := client.WriteJSON(message); err != nil {
			logger.Error("error sending message to client", "error", err)
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
		logger.Error("room not found", "roomId", roomId)
		c.Close()
		return
	}

	clientsMu.Lock()
	clients[c] = true
	clientsMu.Unlock()

	uid := c.Params("uid")
	logger.Info("ws client connected", "roomId", roomId, "uid", uid)

	defer func() {
		clientsMu.Lock()
		delete(clients, c)
		clientsMu.Unlock()

		logger.Info("ws client disconnected", "roomId", roomId, "uid", uid)
		_ = c.Close()
	}()
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
			logger.Error("ws read error", "roomId", roomId, "uid", uid, "error", err)
			break
		}
		var receivedMessage messageAction
		if err := json.Unmarshal(msg, &receivedMessage); err != nil {
			logger.Error("ws unmarshal error", "roomId", roomId, "uid", uid, "error", err)
			break
		}

		logger.Info("ws action received", "action", receivedMessage.Action, "roomId", roomId, "uid", uid)

		switch receivedMessage.Action {
		case "JOIN_ROOM":
			joinRoomPayload, err := transformPayloadToJoinRoom(receivedMessage.Payload)
			if err != nil {
				c.WriteJSON(fiber.Map{"error": "INVALID_PAYLOAD"})
				return
			}
			roomInfo, err := socketService.JoinRoom(uid, joinRoomPayload.Name, joinRoomPayload.Profile, roomId)
			if err != nil {
				logger.Error("JOIN_ROOM failed", "roomId", roomId, "uid", uid, "error", err)
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
					logger.Error("UPDATE_ESTIMATED_VALUE failed", "roomId", roomId, "uid", uid, "error", err)
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
					logger.Error("REVEAL_CARDS failed", "roomId", roomId, "uid", uid, "error", err)
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
				logger.Error("RESET_ROOM failed", "roomId", roomId, "error", err)
				c.WriteJSON(fiber.Map{"error": "RESET_ROOM_FAILED"})
				return
			}
			noticeUpdateRoom(roomId, roomInfo)

		case "PING":
			roomInfo, err := socketService.TouchMember(uid, roomId)
			if err != nil {
				logger.Error("PING update failed", "roomId", roomId, "uid", uid, "error", err)
				continue
			}
			noticeUpdateRoom(roomId, roomInfo)

		case "SET_TICKET_ESTIMATION":
			ticketPayload, err := transformPayloadToSetTicketEstimation(receivedMessage.Payload)
			if err != nil {
				logger.Error("SET_TICKET_ESTIMATION invalid payload", "roomId", roomId, "uid", uid, "error", err)
				c.WriteJSON(fiber.Map{"error": "INVALID_PAYLOAD"})
				continue
			}
			var est *domain.TicketEstimation
			if ticketPayload.TicketEstimation != nil {
				est = &domain.TicketEstimation{
					Name:             ticketPayload.TicketEstimation.Name,
					Source:           ticketPayload.TicketEstimation.Source,
					JiraKey:          ticketPayload.TicketEstimation.JiraKey,
					JiraIssueID:      ticketPayload.TicketEstimation.JiraIssueID,
					JiraCloudID:      ticketPayload.TicketEstimation.JiraCloudID,
					JiraURL:          ticketPayload.TicketEstimation.JiraURL,
					JiraType:         ticketPayload.TicketEstimation.JiraType,
					StoryPointsField: ticketPayload.TicketEstimation.StoryPointsField,
				}
			}
			roomInfo, err := socketService.SetTicketEstimation(est, roomId)
			if err != nil {
				logger.Error("SET_TICKET_ESTIMATION failed", "roomId", roomId, "uid", uid, "error", err)
				c.WriteJSON(fiber.Map{"error": "SET_TICKET_ESTIMATION_FAILED"})
				continue
			}
			noticeUpdateRoom(roomId, roomInfo)

		case "THROW_EMOJI":
			throwPayload, err := transformPayloadToThrowEmoji(receivedMessage.Payload)
			if err != nil {
				logger.Error("THROW_EMOJI invalid payload", "roomId", roomId, "uid", uid, "error", err)
				continue
			}
			broadcastToOthers(c, roomId, messageAction{
				Action: "EMOJI_THROWN",
				Payload: emojiThrownPayload{
					FromUserID:          uid,
					Emoji:               throwPayload.Emoji,
					TargetMemberID:      throwPayload.TargetMemberID,
					TargetTableMemberID: throwPayload.TargetTableMemberID,
					TargetPanelMemberID: throwPayload.TargetPanelMemberID,
					TargetXRatio:        throwPayload.TargetXRatio,
					TargetYRatio:        throwPayload.TargetYRatio,
				},
			})
			roomInfo, err := socketService.TouchMember(uid, roomId)
			if err != nil {
				logger.Error("THROW_EMOJI touch member failed", "roomId", roomId, "uid", uid, "error", err)
				continue
			}
			noticeUpdateRoom(roomId, roomInfo)
		}
	}
}
