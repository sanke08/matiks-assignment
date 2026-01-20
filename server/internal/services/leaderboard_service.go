package services

import (
	"errors"
	"leaderboard/internal/models"
	"leaderboard/internal/repository"
)

type LeaderboardService struct {
	userRepo repository.UserRepository
}

func NewLeaderboardService(userRepo repository.UserRepository) *LeaderboardService {
	return &LeaderboardService{userRepo: userRepo}
}

func (s *LeaderboardService) CreateUser(username string, rating int) (*models.User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}

	if rating < 0 {
		return nil, errors.New("rating cannot be negative")
	}

	user := &models.User{
		Username: username,
		Rating:   rating,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *LeaderboardService) UpdateRating(userId, newRating int) error {
	if newRating < 0 {
		return errors.New("rating cannot be negative")
	}

	return s.userRepo.UpdateRating(userId, newRating)
}

func (s *LeaderboardService) GetLeaderboard(limit, offset int) ([]repository.UserWithRank, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}

	if limit > 100 {
		limit = 100
	}

	return s.userRepo.GetLeaderboard(limit, offset)
}

func (s *LeaderboardService) GetUserWithRank(username string) (*models.User, int, error) {
	if username == "" {
		return nil, 0, errors.New("username is required")
	}
	return s.userRepo.GetUserWithRank(username)
}
