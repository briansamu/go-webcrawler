package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBAccess bool
	MongoURI string
	SeedURL  string
}

func Load() *Config {
	dbAccess := true

	if godotenv.Load() != nil {
		fmt.Println("Error loading .env file")
		dbAccess = false
	}

	return &Config{
		DBAccess: dbAccess,
		MongoURI: os.Getenv("MONGO_URI"),
		SeedURL:  os.Getenv("SEED_URL"),
	}
}
