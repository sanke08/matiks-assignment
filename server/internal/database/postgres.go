package database

import (
	"leaderboard/internal/config"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(config *config.Config) *gorm.DB {
	db, err := gorm.Open(postgres.Open(config.DatabaseURL), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect to database")
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Fatal("failed to get sqlDB")
	}

	sqlDb.SetMaxIdleConns(25)
	sqlDb.SetMaxOpenConns(125)
	sqlDb.SetConnMaxLifetime(10 * time.Minute)

	return db
}
