package repository

import (
	"errors"
	"leaderboard/internal/models"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(u models.User) error
	UpdateRating(u models.User, newRating int) error
	GetByUsername(username string) (*models.User, error)
	GetLeaderboard(limit, offset int) ([]models.User, error)
	GetUserWithRank(username string) (models.User, int, error)
}

type PostgresUserRepository struct {
	db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) UserRepository {
	return &PostgresUserRepository{db: db}
}

// Create implements UserRepository.
func (r *PostgresUserRepository) Create(u models.User) error {
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
func (r *PostgresUserRepository) GetLeaderboard(limit int, offset int) ([]models.User, error) {
	var users []models.User
	err := r.db.
		Order("rating DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).
		Error

	if err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserWithRank implements UserRepository.
func (r *PostgresUserRepository) GetUserWithRank(username string) (models.User, int, error) {
	type result struct {
		ID       int
		Username string
		Rating   int
		Rank     int
	}

	var res result
	query := `
		SELECT id, username, rating, rank FROM (
			SELECT id, username, rating,
			DENSE_RANK() OVER (ORDER BY rating DESC) as rank
			FROM users
		) ranked
		WHERE username = ?
	`
	err := r.db.Raw(query, username).Scan(&res).Error
	if err != nil {
		return models.User{}, 0, err
	}
	return models.User{
		ID:       res.ID,
		Username: res.Username,
		Rating:   res.Rating,
	}, res.Rank, nil
}

// UpdateRating implements UserRepository.
func (r *PostgresUserRepository) UpdateRating(u models.User, newRating int) error {
	res := r.db.Model(&models.User{}).
		Where("id = ?", u.ID).
		Update("rating", newRating)

	if res.RowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}
