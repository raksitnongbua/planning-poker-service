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

	return joinRoomData, nil
}
