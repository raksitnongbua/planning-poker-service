package core

type Member struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LastActiveAt   string `json:"last_active_at"`
	EstimatedValue string `json:"estimated_value"`
}

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
