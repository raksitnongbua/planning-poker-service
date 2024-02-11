package profile

import (
	"github.com/raksitnongbua/planning-poker-service/internal/core/auth/session"
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
)

func GetProfile(token string) (*domain.Profile, error) {
	profile, err := session.GetProfile(token)
	if err != nil {
		return profile, err
	}
	return profile, nil
}
