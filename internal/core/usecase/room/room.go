package room

import (
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
)

func IsUserInRoom(userID string, members []domain.Member) bool {
	for _, member := range members {
		if member.ID == userID {
			return true
		}
	}
	return false
}

func FindMemberIndex(members []domain.Member, targetId string) int {
	for i, user := range members {
		if user.ID == targetId {
			return i
		}
	}
	return -1
}

func CalculateResult(members []domain.Member) map[string]int {
	result := make(map[string]int)
	for _, member := range members {
		if member.EstimatedValue != "" {
			result[member.EstimatedValue] = result[member.EstimatedValue] + 1
		}
	}
	return result
}

func UpdateEstimatedValue(members []domain.Member, index int, value string) []domain.Member {
	members[index].EstimatedValue = value
	members[index].LastActiveAt = timer.GetTimeNow()
	return members
}
