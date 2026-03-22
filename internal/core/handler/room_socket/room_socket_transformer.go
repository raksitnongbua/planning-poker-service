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

func transformPayloadToSetJiraIssue(payload interface{}) (data setJiraIssuePayload, err error) {
	p, ok := payload.(map[string]interface{})
	if !ok {
		return setJiraIssuePayload{}, fmt.Errorf("Invalid payload format for SET_JIRA_ISSUE action")
	}

	var result setJiraIssuePayload
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return setJiraIssuePayload{}, fmt.Errorf("Error marshaling payload: %v", err)
	}

	err = json.Unmarshal(payloadBytes, &result)
	if err != nil {
		return setJiraIssuePayload{}, fmt.Errorf("Error unmarshal payload: %v", err)
	}

	return result, nil
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
