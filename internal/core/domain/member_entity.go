package domain

import "time"

type Member struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Picture        string    `json:"picture"`
	LastActiveAt   time.Time `json:"last_active_at"`
	EstimatedValue string    `json:"estimated_value"`
}

func NewMember(id, name, picture string, lastActiveAt time.Time) *Member {
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

func (m *Member) SetLastActiveAt(lastActiveAt time.Time) {
	m.LastActiveAt = lastActiveAt
}
