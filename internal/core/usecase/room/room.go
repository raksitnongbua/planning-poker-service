package room

import (
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	repo "github.com/raksitnongbua/planning-poker-service/internal/repository/room"
)

func isUserInRoom(userId string, members []domain.Member) bool {
	for _, member := range members {
		if member.ID == userId {
			return true
		}
	}

	return false
}

func IsUserInRoomWithId(userId, roomId string) bool {
	roomInfo := GetRoomInfo(roomId)
	return isUserInRoom(userId, roomInfo.Members)
}

func GetRoomInfo(roomId string) domain.Room {
	return repo.GetRoomInfo(roomId)
}

func IsRoomExists(roomId string) bool {
	return repo.RoomExists(roomId)
}
