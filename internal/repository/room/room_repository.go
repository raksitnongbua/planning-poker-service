package room

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"

	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/common"
	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/timer"
	"github.com/raksitnongbua/planning-poker-service/internal/repository"
)

func QueryRecentRooms(id string) (recentRooms []map[string]interface{}, err error) {
	query := repository.RoomsColRef.Where("MemberIDs", "array-contains", id).OrderBy("UpdatedAt", firestore.Desc)

	docs, err := query.Documents(context.Background()).GetAll()
	if err != nil {
		log.Fatalf("error get recent rooms: %v", err)
		return nil, err
	}

	var rooms []map[string]interface{}
	for _, doc := range docs {
		var room domain.Room
		if err := doc.DataTo(&room); err != nil {
			log.Fatalf("Failed to map Firestore document data: %v", err)
		}
		var newRoom map[string]interface{}
		newRoom = common.StructToMap(room)
		newRoom["id"] = doc.Ref.ID

		rooms = append(rooms, newRoom)
	}
	return rooms, nil
}

func CreateNewRoom(roomId string, room *domain.Room) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Set(context.Background(), room)
	return err
}

func RoomExists(roomId string) bool {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Get(context.Background())
	return err == nil
}

func GetRoomInfo(roomId string) domain.Room {
	docRef := repository.RoomsColRef.Doc(roomId)
	docSnapshot, err := docRef.Get(context.Background())
	if err != nil {
		log.Fatalf("Failed to get document: %v", err)
	}
	var roomInfo domain.Room
	if err := docSnapshot.DataTo(&roomInfo); err != nil {
		log.Fatalf("Failed to map Firestore document data: %v", err)
	}
	return roomInfo
}

func UpdateEstimatedValue(roomId string, roomInfo domain.Room) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{{Path: "Members", Value: roomInfo.Members}, {Path: "Result", Value: roomInfo.Result}, {Path: "UpdatedAt", Value: roomInfo.UpdatedAt}})
	return err
}

func UpdateNewJoiner(members []domain.Member, memberIds []string, roomId string) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{{Path: "Members", Value: members}, {Path: "MemberIDs", Value: memberIds}, {Path: "UpdatedAt", Value: timer.GetTimeNow()}})
	return err
}

func SetRevealCards(roomId string, roomInfo domain.Room) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{{Path: "Members", Value: roomInfo.Members}, {Path: "Status", Value: roomInfo.Status}, {Path: "UpdatedAt", Value: roomInfo.UpdatedAt}})
	return err
}

func ResetRoom(roomId string, roomInfo domain.Room) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "Status", Value: roomInfo.Status},
		{Path: "UpdatedAt", Value: roomInfo.UpdatedAt},
		{Path: "Members", Value: roomInfo.Members},
		{Path: "Result", Value: roomInfo.Result}})
	return err
}
