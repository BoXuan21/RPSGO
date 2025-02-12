package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	GamesPlayed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "rps_games_total",
		Help: "The total number of games played",
	})

	GameResults = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rps_game_results_total",
		Help: "The total number of wins/losses/draws",
	}, []string{"result"})

	ResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "rps_response_time_seconds",
		Help:    "Response time of game moves",
		Buckets: prometheus.LinearBuckets(0.1, 0.1, 10),
	})

	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "rps_http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "endpoint"})
)

func InitPrometheus() {
	// Register metrics with prometheus
	// This is a placeholder for any future initialization needs
}
