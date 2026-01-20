package database

import (
	"fmt"
	"leaderboard/internal/config"
	"log"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg *config.Config) *gorm.DB {
	// Parse the database URL to get the database name and base URL
	u, err := url.Parse(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("invalid database url")
	}

	dbName := strings.TrimPrefix(u.Path, "/")

	// Create a connection string for the default 'postgres' database
	u.Path = "/postgres"
	postgresURL := u.String()

	// 1. Connect to 'postgres' database to ensure the target database exists
	tempDb, err := gorm.Open(postgres.Open(postgresURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to default postgres database: %v", err)
	}

	if err := createDatabaseIfNotExists(tempDb, dbName); err != nil {
		log.Fatalf("failed to create database %s: %v", dbName, err)
	}

	// Close temporary connection
	sqlTempDb, _ := tempDb.DB()
	sqlTempDb.Close()

	// 2. Connect to the actual target database
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to target database %s: %v", dbName, err)
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

func createDatabaseIfNotExists(db *gorm.DB, dbName string) error {
	var exists bool
	// Check if database exists
	err := db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = ?)", dbName).Scan(&exists).Error
	if err != nil {
		return err
	}

	if !exists {
		log.Printf("Creating database %s...", dbName)
		// Only create if not exists
		// Note: CREATE DATABASE cannot be executed within a transaction block
		return db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName)).Error
	}

	return nil
}
