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

	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /users", leaderboardHandler.CreateUser)
	mux.HandleFunc("PUT /users/rating", leaderboardHandler.UpdateRating)
	mux.HandleFunc("GET /leaderboard", leaderboardHandler.GetLeaderboard)
	mux.HandleFunc("GET /users/rank", leaderboardHandler.GetUserWithRank)

	log.Println("Server started at :" + strconv.Itoa(cfg.SrvPort))

	if err := http.ListenAndServe(":"+strconv.Itoa(cfg.SrvPort), mux); err != nil {
		log.Fatal(err)
	}

}
