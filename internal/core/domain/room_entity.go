package domain

import "time"

type Room struct {
	Name                string         `json:"name"`
	Members             []Member       `json:"members"`
	Status              string         `json:"status"`
	Result              map[string]int `json:"result"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	MemberIDs           []string       `json:"member_ids"`
	EverJoinedMemberIDs []string       `json:"ever_joined_member_ids"`
	DeskConfig          string         `json:"desk_config"`
}

func NewRoom(name, roomId, deskConfig string) *Room {
	now := time.Now()
	return &Room{
		Name:                name,
		Members:             []Member{},
		Status:              "VOTING",
		Result:              map[string]int{},
		CreatedAt:           now,
		UpdatedAt:           now,
		MemberIDs:           []string{},
		EverJoinedMemberIDs: []string{},
		DeskConfig:          deskConfig,
	}
}

func (r *Room) JoinRoom(member *Member, updatedAt time.Time) {
	r.UpdatedAt = updatedAt
	r.Members = append(r.Members, *member)
	r.MemberIDs = append(r.MemberIDs, member.ID)

	alreadyRecorded := false
	for _, id := range r.EverJoinedMemberIDs {
		if id == member.ID {
			alreadyRecorded = true
			break
		}
	}
	if !alreadyRecorded {
		r.EverJoinedMemberIDs = append(r.EverJoinedMemberIDs, member.ID)
	}
}

func (r *Room) KickMember(memberID string, updatedAt time.Time) bool {
	newMembers := make([]Member, 0, len(r.Members))
	found := false
	for _, m := range r.Members {
		if m.ID == memberID {
			found = true
			continue
		}
		newMembers = append(newMembers, m)
	}
	if !found {
		return false
	}
	r.Members = newMembers

	newMemberIDs := make([]string, 0, len(r.MemberIDs))
	for _, id := range r.MemberIDs {
		if id != memberID {
			newMemberIDs = append(newMemberIDs, id)
		}
	}
	r.MemberIDs = newMemberIDs
	r.UpdatedAt = updatedAt
	return true
}

func (r *Room) CheckMember(id string) bool {
	for _, member := range r.Members {
		if member.ID == id {
			return true
		}
	}

	return false
}

func (r *Room) UpdateEstimatedValue(index int, value string, updatedAt time.Time) {
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

func (r *Room) RevealCards(actorIndex int, updatedAt time.Time) {
	r.Status = "REVEALED_CARDS"
	r.UpdatedAt = updatedAt
	r.Members[actorIndex].LastActiveAt = updatedAt
}

func (r *Room) TouchMember(index int, updatedAt time.Time) {
	r.Members[index].LastActiveAt = updatedAt
	r.UpdatedAt = updatedAt
}

func (r *Room) Restart(updatedAt time.Time) {
	r.Status = "VOTING"
	r.UpdatedAt = updatedAt
	r.Result = map[string]int{}

	for i := range r.Members {
		r.Members[i].EstimatedValue = ""
	}
}

type DeletedRoom struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CleanupResult struct {
	Message   string        `json:"message"`
	Deleted   int           `json:"deleted"`
	Rooms     []DeletedRoom `json:"rooms"`
	CleanedAt time.Time     `json:"cleaned_at"`
}
