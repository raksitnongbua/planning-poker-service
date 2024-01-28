package room

import (
	"fmt"

	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
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

func FindMemberIndex(members []domain.Member, targetId string) int {
	for i, user := range members {
		if user.ID == targetId {
			return i
		}
	}

	return -1
}

func calculateResult(members []domain.Member) map[string]int {
	result := make(map[string]int)
	for _, member := range members {
		if member.EstimatedValue != "" {
			result[member.EstimatedValue] = result[member.EstimatedValue] + 1
		}
	}

	return result
}

func JoinRoom(name, id, roomId string) (room domain.Room, err error) {
	if !isUserInRoom(id, room.Members) {
		return domain.Room{}, fmt.Errorf("user is not in the room")
	}

	roomInfo := GetRoomInfo(roomId)

	newMember := domain.Member{
		ID: id, Name: name, LastActiveAt: timer.GetTimeNow(), EstimatedValue: ""}

	roomInfo.Members = append(roomInfo.Members, newMember)
	roomInfo.MemberIDs = append(roomInfo.MemberIDs, id)

	err = repo.UpdateNewJoiner(roomInfo.Members, roomInfo.MemberIDs, roomId)

	if err != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}

func UpdateEstimatedValue(index int, value, roomId string) (room domain.Room, err error) {
	now := timer.GetTimeNow()
	roomInfo := GetRoomInfo(roomId)
	calculatedResult := calculateResult(roomInfo.Members)

	roomInfo.Result = calculatedResult
	roomInfo.Members[index].EstimatedValue = value
	roomInfo.Members[index].LastActiveAt = now
	roomInfo.UpdatedAt = now

	if repo.UpdateEstimatedValue(roomId, roomInfo) != nil {
		return domain.Room{}, err
	}

	return roomInfo, nil
}

func RevealCards(commanderIndex int, roomId string) (room domain.Room, err error) {
	now := timer.GetTimeNow()

	roomInfo := GetRoomInfo(roomId)

	roomInfo.Status = "REVEALED_CARDS"
	roomInfo.UpdatedAt = now
	roomInfo.Members[commanderIndex].LastActiveAt = now

	if repo.SetRevealCards(roomId, roomInfo) != nil {
		return domain.Room{}, err
	}

	return roomInfo, nil
}

func resetEstimatedPointMembers(members []domain.Member) []domain.Member {
	for i := range members {
		members[i].EstimatedValue = ""
	}

	return members
}

func ResetRoom(roomId string) (room domain.Room, err error) {
	now := timer.GetTimeNow()
	roomInfo := GetRoomInfo(roomId)
	roomInfo.UpdatedAt = now

	roomInfo.Status = "VOTING"
	roomInfo.Result = make(map[string]int)
	roomInfo.Members = resetEstimatedPointMembers(roomInfo.Members)

	if repo.ResetRoom(roomId, roomInfo) != nil {
		return domain.Room{}, err
	}
	return roomInfo, nil
}
