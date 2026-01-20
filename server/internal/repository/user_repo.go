package repository

import (
	"context"
	"fmt"
	"leaderboard/internal/models"
	"log"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const LeaderboardKey = "global_leaderboard"

type UserWithRank struct {
	models.User
	Rank int `json:"rank"`
}

type UserRepository interface {
	Create(u *models.User) error
	UpdateRating(userID int, newRating int) error
	GetByUsername(username string) (*models.User, error)
	GetLeaderboard(limit, offset int, query string) ([]UserWithRank, error)
	SyncToRedis() error
}

type PostgresUserRepository struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewPostgresUserRepository(db *gorm.DB, rdb *redis.Client) UserRepository {
	repo := &PostgresUserRepository{db: db, rdb: rdb}
	// Initial sync on startup
	go func() {
		if rdb != nil {
			log.Println("🔄 Initializing Redis leaderboard sync...")
			if err := repo.SyncToRedis(); err != nil {
				log.Printf("❌ Redis sync failed: %v", err)
			} else {
				log.Println("✅ Redis leaderboard sync completed")
			}
		}
	}()
	return repo
}

func (r *PostgresUserRepository) SyncToRedis() error {
	var users []models.User
	if err := r.db.Find(&users).Error; err != nil {
		return err
	}

	ctx := context.Background()
	pipe := r.rdb.Pipeline()

	// Clear existing
	pipe.Del(ctx, LeaderboardKey)

	for _, u := range users {
		// Store as "username:id" to avoid profile lookups
		member := fmt.Sprintf("%s:%d", u.Username, u.ID)
		pipe.ZAdd(ctx, LeaderboardKey, redis.Z{
			Score:  float64(u.Rating),
			Member: member,
		})
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Create implements UserRepository.
func (r *PostgresUserRepository) Create(u *models.User) error {
	if err := r.db.Create(&u).Error; err != nil {
		return err
	}
	if r.rdb != nil {
		ctx := context.Background()
		member := fmt.Sprintf("%s:%d", u.Username, u.ID)
		r.rdb.ZAdd(ctx, LeaderboardKey, redis.Z{
			Score:  float64(u.Rating),
			Member: member,
		})
	}
	return nil
}

// GetByUsername implements UserRepository.
func (r *PostgresUserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.Where("username LIKE ?", "%"+username+"%").First(&user).Error
	return &user, err
}

// GetLeaderboard implements UserRepository.
func (r *PostgresUserRepository) GetLeaderboard(limit int, offset int, query string) ([]UserWithRank, error) {
	var users []models.User
	var err error

	// 1. Fetch matching users depending on if there's a search query
	if query != "" {
		// Search Case: Fetch from DB first (SQL query with LIKE)
		err = r.db.Where("username LIKE ?", "%"+query+"%").
			Order("rating DESC").
			Limit(limit).
			Offset(offset).
			Find(&users).Error
	} else if r.rdb != nil {
		// Global Leaderboard Case: Fetch directly from Redis
		res, err := r.rdb.ZRevRangeWithScores(context.Background(), LeaderboardKey, int64(offset), int64(offset+limit-1)).Result()
		if err != nil || len(res) == 0 {
			return []UserWithRank{}, nil
		}

		// Convert Redis members back to user objects
		for _, z := range res {
			member := z.Member.(string)
			parts := strings.Split(member, ":")
			if len(parts) < 2 {
				continue
			}
			id, _ := strconv.Atoi(parts[1])
			users = append(users, models.User{
				ID:       id,
				Username: parts[0],
				Rating:   int(z.Score),
			})
		}
	} else {
		// Fallback: No Redis, use SQL
		return r.getLeaderboardSQL(limit, offset)
	}

	if err != nil || (len(users) == 0 && query != "") {
		return []UserWithRank{}, err
	}

	// 2. Fetch/Calculate Ranks (Tie-aware: 1, 1, 3)
	results := make([]UserWithRank, 0, len(users))
	ctx := context.Background()

	if r.rdb != nil {
		// Optimization: Batch rank calculation via Pipeline
		pipe := r.rdb.Pipeline()
		uniqueScores := make(map[float64]*redis.IntCmd)

		for _, user := range users {
			score := float64(user.Rating)
			if _, exists := uniqueScores[score]; !exists {
				uniqueScores[score] = pipe.ZCount(ctx, LeaderboardKey, "("+strconv.FormatFloat(score, 'f', -1, 64), "+inf")
			}
		}
		_, _ = pipe.Exec(ctx)

		for _, user := range users {
			higherCount, _ := uniqueScores[float64(user.Rating)].Result()
			results = append(results, UserWithRank{
				User: user,
				Rank: int(higherCount) + 1,
			})
		}
	} else {
		// Fallback to SQL Rank if Redis is down
		for _, user := range users {
			var rank int64
			r.db.Model(&models.User{}).
				Where("rating > ?", user.Rating).
				Select("COUNT(DISTINCT rating)").
				Scan(&rank)
			results = append(results, UserWithRank{
				User: user,
				Rank: int(rank) + 1,
			})
		}
	}

	return results, nil
}

func (r *PostgresUserRepository) getLeaderboardSQL(limit int, offset int) ([]UserWithRank, error) {
	var users []UserWithRank
	query := `
		SELECT *, RANK() OVER (ORDER BY rating DESC) as rank
		FROM users
		ORDER BY rating DESC
		LIMIT ? OFFSET ?
	`
	err := r.db.Raw(query, limit, offset).Scan(&users).Error
	return users, err
}

// UpdateRating implements UserRepository.
func (r *PostgresUserRepository) UpdateRating(userID int, newRating int) error {
	var user models.User
	if err := r.db.First(&user, userID).Error; err != nil {
		return err
	}

	if err := r.db.Model(&user).Update("rating", newRating).Error; err != nil {
		return err
	}

	if r.rdb != nil {
		ctx := context.Background()
		member := fmt.Sprintf("%s:%d", user.Username, user.ID)
		r.rdb.ZAdd(ctx, LeaderboardKey, redis.Z{
			Score:  float64(newRating),
			Member: member,
		})
	}

	return nil
}
