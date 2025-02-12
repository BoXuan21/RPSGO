package api

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"rps-game/internal/db"
	"rps-game/internal/game"
	"rps-game/internal/metrics"
)

type Handler struct {
	db *db.Database
}

func NewHandler(database *db.Database) *Handler {
	return &Handler{
		db: database,
	}
}

func (h *Handler) PlayGame(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		metrics.ResponseTime.Observe(time.Since(start).Seconds())
	}()

	metrics.HTTPRequestsTotal.With(prometheus.Labels{
		"method":   r.Method,
		"endpoint": "/play",
	}).Inc()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var g game.Game
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !game.ValidateChoice(g.PlayerChoice) {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
		return
	}

	g.ComputerChoice = game.Choices[rand.Intn(len(game.Choices))]
	g.Result = game.DetermineWinner(g.PlayerChoice, g.ComputerChoice)

	metrics.GamesPlayed.Inc()
	metrics.GameResults.With(prometheus.Labels{"result": g.Result}).Inc()

	if err := h.db.SaveGameResult(g.Result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	metrics.HTTPRequestsTotal.With(prometheus.Labels{
		"method":   r.Method,
		"endpoint": "/stats",
	}).Inc()

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.db.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
