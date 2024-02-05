package nextauth

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	"golang.org/x/crypto/hkdf"
)

type NextAuthProfile struct {
	Name        string `json:"name"`
	Email       string `json:"email"`
	Picture     string `json:"picture"`
	Sub         string `json:"sub"`
	AccessToken string `json:"accessToken"`
	Iat         int64  `json:"iat"`
	Exp         int64  `json:"exp"`
	Jti         string `json:"jti"`
}

func GetProfile(tokenEncrypted string) (*domain.Profile, error) {
	nextAuthProfile, err := decryptJWE(tokenEncrypted)
	if err != nil {
		return nil, err
	}
	r := domain.NewProfile(nextAuthProfile.Sub, nextAuthProfile.Name, nextAuthProfile.Email, nextAuthProfile.Picture)

	return r, err
}

func decryptJWE(tokenEncrypted string) (NextAuthProfile, error) {
	var profile NextAuthProfile
	nextAuthSecret := os.Getenv("NEXTAUTH_SECRET")
	info := "NextAuth.js Generated Encryption Key"

	hash := sha256.New
	kdf := hkdf.New(hash, []byte(nextAuthSecret), []byte(""), []byte(info))
	key := make([]byte, 32)
	_, _ = io.ReadFull(kdf, key)

	decrypted, err := jwe.Decrypt([]byte(tokenEncrypted),
		jwe.WithKey(jwa.DIRECT, key))

	if err != nil {
		fmt.Printf("failed to decrypt: %s", err)
		return profile, errors.New("failed to decrypt")
	}

	profile, err = UnmarshalNextAuthProfile(decrypted)
	if err != nil {
		fmt.Printf("failed to unmarshal: %s", err)
		return profile, errors.New("failed to unmarshal")
	}

	return profile, nil

}

func UnmarshalNextAuthProfile(data []byte) (NextAuthProfile, error) {
	var r NextAuthProfile
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *NextAuthProfile) Marshal() ([]byte, error) {
	return json.Marshal(r)
}
