package room

import (
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	idgenerator "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/id_generator"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
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

func GetResendRooms(id string) (rooms []map[string]interface{}, err error) {
	rooms, err = repo.QueryRecentRooms(id)
	return rooms, err
}

func CreateNewRoom(roomName, deskConfig string) (roomId string, err error) {
	now := timer.GetTimeNow()
	roomId = idgenerator.GenerateUniqueRoomID()
	room := domain.Room{Name: roomName, Status: "VOTING", CreatedAt: now, UpdatedAt: now, DeskConfig: deskConfig}

	err = repo.CreateNewRoom(roomId, room)

	if err != nil {
		return "", err
	}

	return roomId, nil
}
