package domain

type Member struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	LastActiveAt   string `json:"last_active_at"`
	EstimatedValue string `json:"estimated_value"`
}
