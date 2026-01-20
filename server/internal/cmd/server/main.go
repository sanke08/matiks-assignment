package main

import (
	"leaderboard/internal/config"
	"leaderboard/internal/database"
	"leaderboard/internal/handlers"
	"leaderboard/internal/repository"
	"leaderboard/internal/services"
	"log"
	"net/http"
	"strconv"
)

func main() {
	cfg := config.Load()

	db := database.New(cfg)

	userRepo := repository.NewPostgresUserRepository(db)

	leaderboardService := services.NewLeaderboardService(userRepo)

	simulationService := services.NewSimulationService(userRepo)

	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService, simulationService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", leaderboardHandler.CreateUser)
	mux.HandleFunc("PUT /users/rating", leaderboardHandler.UpdateRating)
	mux.HandleFunc("GET /leaderboard", leaderboardHandler.GetLeaderboard)
	mux.HandleFunc("GET /users/rank", leaderboardHandler.GetUserWithRank)

	// Simulation routes
	mux.HandleFunc("POST /simulation/start", leaderboardHandler.StartSimulation)
	mux.HandleFunc("POST /simulation/stop", leaderboardHandler.StopSimulation)
	mux.HandleFunc("GET /simulation/status", leaderboardHandler.GetSimulationStatus)

	log.Println("Server started at :" + strconv.Itoa(cfg.SrvPort))

	if err := http.ListenAndServe(":"+strconv.Itoa(cfg.SrvPort), mux); err != nil {
		log.Fatal(err)
	}

}
