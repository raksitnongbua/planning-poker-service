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

func SetTicketEstimation(est *domain.TicketEstimation, roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.SetTicketEstimation(est, now)
	err := repo.SetTicketEstimation(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}

func SetTicketQueue(queue []domain.TicketEstimation, roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.SetTicketQueue(queue, now)
	err := repo.SetTicketQueue(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}

func SetTicketQueueWithEstimation(queue []domain.TicketEstimation, est *domain.TicketEstimation, roomId string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.SetTicketQueue(queue, now)
	roomInfo.SetTicketEstimation(est, now)
	err := repo.SetTicketQueue(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}

func SetFinalStoryPoint(roomId string, value string) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)
	roomInfo.ConfirmFinalStoryPoint(value, now)
	err := repo.SetFinalStoryPoint(roomId, roomInfo)
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

func ResetRoomWithTicket(roomId string, ticket domain.TicketEstimation, queue []domain.TicketEstimation) (domain.Room, error) {
	now := timer.GetTimeNow()
	roomInfo := roomService.GetRoomInfo(roomId)

	roomInfo.RestartWithTicket(ticket, queue, now)

	err := repo.ResetRoom(roomId, roomInfo)
	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}
