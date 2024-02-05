package domain

type Profile struct {
	UID     string `json:"uid"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Picture string `json:"picture"`
}

func NewProfile(uid, name, email, picture string) *Profile {
	return &Profile{
		UID:     uid,
		Name:    name,
		Email:   email,
		Picture: picture,
	}
}
