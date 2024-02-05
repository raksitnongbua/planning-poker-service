package domain

type Member struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Picture        string `json:"picture"`
	LastActiveAt   string `json:"last_active_at"`
	EstimatedValue string `json:"estimated_value"`
}

func NewMember(id, name, picture, lastActiveAt string) *Member {
	return &Member{
		ID:             id,
		Name:           name,
		Picture:        picture,
		LastActiveAt:   lastActiveAt,
		EstimatedValue: "",
	}
}

func (m *Member) SetEstimatedValue(value string) {
	m.EstimatedValue = value
}

func (m *Member) SetLastActiveAt(lastActiveAt string) {
	m.LastActiveAt = lastActiveAt
}
