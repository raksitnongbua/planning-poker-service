package roomsocket

type joinRoomPayload struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}
type estimatedPointPayload struct {
	Value string `json:"value"`
}

type setJiraIssuePayload struct {
	JiraIssue *jiraIssueDTO `json:"jiraIssue"`
}

type jiraIssueDTO struct {
	ID               string `json:"id"`
	Key              string `json:"key"`
	Summary          string `json:"summary"`
	Type             string `json:"type"`
	CloudID          string `json:"cloudId"`
	StoryPointsField string `json:"storyPointsField"`
	URL              string `json:"url"`
}

type throwEmojiPayload struct {
	Emoji                string   `json:"emoji"`
	TargetMemberID       *string  `json:"target_member_id,omitempty"`
	TargetTableMemberID  *string  `json:"target_table_member_id,omitempty"`
	TargetPanelMemberID  *string  `json:"target_panel_member_id,omitempty"`
	TargetXRatio         *float64 `json:"target_x_ratio,omitempty"`
	TargetYRatio         *float64 `json:"target_y_ratio,omitempty"`
}

type emojiThrownPayload struct {
	FromUserID           string   `json:"from_user_id"`
	Emoji                string   `json:"emoji"`
	TargetMemberID       *string  `json:"target_member_id,omitempty"`
	TargetTableMemberID  *string  `json:"target_table_member_id,omitempty"`
	TargetPanelMemberID  *string  `json:"target_panel_member_id,omitempty"`
	TargetXRatio         *float64 `json:"target_x_ratio,omitempty"`
	TargetYRatio         *float64 `json:"target_y_ratio,omitempty"`
}
