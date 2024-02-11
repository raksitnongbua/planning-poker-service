package roomsocket

type joinRoomPayload struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}
type estimatedPointPayload struct {
	Value string `json:"value"`
}
