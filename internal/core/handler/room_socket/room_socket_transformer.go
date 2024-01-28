package roomsocket

import (
	"encoding/json"
	"fmt"
)

func TransformPayloadToEstimatedPoint(payload interface{}) (data EstimatedPointPayload, err error) {
	updatePointPayload, ok := payload.(map[string]interface{})
	if !ok {
		return EstimatedPointPayload{}, fmt.Errorf("invalid payload format for UPDATE_POINT action")
	}

	var updatePointData EstimatedPointPayload
	payloadBytes, err := json.Marshal(updatePointPayload)
	if err != nil {
		return EstimatedPointPayload{}, fmt.Errorf("Error marshaling payload:", err)

	}

	err = json.Unmarshal(payloadBytes, &updatePointData)
	if err != nil {
		return EstimatedPointPayload{}, fmt.Errorf("Error unmarshal payload:", err)
	}

	return updatePointData, nil
}
