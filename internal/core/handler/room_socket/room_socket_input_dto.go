package roomsocket

type JoinRoomPayload struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}
type EstimatedPointPayload struct {
	Value string `json:"value"`
}
