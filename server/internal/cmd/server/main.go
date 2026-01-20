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
	rdb := database.NewRedis(cfg)

	userRepo := repository.NewPostgresUserRepository(db, rdb)

	leaderboardService := services.NewLeaderboardService(userRepo)

	simulationService := services.NewSimulationService(userRepo)
	simulationService.Start() // Start automatically on boot

	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService, simulationService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", leaderboardHandler.CreateUser)
	mux.HandleFunc("PUT /users/rating", leaderboardHandler.UpdateRating)
	mux.HandleFunc("GET /leaderboard", leaderboardHandler.GetLeaderboard)

	// Simulation routes
	// mux.HandleFunc("POST /simulation/start", leaderboardHandler.StartSimulation)
	// mux.HandleFunc("POST /simulation/stop", leaderboardHandler.StopSimulation)
	// mux.HandleFunc("GET /simulation/status", leaderboardHandler.GetSimulationStatus)

	log.Println("Server started at :" + strconv.Itoa(cfg.SrvPort))

	handler := enableCORS(mux)

	if err := http.ListenAndServe(":"+strconv.Itoa(cfg.SrvPort), handler); err != nil {
		log.Fatal(err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
