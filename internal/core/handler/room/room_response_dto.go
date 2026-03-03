package room

import "time"

type roomResponse struct {
	RoomID    string    `json:"room_id"`
	CreatedAt time.Time `json:"created_at"`
}
