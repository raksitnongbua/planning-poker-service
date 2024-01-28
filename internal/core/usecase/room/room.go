package room

import "github.com/raksitnongbua/planning-poker-service/internal/core/domain"

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
