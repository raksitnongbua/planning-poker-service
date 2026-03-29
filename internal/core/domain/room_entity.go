package domain

import (
	"math"
	"strconv"
	"strings"
	"time"
)

type TicketEstimation struct {
	Name             string  `json:"name" firestore:"name"`
	Source           string  `json:"source" firestore:"source"`
	JiraKey          string  `json:"jiraKey" firestore:"jiraKey"`
	JiraIssueID      string  `json:"jiraIssueId" firestore:"jiraIssueId"`
	JiraCloudID      string  `json:"jiraCloudId" firestore:"jiraCloudId"`
	JiraURL          string  `json:"jiraUrl" firestore:"jiraUrl"`
	JiraType         string  `json:"jiraType" firestore:"jiraType"`
	StoryPointsField string  `json:"storyPointsField" firestore:"storyPointsField"`
	AvgScore         float64 `json:"avgScore,omitempty" firestore:"avgScore"`
	FinalScore       string  `json:"finalScore,omitempty" firestore:"finalScore"`
}

type Room struct {
	Name                string            `json:"name"`
	Members             []Member          `json:"members"`
	Status              string            `json:"status"`
	Result              map[string]int    `json:"result"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
	MemberIDs           []string          `json:"member_ids"`
	EverJoinedMemberIDs []string          `json:"ever_joined_member_ids"`
	DeskConfig          string            `json:"desk_config"`
	TicketEstimation    *TicketEstimation  `json:"ticket_estimation" firestore:"TicketEstimation"`
	TicketQueue         []TicketEstimation `json:"ticket_queue" firestore:"TicketQueue"`
	FinalStoryPoint     string             `json:"final_story_point" firestore:"FinalStoryPoint"`
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
	r.stampTicketScoresOnReveal()
}

func (r *Room) stampTicketScoresOnReveal() {
	if r.TicketEstimation == nil {
		return
	}
	avg := r.computeAvgFromVotes()
	autoFinal := nearestDeckOption(r.DeskConfig, avg)

	r.TicketEstimation.AvgScore = avg
	if autoFinal != "" {
		r.TicketEstimation.FinalScore = autoFinal
		r.FinalStoryPoint = autoFinal
	}

	estKey := r.TicketEstimation.JiraKey
	if estKey == "" {
		estKey = r.TicketEstimation.Name
	}
	for i, t := range r.TicketQueue {
		key := t.JiraKey
		if key == "" {
			key = t.Name
		}
		if key == estKey {
			r.TicketQueue[i].AvgScore = avg
			if autoFinal != "" {
				r.TicketQueue[i].FinalScore = autoFinal
			}
			break
		}
	}
}

func (r *Room) ConfirmFinalStoryPoint(value string, updatedAt time.Time) {
	r.FinalStoryPoint = value
	r.UpdatedAt = updatedAt
	if r.TicketEstimation == nil {
		return
	}
	avg := r.computeAvgFromVotes()
	r.TicketEstimation.FinalScore = value
	r.TicketEstimation.AvgScore = avg

	estKey := r.TicketEstimation.JiraKey
	if estKey == "" {
		estKey = r.TicketEstimation.Name
	}
	for i, t := range r.TicketQueue {
		key := t.JiraKey
		if key == "" {
			key = t.Name
		}
		if key == estKey {
			r.TicketQueue[i].FinalScore = value
			r.TicketQueue[i].AvgScore = avg
			break
		}
	}
}

func (r *Room) computeAvgFromVotes() float64 {
	var sum float64
	var count int
	for _, m := range r.Members {
		v, err := strconv.ParseFloat(m.EstimatedValue, 64)
		if err == nil {
			sum += v
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return math.Round((sum/float64(count))*10) / 10
}

func nearestDeckOption(deskConfig string, avg float64) string {
	opts := strings.Split(deskConfig, ",")
	nearest := ""
	minDist := math.MaxFloat64
	for _, opt := range opts {
		opt = strings.TrimSpace(opt)
		v, err := strconv.ParseFloat(opt, 64)
		if err != nil {
			continue
		}
		dist := math.Abs(v - avg)
		if dist < minDist {
			minDist = dist
			nearest = opt
		}
	}
	return nearest
}

func (r *Room) TouchMember(index int, updatedAt time.Time) {
	r.Members[index].LastActiveAt = updatedAt
	r.UpdatedAt = updatedAt
}

func (r *Room) Restart(updatedAt time.Time) {
	r.Status = "VOTING"
	r.UpdatedAt = updatedAt
	r.Result = map[string]int{}
	r.FinalStoryPoint = ""

	for i := range r.Members {
		r.Members[i].EstimatedValue = ""
	}

	// Auto-select the first unvoted ticket from the queue
	r.TicketEstimation = nil
	for _, t := range r.TicketQueue {
		if t.AvgScore == 0 && t.FinalScore == "" {
			ticket := t
			r.TicketEstimation = &ticket
			break
		}
	}
}

// RestartWithTicket resets the room and explicitly sets the active ticket,
// optionally replacing the queue. Use when the user re-votes a specific ticket
// rather than taking the auto-selected next one.
func (r *Room) RestartWithTicket(ticket TicketEstimation, queue []TicketEstimation, updatedAt time.Time) {
	if len(queue) > 0 {
		r.TicketQueue = queue
	}
	r.Restart(updatedAt)
	r.TicketEstimation = &ticket
}

func (r *Room) SetTicketEstimation(est *TicketEstimation, updatedAt time.Time) {
	r.TicketEstimation = est
	r.UpdatedAt = updatedAt
}

func (r *Room) SetTicketQueue(queue []TicketEstimation, updatedAt time.Time) {
	r.TicketQueue = queue
	if len(queue) == 0 {
		r.TicketEstimation = nil
	} else if r.TicketEstimation == nil {
		// No active ticket yet — set first in queue
		r.TicketEstimation = &queue[0]
	} else {
		// Keep active ticket if it's still in the new queue; otherwise set first
		activeKey := r.TicketEstimation.JiraKey
		if activeKey == "" {
			activeKey = r.TicketEstimation.Name
		}
		found := false
		for _, t := range queue {
			ticketKey := t.JiraKey
			if ticketKey == "" {
				ticketKey = t.Name
			}
			if ticketKey == activeKey {
				found = true
				break
			}
		}
		if !found {
			r.TicketEstimation = &queue[0]
		}
	}
	r.UpdatedAt = updatedAt
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
