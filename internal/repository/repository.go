package repository

import (
	"context"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var (
	clientFirestore *firestore.Client
	RoomsColRef     *firestore.CollectionRef
)

func newRoomsCollectionRef() *firestore.CollectionRef {
	return clientFirestore.Collection("rooms")
}

func init() {
	// Load dotenv configuration
	if err := godotenv.Load(".env"); err != nil {
		panic(err.Error())
	}
	firebaseCredentials := os.Getenv("FIREBASE_CREDENTIALS")
	if firebaseCredentials == "" {
		log.Fatal("FIREBASE_CREDENTIALS is not set")
	}
	opt := option.WithCredentialsJSON([]byte(firebaseCredentials))

	client, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	firestore, err := client.Firestore(context.Background())
	if err != nil {
		log.Fatalf("error initializing firestore: %v", err)
	}
	clientFirestore = firestore
	RoomsColRef = newRoomsCollectionRef()
}
