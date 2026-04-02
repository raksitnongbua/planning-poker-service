package idgenerator

import (
	"crypto/rand"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GenerateUniqueRoomID() string {
	randomString := generateRandomString(5)
	uuid := uuid.New().String()

	return strings.Join([]string{randomString, uuid}, "-")
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			// Critical: UUID generation must not fail. Panic to prevent insecure fallback.
			panic("crypto/rand failure: " + err.Error())
		}
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func GenerateUUID() string {
	randomString := generateRandomString(5)
	uuid := uuid.New().String()
	timestamp := time.Now().Format("20060102150405")

	return strings.Join([]string{randomString, uuid, timestamp}, "-")
}
