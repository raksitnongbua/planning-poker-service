package roomsocket

type joinRoomPayload struct {
	Name    string `json:"name"`
	Profile string `json:"profile"`
}
type estimatedPointPayload struct {
	Value string `json:"value"`
}

type setTicketEstimationPayload struct {
	TicketEstimation *ticketEstimationDTO `json:"ticketEstimation"`
}

type ticketEstimationDTO struct {
	Name             string  `json:"name"`
	Source           string  `json:"source"`
	JiraKey          string  `json:"jiraKey"`
	JiraIssueID      string  `json:"jiraIssueId"`
	JiraCloudID      string  `json:"jiraCloudId"`
	JiraURL          string  `json:"jiraUrl"`
	JiraType         string  `json:"jiraType"`
	StoryPointsField string  `json:"storyPointsField"`
	AvgScore         float64 `json:"avgScore,omitempty"`
	FinalScore       string  `json:"finalScore,omitempty"`
}

type setTicketQueuePayload struct {
	TicketQueue []ticketEstimationDTO `json:"ticketQueue"`
}

type setTicketQueueWithEstimationPayload struct {
	TicketQueue      []ticketEstimationDTO `json:"ticketQueue"`
	TicketEstimation *ticketEstimationDTO  `json:"ticketEstimation"`
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
