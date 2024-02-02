package roomsocket

import (
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	roomService "github.com/raksitnongbua/planning-poker-service/internal/core/usecase/room"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"

	repo "github.com/raksitnongbua/planning-poker-service/internal/repository/room"
)

func FindMemberIndex(members []domain.Member, targetId string) int {
	for i, user := range members {
		if user.ID == targetId {
			return i
		}
	}

	return -1
}

func JoinRoom(name, id, roomId string) (domain.Room, error) {
	roomInfo := roomService.GetRoomInfo(roomId)

	newMember := domain.NewMember(id, name, timer.GetTimeNow())

	roomInfo.JoinRoom(newMember, timer.GetTimeNow())

	err := repo.UpdateNewJoiner(roomInfo.Members, roomInfo.MemberIDs, roomId)

	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}

func UpdateEstimatedValue(index int, value, roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.UpdateEstimatedValue(index, value, now)

	// After update estimated value we need to recalculate result and update it.
	roomInfo.UpdateResult()

	err := repo.UpdateEstimatedValue(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}

	return roomInfo, nil
}

func RevealCards(actorIndex int, roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()

	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.RevealCards(actorIndex, now)

	err := repo.SetRevealCards(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}

	return roomInfo, nil
}

func ResetRoom(roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.UpdatedAt = now

	roomInfo.Restart(now)

	err := repo.ResetRoom(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}
