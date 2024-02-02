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

func NewRoom(name, roomId, deskConfig, now string) *Room {
	return &Room{
		Name:       name,
		Members:    []Member{},
		Status:     "VOTING",
		Result:     map[string]int{},
		CreatedAt:  now,
		UpdatedAt:  now,
		MemberIDs:  []string{},
		DeskConfig: deskConfig,
	}
}

func (r *Room) JoinRoom(member *Member, updatedAt string) {
	r.UpdatedAt = updatedAt
	r.Members = append(r.Members, *member)
	r.MemberIDs = append(r.MemberIDs, member.ID)
}

func (r *Room) CheckMember(id string) bool {
	for _, member := range r.Members {
		if member.ID == id {
			return true
		}
	}

	return false
}

func (r *Room) UpdateEstimatedValue(index int, value string, updatedAt string) {
	r.Members[index].EstimatedValue = value
	r.Members[index].LastActiveAt = updatedAt
	r.UpdatedAt = updatedAt
}

func (r *Room) UpdateResult() {
	r.Result = map[string]int{}
	for _, member := range r.Members {
		if member.EstimatedValue != "" {
			r.Result[member.EstimatedValue] = r.Result[member.EstimatedValue] + 1
		}
	}
}

func (r *Room) RevealCards(actorIndex int, updateAt string) {
	r.Status = "REVEALED_CARDS"
	r.UpdatedAt = updateAt
	r.Members[actorIndex].LastActiveAt = updateAt
}

func (r *Room) Restart(updateAt string) {
	r.Status = "VOTING"
	r.UpdatedAt = updateAt
	r.Result = map[string]int{}

	for i := range r.Members {
		r.Members[i].EstimatedValue = ""
	}
}
