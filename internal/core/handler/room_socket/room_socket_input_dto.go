package roomsocket

type joinRoomPayload struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}
type estimatedPointPayload struct {
	Value string `json:"value"`
}

type throwEmojiPayload struct {
	Emoji               string   `json:"emoji"`
	TargetMemberID      *string  `json:"target_member_id,omitempty"`
	TargetTableMemberID *string  `json:"target_table_member_id,omitempty"`
	TargetXRatio        *float64 `json:"target_x_ratio,omitempty"`
	TargetYRatio        *float64 `json:"target_y_ratio,omitempty"`
}

type emojiThrownPayload struct {
	FromUserID          string   `json:"from_user_id"`
	Emoji               string   `json:"emoji"`
	TargetMemberID      *string  `json:"target_member_id,omitempty"`
	TargetTableMemberID *string  `json:"target_table_member_id,omitempty"`
	TargetXRatio        *float64 `json:"target_x_ratio,omitempty"`
	TargetYRatio        *float64 `json:"target_y_ratio,omitempty"`
}
