package configs

import (
	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type config struct {
	FirebaseCredentials string `env:"FIREBASE_CREDENTIALS,required"`
	AuthSecret          string `env:"NEXTAUTH_SECRET,required"`
}

var Conf config

func Init() {
	// Load dotenv configuration
	if err := godotenv.Load(".env"); err != nil {
		panic(err.Error())
	}
	Conf = config{}
	if err := env.Parse(&Conf); err != nil {
		panic(err.Error())
	}
}
