package room

import "encoding/json"

type roomRequest struct {
	RoomName   string `json:"room_name"`
	HostingID  string `json:"hosting_id"`
	DeskConfig string `json:"desk_config"`
}

func unmarshalRoomRequest(data []byte) (roomRequest, error) {
	var r roomRequest
	err := json.Unmarshal(data, &r)
	return r, err
}
