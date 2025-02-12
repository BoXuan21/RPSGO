package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"rps-game/internal/api"
	"rps-game/internal/config"
	"rps-game/internal/db"
	"rps-game/internal/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Load configuration
	cfg := config.Load() // No longer a pointer

	// Initialize database
	database, err := db.NewPostgres(cfg.DB)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer database.Close()

	// Initialize metrics
	metrics.InitPrometheus()

	// Initialize handlers
	handler := api.NewHandler(database)

	// Set up routes
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
	http.HandleFunc("/play", api.EnableCORS(handler.PlayGame))
	http.HandleFunc("/stats", api.EnableCORS(handler.GetStats))
	http.Handle("/metrics", promhttp.Handler())

	// Start server
	log.Printf("Server starting on port %s...", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
