package profile

import (
	"github.com/raksitnongbua/planning-poker-service/internal/core/domain"
	"github.com/raksitnongbua/planning-poker-service/internal/repository/nextauth"
)

func GetProfile(token string) (*domain.Profile, error) {
	profile, err := nextauth.GetProfile(token)
	if err != nil {
		return profile, err
	}
	return profile, nil
}
