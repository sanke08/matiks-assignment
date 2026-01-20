package repository

import (
	"errors"
	"leaderboard/internal/models"

	"gorm.io/gorm"
)

type UserWithRank struct {
	models.User
	Rank int `json:"rank"`
}

type UserRepository interface {
	Create(u *models.User) error
	UpdateRating(userID int, newRating int) error
	GetByUsername(username string) (*models.User, error)
	GetLeaderboard(limit, offset int) ([]UserWithRank, error)
	GetUserWithRank(username string) (*models.User, int, error)
}

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// Create implements UserRepository.
func (r *PostgresUserRepository) Create(u *models.User) error {
	return r.db.Create(&u).Error
}

// GetByUsername implements UserRepository.
func (r *PostgresUserRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	err := r.db.
		Where("username = ?", username).
		First(&user).
		Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetLeaderboard implements UserRepository.
func (r *PostgresUserRepository) GetLeaderboard(limit int, offset int) ([]UserWithRank, error) {
	var users []UserWithRank
	query := `
		SELECT *, DENSE_RANK() OVER (ORDER BY rating DESC) as rank
		FROM users
		ORDER BY rating DESC
		LIMIT ? OFFSET ?
	`
	err := r.db.Raw(query, limit, offset).Scan(&users).Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserWithRank implements UserRepository.
func (r *PostgresUserRepository) GetUserWithRank(username string) (*models.User, int, error) {
	var user models.User
	// 1. Get the user's rating
	err := r.db.Where("username LIKE ?", "%"+username+"%").First(&user).Error
	if err != nil {
		return nil, 0, err
	}

	// 2. Count distinct ratings higher than this user's rating
	// Rank = (count of distinct ratings > current_rating) + 1
	var rank int64
	err = r.db.Model(&models.User{}).
		Where("rating > ?", user.Rating).
		Select("COUNT(DISTINCT rating)").
		Scan(&rank).Error

	if err != nil {
		return nil, 0, err
	}

	return &user, int(rank) + 1, nil
}

// UpdateRating implements UserRepository.
func (r *PostgresUserRepository) UpdateRating(userID int, newRating int) error {
	res := r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("rating", newRating)

	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
