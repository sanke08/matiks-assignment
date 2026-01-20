package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseURL string
	SrvPort     int
}

func Load() *Config {

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	return &Config{
		DatabaseURL: dbUrl,
		SrvPort:     8080,
	}
}
