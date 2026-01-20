package services

import (
	"context"
	"leaderboard/internal/repository"
	"log"
	"math/rand"
	"sync"
	"time"
)

type SimulationService struct {
	userRepo repository.UserRepository
	cancel   context.CancelFunc
	running  bool
	mu       sync.Mutex
}

func NewSimulationService(userRepo repository.UserRepository) *SimulationService {
	return &SimulationService{userRepo: userRepo}
}

func (s *SimulationService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.running = true

	go s.run(ctx)
	log.Println("🚀 Simulation started: Randomly updating user ratings...")
}

func (s *SimulationService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}
	s.cancel()
	s.running = false
	log.Println("🛑 Simulation stopped")
}

func (s *SimulationService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	log.Printf("Checking simulation status: %v", s.running)
	return s.running
}

func (s *SimulationService) run(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond) // Update every 500ms
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Update 10 random users every tick for more visible changes
			for j := 0; j < 10; j++ {
				randomID := rand.Intn(10000) + 1
				newRating := rand.Intn(4901) + 100 // 100 to 5000

				err := s.userRepo.UpdateRating(randomID, newRating)
				if err != nil {
					log.Printf("Simulation error updating user %d: %v", randomID, err)
				}
			}
		}
	}
}
