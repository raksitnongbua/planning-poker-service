package room

import (
	"encoding/json"
	"errors"
)

type roomRequest struct {
	RoomName   string `json:"room_name"`
	HostingID  string `json:"hosting_id"`
	DeskConfig string `json:"desk_config"`
}

func unmarshalRoomRequest(data []byte) (roomRequest, error) {
	var r roomRequest
	err := json.Unmarshal(data, &r)
	if err != nil {
		return r, err
	}

	// Input validation (security: prevent resource exhaustion and DB bloat)
	if err := r.Validate(); err != nil {
		return r, err
	}

	return r, nil
}

// Validate performs security input validation on room request fields
func (r *roomRequest) Validate() error {
	if len(r.RoomName) > 100 {
		return errors.New("room_name exceeds 100 characters")
	}
	if len(r.DeskConfig) > 500 {
		return errors.New("desk_config exceeds 500 characters")
	}
	if len(r.HostingID) > 150 {
		return errors.New("hosting_id exceeds 150 characters")
	}
	return nil
}
