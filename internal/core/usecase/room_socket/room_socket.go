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

func JoinRoom(id, name, picture, roomId string) (domain.Room, error) {
	roomInfo := roomService.GetRoomInfo(roomId)

	now := timer.GetTimeNow()
	newMember := domain.NewMember(id, name, picture, now)

	roomInfo.JoinRoom(newMember, now)

	err := repo.UpdateNewJoiner(roomId, roomInfo)

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

func TouchMember(uid, roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	index := FindMemberIndex(roomInfo.Members, uid)
	if index == -1 {
		return roomInfo, nil
	}
	roomInfo.TouchMember(index, now)
	return roomInfo, repo.UpdateLastActive(roomId, roomInfo.Members, roomInfo.UpdatedAt)
}

func SetJiraIssue(issue *domain.JiraIssue, roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.SetJiraIssue(issue, now)
	err := repo.SetJiraIssue(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}

func ResetRoom(roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)

	roomInfo.Restart(now)

	err := repo.ResetRoom(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}
