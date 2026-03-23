package room

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"

	"github.com/raksitnongbua/planning-poker-service/internal/core/usecase/common"
	"github.com/raksitnongbua/planning-poker-service/internal/repository"
	"github.com/raksitnongbua/planning-poker-service/pkg/logger"
)

const roomRetention = 30 * 24 * time.Hour

func QueryRecentRooms(id string) (recentRooms []map[string]interface{}, err error) {
	query := repository.RoomsColRef.Where("EverJoinedMemberIDs", "array-contains", id).OrderBy("UpdatedAt", firestore.Desc)

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
	logger.Info("firestore create room", "roomId", roomId)
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
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "Members", Value: roomInfo.Members},
		{Path: "Result", Value: roomInfo.Result},
		{Path: "UpdatedAt", Value: roomInfo.UpdatedAt},
	})
	return err
}

func UpdateNewJoiner(roomId string, roomInfo domain.Room) error {
	logger.Info("firestore update new joiner", "roomId", roomId)
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "Members", Value: roomInfo.Members},
		{Path: "MemberIDs", Value: roomInfo.MemberIDs},
		{Path: "EverJoinedMemberIDs", Value: roomInfo.EverJoinedMemberIDs},
		{Path: "UpdatedAt", Value: time.Now()},
	})
	return err
}

func KickMember(roomId string, roomInfo domain.Room) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "Members", Value: roomInfo.Members},
		{Path: "MemberIDs", Value: roomInfo.MemberIDs},
		{Path: "UpdatedAt", Value: roomInfo.UpdatedAt},
	})
	return err
}

func SetRevealCards(roomId string, roomInfo domain.Room) error {
	logger.Info("firestore reveal cards", "roomId", roomId)
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "Members", Value: roomInfo.Members},
		{Path: "Status", Value: roomInfo.Status},
		{Path: "UpdatedAt", Value: roomInfo.UpdatedAt},
	})
	return err
}

func ResetRoom(roomId string, roomInfo domain.Room) error {
	logger.Info("firestore reset room", "roomId", roomId)
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "Status", Value: roomInfo.Status},
		{Path: "UpdatedAt", Value: roomInfo.UpdatedAt},
		{Path: "Members", Value: roomInfo.Members},
		{Path: "Result", Value: roomInfo.Result},
		{Path: "TicketEstimation", Value: firestore.Delete},
		{Path: "FinalStoryPoint", Value: ""},
	})
	return err
}

func SetFinalStoryPoint(roomId string, value string, updatedAt time.Time) error {
	logger.Info("firestore set final story point", "roomId", roomId, "value", value)
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "FinalStoryPoint", Value: value},
		{Path: "UpdatedAt", Value: updatedAt},
	})
	return err
}

func UpdateLastActive(roomId string, members []domain.Member, updatedAt time.Time) error {
	docRef := repository.RoomsColRef.Doc(roomId)
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "Members", Value: members},
		{Path: "UpdatedAt", Value: updatedAt},
	})
	return err
}

func SetTicketEstimation(roomId string, roomInfo domain.Room) error {
	logger.Info("firestore set ticket estimation", "roomId", roomId)
	docRef := repository.RoomsColRef.Doc(roomId)
	var ticketValue interface{}
	if roomInfo.TicketEstimation != nil {
		ticketValue = roomInfo.TicketEstimation
	} else {
		ticketValue = firestore.Delete
	}
	_, err := docRef.Update(context.Background(), []firestore.Update{
		{Path: "TicketEstimation", Value: ticketValue},
		{Path: "UpdatedAt", Value: roomInfo.UpdatedAt},
	})
	return err
}

func DeleteExpiredRooms() (domain.CleanupResult, error) {
	ctx := context.Background()
	threshold := time.Now().Add(-roomRetention)

	docs, err := repository.RoomsColRef.Where("UpdatedAt", "<", threshold).Documents(ctx).GetAll()
	if err != nil {
		return domain.CleanupResult{}, err
	}

	var deletedRooms []domain.DeletedRoom
	for _, doc := range docs {
		var room domain.Room
		if err := doc.DataTo(&room); err != nil {
			return domain.CleanupResult{}, err
		}
		if _, err := doc.Ref.Delete(ctx); err != nil {
			return domain.CleanupResult{}, err
		}
		deletedRooms = append(deletedRooms, domain.DeletedRoom{
			ID:        doc.Ref.ID,
			Name:      room.Name,
			UpdatedAt: room.UpdatedAt,
		})
	}

	count := len(deletedRooms)
	message := fmt.Sprintf("Successfully deleted %d expired room(s) inactive for more than 30 days", count)
	if count == 0 {
		message = "No expired rooms found"
	}

	return domain.CleanupResult{
		Message:   message,
		Deleted:   count,
		Rooms:     deletedRooms,
		CleanedAt: time.Now(),
	}, nil
}
