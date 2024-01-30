package room

type RoomRequest struct {
	RoomName   string `json:"room_name"`
	HostingID  string `json:"hosting_id"`
	DeskConfig string `json:"desk_config"`
}
