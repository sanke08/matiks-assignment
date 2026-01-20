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
	GetLeaderboard(limit, offset int) ([]UserWithRank, error)
	SearchUsersWithRank(query string) ([]UserWithRank, error)
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
func (r *PostgresUserRepository) GetLeaderboard(limit int, offset int) ([]UserWithRank, error) {
	if r.rdb == nil {
		return r.getLeaderboardSQL(limit, offset)
	}

	ctx := context.Background()
	// 1. Fetch Top N members from Redis
	res, err := r.rdb.ZRevRangeWithScores(ctx, LeaderboardKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil || len(res) == 0 {
		return []UserWithRank{}, nil
	}

	// 2. Optimization: Only count unique scores to save Redis resources
	uniqueScores := make(map[float64]*redis.IntCmd)
	pipe := r.rdb.Pipeline()

	for _, z := range res {
		if _, exists := uniqueScores[z.Score]; !exists {
			// Optimization: Request rank only once per unique score
			uniqueScores[z.Score] = pipe.ZCount(ctx, LeaderboardKey, "("+strconv.FormatFloat(z.Score, 'f', -1, 64), "+inf")
		}
	}
	_, _ = pipe.Exec(ctx)

	// 3. Assemble Final Response without any DB or extra Redis hits
	userWithRanks := make([]UserWithRank, 0, len(res))
	for _, z := range res {
		member := z.Member.(string)
		parts := strings.Split(member, ":")
		if len(parts) < 2 {
			continue
		}

		username := parts[0]
		id, _ := strconv.Atoi(parts[1])

		// Get cached rank from our unique scores map
		higherCount, _ := uniqueScores[z.Score].Result()

		userWithRanks = append(userWithRanks, UserWithRank{
			User: models.User{
				ID:       id,
				Username: username,
				Rating:   int(z.Score),
			},
			Rank: int(higherCount) + 1,
		})
	}

	return userWithRanks, nil
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

// SearchUsersWithRank implements UserRepository.
func (r *PostgresUserRepository) SearchUsersWithRank(query string) ([]UserWithRank, error) {
	var users []models.User
	err := r.db.Where("username LIKE ?", "%"+query+"%").
		Order("rating DESC").
		Limit(10).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	results := make([]UserWithRank, 0, len(users))
	ctx := context.Background()

	if r.rdb == nil {
		for _, u := range users {
			rank, _ := r.getUserWithRankSQL(&u)
			results = append(results, UserWithRank{User: u, Rank: rank})
		}
		return results, nil
	}

	pipe := r.rdb.Pipeline()
	rankCmds := make([]*redis.IntCmd, len(users))
	for i, u := range users {
		rankCmds[i] = pipe.ZCount(ctx, LeaderboardKey, "("+strconv.Itoa(u.Rating), "+inf")
	}
	_, _ = pipe.Exec(ctx)

	for i, u := range users {
		count, _ := rankCmds[i].Result()
		results = append(results, UserWithRank{
			User: u,
			Rank: int(count) + 1,
		})
	}

	return results, nil
}

func (r *PostgresUserRepository) getUserWithRankSQL(user *models.User) (int, error) {
	var rank int
	err := r.db.Raw("SELECT rank FROM (SELECT id, RANK() OVER (ORDER BY rating DESC) as rank FROM users) s WHERE id = ?", user.ID).Scan(&rank).Error
	return rank, err
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
