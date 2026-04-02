package roomsocket

import (
	"encoding/json"
	"fmt"
)

func transformPayloadToEstimatedPoint(payload interface{}) (data estimatedPointPayload, err error) {
	updatePointPayload, ok := payload.(map[string]interface{})
	if !ok {
		return estimatedPointPayload{}, fmt.Errorf("Invalid payload format for UPDATE_POINT action")
	}

	var updatePointData estimatedPointPayload
	payloadBytes, err := json.Marshal(updatePointPayload)
	if err != nil {
		return estimatedPointPayload{}, fmt.Errorf("Error marshaling payload: %d", err)

	}

	err = json.Unmarshal(payloadBytes, &updatePointData)
	if err != nil {
		return estimatedPointPayload{}, fmt.Errorf("Error unmarshal payload: %d", err)
	}

	return updatePointData, nil
}

func transformPayloadToThrowEmoji(payload interface{}) (data throwEmojiPayload, err error) {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return throwEmojiPayload{}, fmt.Errorf("Invalid payload format for THROW_EMOJI action")
	}

	var throwData throwEmojiPayload
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return throwEmojiPayload{}, fmt.Errorf("Error marshaling payload: %d", err)
	}

	err = json.Unmarshal(payloadBytes, &throwData)
	if err != nil {
		return throwEmojiPayload{}, fmt.Errorf("Error unmarshal payload: %d", err)
	}

	return throwData, nil
}

func transformPayloadToSetTicketEstimation(payload interface{}) (data setTicketEstimationPayload, err error) {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return setTicketEstimationPayload{}, fmt.Errorf("Invalid payload format for SET_TICKET_ESTIMATION action")
	}

	var result setTicketEstimationPayload
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return setTicketEstimationPayload{}, fmt.Errorf("Error marshaling payload: %v", err)
	}

	err = json.Unmarshal(payloadBytes, &result)
	if err != nil {
		return setTicketEstimationPayload{}, fmt.Errorf("Error unmarshal payload: %v", err)
	}

	return result, nil
}

func transformPayloadToSetTicketQueue(payload interface{}) (data setTicketQueuePayload, err error) {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return setTicketQueuePayload{}, fmt.Errorf("Invalid payload format for SET_TICKET_QUEUE action")
	}

	var result setTicketQueuePayload
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return setTicketQueuePayload{}, fmt.Errorf("Error marshaling payload: %v", err)
	}

	err = json.Unmarshal(payloadBytes, &result)
	if err != nil {
		return setTicketQueuePayload{}, fmt.Errorf("Error unmarshal payload: %v", err)
	}

	return result, nil
}

func transformPayloadToSetTicketQueueWithEstimation(payload interface{}) (data setTicketQueueWithEstimationPayload, err error) {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return setTicketQueueWithEstimationPayload{}, fmt.Errorf("Invalid payload format for SET_TICKET_QUEUE_WITH_ESTIMATION action")
	}

	var result setTicketQueueWithEstimationPayload
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return setTicketQueueWithEstimationPayload{}, fmt.Errorf("Error marshaling payload: %v", err)
	}

	err = json.Unmarshal(payloadBytes, &result)
	if err != nil {
		return setTicketQueueWithEstimationPayload{}, fmt.Errorf("Error unmarshal payload: %v", err)
	}

	return result, nil
}

// transformPayloadToNextRound parses the optional NEXT_ROUND payload.
// Returns zero-value (nil ticket, empty queue) when payload is absent.
func transformPayloadToNextRound(payload interface{}) nextRoundPayload {
	if payload == nil {
		return nextRoundPayload{}
	}
	p, ok := payload.(map[string]interface{})
	if !ok {
		return nextRoundPayload{}
	}
	var result nextRoundPayload
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return nextRoundPayload{}
	}
	if err := json.Unmarshal(payloadBytes, &result); err != nil {
		return nextRoundPayload{}
	}
	return result
}

func transformPayloadToJoinRoom(payload interface{}) (data joinRoomPayload, err error) {
	payload, ok := payload.(map[string]interface{})
	if !ok {
		return joinRoomPayload{}, fmt.Errorf("Invalid payload format for JOIN_ROOM action")
	}

	var joinRoomData joinRoomPayload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return joinRoomPayload{}, fmt.Errorf("Error marshaling payload: %d", err)
	}

	err = json.Unmarshal(payloadBytes, &joinRoomData)
	if err != nil {
		return joinRoomPayload{}, fmt.Errorf("Error unmarshal payload: %d", err)
	}

	// Input validation (security: prevent resource exhaustion and XSS)
	if len(joinRoomData.Name) == 0 || len(joinRoomData.Name) > 100 {
		return joinRoomPayload{}, fmt.Errorf("name must be 1-100 characters")
	}
	if len(joinRoomData.Profile) > 500 {
		return joinRoomPayload{}, fmt.Errorf("profile URL too long (max 500)")
	}

	return joinRoomData, nil
}
