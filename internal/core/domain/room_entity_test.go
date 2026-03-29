package domain

import (
	"testing"
	"time"
)

// helpers

func makeRoom() *Room {
	return &Room{
		Name:    "Test Room",
		Members: []Member{},
		Status:  "VOTING",
		Result:  map[string]int{},
	}
}

func makeMember(id, estimatedValue string) Member {
	return Member{
		ID:             id,
		Name:           "Member " + id,
		EstimatedValue: estimatedValue,
	}
}

// ---------------------------------------------------------------------------
// Restart() tests
// ---------------------------------------------------------------------------

func TestRestart_EmptyQueue_TicketEstimationNil(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.Status = "REVEALED_CARDS"
	room.TicketQueue = []TicketEstimation{}
	old := TicketEstimation{Name: "Old"}
	room.TicketEstimation = &old
	room.Members = []Member{
		makeMember("1", "5"),
		makeMember("2", "5"),
	}
	room.FinalStoryPoint = "5"
	room.Result = map[string]int{"5": 2}

	room.Restart(now)

	if room.Status != "VOTING" {
		t.Errorf("expected Status VOTING, got %s", room.Status)
	}
	if room.TicketEstimation != nil {
		t.Errorf("expected TicketEstimation nil, got %+v", room.TicketEstimation)
	}
	if len(room.Result) != 0 {
		t.Errorf("expected empty Result, got %v", room.Result)
	}
	if room.FinalStoryPoint != "" {
		t.Errorf("expected empty FinalStoryPoint, got %s", room.FinalStoryPoint)
	}
	for i, m := range room.Members {
		if m.EstimatedValue != "" {
			t.Errorf("expected member[%d].EstimatedValue empty, got %s", i, m.EstimatedValue)
		}
	}
}

func TestRestart_AllVotedQueue_TicketEstimationNil(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketQueue = []TicketEstimation{
		{Name: "A", AvgScore: 3.5, FinalScore: "3"},
		{Name: "B", AvgScore: 5, FinalScore: "5"},
	}

	room.Restart(now)

	if room.TicketEstimation != nil {
		t.Errorf("expected TicketEstimation nil (all tickets voted), got %+v", room.TicketEstimation)
	}
}

func TestRestart_MixedQueue_SelectsFirstUnvoted(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketQueue = []TicketEstimation{
		{Name: "Voted-1", AvgScore: 3.5, FinalScore: "3"},
		{Name: "Unvoted-1", AvgScore: 0, FinalScore: ""},
		{Name: "Voted-2", AvgScore: 5, FinalScore: "5"},
		{Name: "Unvoted-2", AvgScore: 0, FinalScore: ""},
	}

	room.Restart(now)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be set, got nil")
	}
	if room.TicketEstimation.Name != "Unvoted-1" {
		t.Errorf("expected TicketEstimation.Name == Unvoted-1, got %s", room.TicketEstimation.Name)
	}
}

func TestRestart_SingleUnvotedTicket_SelectsIt(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketQueue = []TicketEstimation{
		{Name: "Only One", AvgScore: 0, FinalScore: ""},
	}

	room.Restart(now)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be set, got nil")
	}
	if room.TicketEstimation.Name != "Only One" {
		t.Errorf("expected TicketEstimation.Name == Only One, got %s", room.TicketEstimation.Name)
	}
}

func TestRestart_ClearsAllMemberEstimates(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.Members = []Member{
		makeMember("1", "5"),
		makeMember("2", "8"),
		makeMember("3", "3"),
	}

	room.Restart(now)

	for i, m := range room.Members {
		if m.EstimatedValue != "" {
			t.Errorf("expected member[%d].EstimatedValue empty, got %s", i, m.EstimatedValue)
		}
	}
}

func TestRestart_ClearsFinalStoryPoint(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.FinalStoryPoint = "13"

	room.Restart(now)

	if room.FinalStoryPoint != "" {
		t.Errorf("expected FinalStoryPoint empty, got %s", room.FinalStoryPoint)
	}
}

func TestRestart_ClearsResultMap(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.Result = map[string]int{"5": 2, "8": 3}

	room.Restart(now)

	if len(room.Result) != 0 {
		t.Errorf("expected empty Result map, got %v", room.Result)
	}
}

func TestRestart_SetsStatusToVoting(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.Status = "REVEALED_CARDS"

	room.Restart(now)

	if room.Status != "VOTING" {
		t.Errorf("expected Status VOTING, got %s", room.Status)
	}
}

func TestRestart_NilTicketEmptyQueue_StaysNil(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketEstimation = nil
	room.TicketQueue = []TicketEstimation{}

	room.Restart(now)

	if room.TicketEstimation != nil {
		t.Errorf("expected TicketEstimation nil, got %+v", room.TicketEstimation)
	}
}

func TestRestart_FirstTimeTicketDuringRevealed(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketEstimation = nil
	room.TicketQueue = []TicketEstimation{
		{Name: "DEMO-1", AvgScore: 0, FinalScore: ""},
		{Name: "DEMO-2", AvgScore: 5, FinalScore: "5"},
	}

	room.Restart(now)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be set, got nil")
	}
	if room.TicketEstimation.Name != "DEMO-1" {
		t.Errorf("expected TicketEstimation.Name == DEMO-1, got %s", room.TicketEstimation.Name)
	}
}

func TestRestart_Idempotent(t *testing.T) {
	t1 := time.Now()
	t2 := t1.Add(5 * time.Second)

	room := makeRoom()
	room.TicketQueue = []TicketEstimation{
		{Name: "DEMO-1", AvgScore: 0, FinalScore: ""},
	}

	room.Restart(t1)
	ticketAfterFirst := room.TicketEstimation

	room.Restart(t2)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be non-nil after second Restart")
	}
	if ticketAfterFirst == nil {
		t.Fatal("expected TicketEstimation to be non-nil after first Restart")
	}
	if room.TicketEstimation.Name != ticketAfterFirst.Name {
		t.Errorf("expected same ticket after second Restart, got %s", room.TicketEstimation.Name)
	}
	if !room.UpdatedAt.Equal(t2) {
		t.Errorf("expected UpdatedAt == t2, got %v", room.UpdatedAt)
	}
}

// ---------------------------------------------------------------------------
// SetTicketQueue() tests
// ---------------------------------------------------------------------------

func TestSetTicketQueue_EmptyQueue_ClearsTicketEstimation(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	current := TicketEstimation{Name: "Current"}
	room.TicketEstimation = &current
	room.TicketQueue = []TicketEstimation{{Name: "A"}, {Name: "B"}}

	room.SetTicketQueue([]TicketEstimation{}, now)

	if room.TicketEstimation != nil {
		t.Errorf("expected TicketEstimation nil after empty queue, got %+v", room.TicketEstimation)
	}
}

func TestSetTicketQueue_NoActivePreviousTicket_SetsFirst(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketEstimation = nil

	room.SetTicketQueue([]TicketEstimation{{Name: "First"}, {Name: "Second"}}, now)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be set, got nil")
	}
	if room.TicketEstimation.Name != "First" {
		t.Errorf("expected TicketEstimation.Name == First, got %s", room.TicketEstimation.Name)
	}
}

func TestSetTicketQueue_ActiveTicketStillPresent_Preserved(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketEstimation = &TicketEstimation{JiraKey: "DEMO-2", Name: "Current"}

	room.SetTicketQueue([]TicketEstimation{
		{JiraKey: "DEMO-1"},
		{JiraKey: "DEMO-2", Name: "Current"},
		{JiraKey: "DEMO-3"},
	}, now)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be preserved, got nil")
	}
	if room.TicketEstimation.JiraKey != "DEMO-2" {
		t.Errorf("expected TicketEstimation.JiraKey == DEMO-2, got %s", room.TicketEstimation.JiraKey)
	}
}

func TestSetTicketQueue_ActiveTicketRemoved_FallbackToFirst(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketEstimation = &TicketEstimation{JiraKey: "DEMO-1", Name: "Removed"}

	room.SetTicketQueue([]TicketEstimation{
		{JiraKey: "DEMO-2"},
		{JiraKey: "DEMO-3"},
	}, now)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be set to first, got nil")
	}
	if room.TicketEstimation.JiraKey != "DEMO-2" {
		t.Errorf("expected TicketEstimation.JiraKey == DEMO-2, got %s", room.TicketEstimation.JiraKey)
	}
}

func TestSetTicketQueue_VotedTicketsStoredCorrectly(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketEstimation = nil

	room.SetTicketQueue([]TicketEstimation{
		{Name: "Voted", AvgScore: 3.5, FinalScore: "3"},
		{Name: "Unvoted"},
	}, now)

	if room.TicketQueue[0].AvgScore != 3.5 {
		t.Errorf("expected TicketQueue[0].AvgScore == 3.5, got %f", room.TicketQueue[0].AvgScore)
	}
	if room.TicketQueue[0].FinalScore != "3" {
		t.Errorf("expected TicketQueue[0].FinalScore == 3, got %s", room.TicketQueue[0].FinalScore)
	}
}

func TestSetTicketQueue_MatchByNameNoJiraKey(t *testing.T) {
	now := time.Now()
	room := makeRoom()
	room.TicketEstimation = &TicketEstimation{JiraKey: "", Name: "My Ticket"}

	room.SetTicketQueue([]TicketEstimation{
		{JiraKey: "", Name: "My Ticket"},
		{JiraKey: "", Name: "Other"},
	}, now)

	if room.TicketEstimation == nil {
		t.Fatal("expected TicketEstimation to be preserved by name, got nil")
	}
	if room.TicketEstimation.Name != "My Ticket" {
		t.Errorf("expected TicketEstimation.Name == My Ticket, got %s", room.TicketEstimation.Name)
	}
}
