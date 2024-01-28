package roomsocket

import (
	"encoding/json"
	"fmt"
)

func TransformPayloadToEstimatedPoint(payload interface{}) (data EstimatedPointPayload, err error) {
	updatePointPayload, ok := payload.(map[string]interface{})
	if !ok {
		return EstimatedPointPayload{}, fmt.Errorf("Invalid payload format for UPDATE_POINT action")
	}

	var updatePointData EstimatedPointPayload
	payloadBytes, err := json.Marshal(updatePointPayload)
	if err != nil {
		return EstimatedPointPayload{}, fmt.Errorf("Error marshaling payload: %d", err)

	}

	err = json.Unmarshal(payloadBytes, &updatePointData)
	if err != nil {
		return EstimatedPointPayload{}, fmt.Errorf("Error unmarshal payload: %d", err)
	}

	return updatePointData, nil
}

func TransformPayloadToJoinRoom(payload interface{}) (data JoinRoomPayload, err error) {
	joinRoomPayload, ok := payload.(map[string]interface{})
	if !ok {
		return JoinRoomPayload{}, fmt.Errorf("Invalid payload format for JOIN_ROOM action")
	}

	var joinRoomData JoinRoomPayload
	payloadBytes, err := json.Marshal(joinRoomPayload)
	if err != nil {
		return JoinRoomPayload{}, fmt.Errorf("Error marshaling payload: %d", err)
	}

	err = json.Unmarshal(payloadBytes, &joinRoomData)
	if err != nil {
		return JoinRoomPayload{}, fmt.Errorf("Error unmarshal payload: %d", err)
	}

	return joinRoomData, nil
}
