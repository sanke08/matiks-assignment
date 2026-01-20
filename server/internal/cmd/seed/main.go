package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"leaderboard/internal/config"
	"leaderboard/internal/database"
	"leaderboard/internal/models"

	"gorm.io/gorm"
)

const (
	totalUsers = 10000
	batchSize  = 500
	minRating  = 100
	maxRating  = 5000
)

func main() {
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// 1. Load config
	cfg := config.Load()

	// 2. Connect to DB
	db := database.New(cfg)

	// 3. Ensure table exists (STRICT, explicit)
	if err := ensureSchema(db); err != nil {
		log.Fatal(err)
	}

	// 4. Seed users
	if err := seedUsers(db); err != nil {
		log.Fatal(err)
	}

	log.Println("✅ Database seeding completed successfully")
}

func ensureSchema(db *gorm.DB) error {
	// Drop table for a clean reset of IDs and data
	db.Exec("DROP TABLE IF EXISTS users CASCADE")

	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		rating INT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_users_rating_desc ON users (rating DESC);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
	`

	return db.Exec(query).Error
}

func seedUsers(db *gorm.DB) error {
	log.Printf("Seeding %d users...\n", totalUsers)

	users := make([]models.User, 0, batchSize)

	for i := 1; i <= totalUsers; i++ {
		user := models.User{
			ID:       i,
			Username: fmt.Sprintf("user_%05d", i),
			Rating:   rand.Intn(maxRating-minRating+1) + minRating,
		}

		users = append(users, user)

		// Insert batch
		if len(users) == batchSize {
			if err := insertBatch(db, users); err != nil {
				return err
			}
			users = users[:0]
		}
	}

	// Insert remaining
	if len(users) > 0 {
		if err := insertBatch(db, users); err != nil {
			return err
		}
	}

	return nil
}

func insertBatch(db *gorm.DB, users []models.User) error {
	return db.CreateInBatches(users, batchSize).Error
}
