package session

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/raksitnongbua/planning-poker-service/configs"
	"github.com/raksitnongbua/planning-poker-service/constants"
	"golang.org/x/crypto/hkdf"
)

func decryptJWE(tokenEncrypted string) (nextAuthProfile, error) {
	var profile nextAuthProfile
	secret := configs.Conf.AuthSecret
	hash := sha256.New
	kdf := hkdf.New(hash, []byte(secret), []byte(""), []byte(constants.EncryptionInfo))
	key := make([]byte, 32)
	_, _ = io.ReadFull(kdf, key)

	decrypted, err := jwe.Decrypt([]byte(tokenEncrypted),
		jwe.WithKey(jwa.DIRECT, key))

	if err != nil {
		fmt.Printf("failed to decrypt: %s", err)
		return profile, errors.New("failed to decrypt")
	}

	profile, err = unmarshalNextAuthProfile(decrypted)
	if err != nil {
		fmt.Printf("failed to unmarshal: %s", err)
		return profile, errors.New("failed to unmarshal")
	}

	return profile, nil

}

func unmarshalNextAuthProfile(data []byte) (nextAuthProfile, error) {
	var r nextAuthProfile
	err := json.Unmarshal(data, &r)
	return r, err
}
