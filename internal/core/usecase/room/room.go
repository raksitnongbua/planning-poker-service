package room

import (
	"errors"
	"time"

	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	idgenerator "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/id_generator"
	repo "github.com/raksitnongbua/planning-poker-service/internal/repository/room"
)

func IsUserInRoomWithId(userId, roomId string) bool {
	roomInfo := GetRoomInfo(roomId)
	return roomInfo.CheckMember(userId)
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

func CleanupExpiredRooms() (domain.CleanupResult, error) {
	return repo.DeleteExpiredRooms()
}

func KickMember(roomId, memberID string) (domain.Room, error) {
	roomInfo := GetRoomInfo(roomId)
	if !roomInfo.KickMember(memberID, time.Now()) {
		return domain.Room{}, errors.New("member not found")
	}
	if err := repo.KickMember(roomId, roomInfo); err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}

func CreateNewRoom(roomName, deskConfig string) (string, error) {
	roomId := idgenerator.GenerateUniqueRoomID()
	room := domain.NewRoom(roomName, roomId, deskConfig)

	err := repo.CreateNewRoom(roomId, room)

	if err != nil {
		return "", err
	}

	return roomId, nil
}
