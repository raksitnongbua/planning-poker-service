package room

import (
	"context"
	"log"

	"cloud.google.com/go/firestore"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"

	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/common"
	"github.com/raksitnongbua/planning-poker-service/internal/repository"
)

func QueryRecentRooms(ctx context.Context, id string) (recentRooms []map[string]interface{}, err error) {
	query := repository.RoomsColRef.Where("MemberIDs", "array-contains", id).OrderBy("UpdatedAt", firestore.Desc)

	docs, err := query.Documents(ctx).GetAll()
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

func CreateNewRoom(ctx context.Context, roomId string, room domain.Room) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Set(context.TODO(), room)
	return err
}

func RoomExists(roomId string) bool {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Get(context.TODO())
	return err == nil
}

func GetRoomInfo(roomId string) domain.Room {
	docRef := repository.RoomsColRef.Doc(roomId)
	docSnapshot, err := docRef.Get(context.TODO())
	if err != nil {
		log.Fatalf("Failed to get document: %v", err)
	}
	var roomInfo domain.Room
	if err := docSnapshot.DataTo(&roomInfo); err != nil {
		log.Fatalf("Failed to map Firestore document data: %v", err)
	}
	return roomInfo
}
