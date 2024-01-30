package domain

type Room struct {
	Name       string         `json:"name"`
	Members    []Member       `json:"members"`
	Status     string         `json:"status"`
	Result     map[string]int `json:"result"`
	CreatedAt  string         `json:"created_at"`
	UpdatedAt  string         `json:"updated_at"`
	MemberIDs  []string       `json:"member_ids"`
	DeskConfig string         `json:"desk_config"`
}
