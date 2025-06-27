package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoURI string
	SeedURL  string
	MaxPages int
	DBAccess bool
	Timeout  time.Duration
}

func Load() *Config {
	if godotenv.Load() != nil {
		fmt.Println("Error loading .env file")
	}

	maxPages, _ := strconv.Atoi(os.Getenv("MAX_PAGES"))
	timeout, _ := strconv.Atoi(os.Getenv("TIMEOUT"))

	return &Config{
		MongoURI: os.Getenv("MONGO_URI"),
		SeedURL:  os.Getenv("SEED_URL"),
		MaxPages: maxPages,
		DBAccess: os.Getenv("DB_ACCESS") == "true",
		Timeout:  time.Duration(timeout) * time.Second,
	}
}
