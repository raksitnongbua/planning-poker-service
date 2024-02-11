package session

import "github.com/raksitnongbua/planning-poker-service/internal/core/domain"

type nextAuthProfile struct {
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
