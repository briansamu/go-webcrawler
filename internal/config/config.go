package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBAccess  bool
	MongoURI  string
	SeedURL   string
	UserAgent string
}

func Load() *Config {
	dbAccess := true

	if godotenv.Load() != nil {
		fmt.Println("Error loading .env file")
		dbAccess = false
	}

	userAgent := os.Getenv("USER_AGENT")
	if userAgent == "" {
		fmt.Println("Error: USER_AGENT is not set")
		os.Exit(1)
	}

	return &Config{
		DBAccess:  dbAccess,
		MongoURI:  os.Getenv("MONGO_URI"),
		SeedURL:   os.Getenv("SEED_URL"),
		UserAgent: userAgent,
	}
}
