package handlers

import (
	"encoding/json"
	"leaderboard/internal/services"
	"net/http"
	"strconv"
)

type LeaderboardHandler struct {
	leaderboardService *services.LeaderboardService
	simulationService  *services.SimulationService
}

func NewLeaderboardHandler(
	leaderboardService *services.LeaderboardService,
	simulationService *services.SimulationService,
) *LeaderboardHandler {
	return &LeaderboardHandler{
		leaderboardService: leaderboardService,
		simulationService:  simulationService,
	}
}

func (h *LeaderboardHandler) StartSimulation(w http.ResponseWriter, r *http.Request) {
	h.simulationService.Start()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "simulation started"}`))
}

func (h *LeaderboardHandler) StopSimulation(w http.ResponseWriter, r *http.Request) {
	h.simulationService.Stop()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "simulation stopped"}`))
}

func (h *LeaderboardHandler) GetSimulationStatus(w http.ResponseWriter, r *http.Request) {
	status := "stopped"
	if h.simulationService.IsRunning() {
		status = "running"
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

type createUserRequest struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}

func (h *LeaderboardHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.leaderboardService.CreateUser(req.Username, req.Rating)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

type updateRatingRequest struct {
	Rating int `json:"rating"`
}

func (h *LeaderboardHandler) UpdateRating(w http.ResponseWriter, r *http.Request) {

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "Missing idStr", http.StatusBadRequest)
		return
	}

	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	var req updateRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.leaderboardService.UpdateRating(userId, req.Rating); err != nil {
		http.Error(w, "Failed to update rating", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LeaderboardHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50 // Default limit
	}

	offset, _ := strconv.Atoi(offsetStr)
	if offset < 0 {
		offset = 0
	}

	users, err := h.leaderboardService.GetLeaderboard(limit, offset)

	if err != nil {
		http.Error(w, "Failed to get leaderboard", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func (h *LeaderboardHandler) GetUserWithRank(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing username", http.StatusBadRequest)
		return
	}

	users, err := h.leaderboardService.SearchUsers(username)
	if err != nil {
		http.Error(w, "Failed to search users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
