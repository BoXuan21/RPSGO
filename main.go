package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Game struct {
	PlayerChoice   string `json:"player_choice"`
	ComputerChoice string `json:"computer_choice"`
	Result         string `json:"result"`
}

type Stats struct {
	Wins   int `json:"wins"`
	Losses int `json:"losses"`
	Draws  int `json:"draws"`
}

var db *sql.DB
var choices = []string{"rock", "paper", "scissors"}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func initDB() {
	// Get database connection details from environment variables
	host := getEnv("DB_HOST", "localhost")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "postgres")
	dbname := getEnv("DB_NAME", "rps_game")
	port := getEnv("DB_PORT", "5432")

	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	var err error
	// Try to connect to the database with retries
	for i := 0; i < 5; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Failed to connect to database, attempt %d/5. Retrying in 5 seconds...", i+1)
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to database after 5 attempts:", err)
	}

	// Create the table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS game_stats (
            id SERIAL PRIMARY KEY,
            result VARCHAR(10),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to database")
}

func determineWinner(player, computer string) string {
	if player == computer {
		return "draw"
	}

	if (player == "rock" && computer == "scissors") ||
		(player == "paper" && computer == "rock") ||
		(player == "scissors" && computer == "paper") {
		return "win"
	}

	return "lose"
}

func playGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var game Game
	err := json.NewDecoder(r.Body).Decode(&game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate player choice
	valid := false
	for _, choice := range choices {
		if game.PlayerChoice == choice {
			valid = true
			break
		}
	}
	if !valid {
		http.Error(w, "Invalid choice", http.StatusBadRequest)
		return
	}

	// Generate computer choice
	game.ComputerChoice = choices[rand.Intn(len(choices))]

	// Determine winner
	game.Result = determineWinner(game.PlayerChoice, game.ComputerChoice)

	// Save result to database
	_, err = db.Exec("INSERT INTO game_stats (result) VALUES ($1)", game.Result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func getStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stats Stats
	rows, err := db.Query(`
        SELECT result, COUNT(*) as count 
        FROM game_stats 
        GROUP BY result
    `)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var result string
		var count int
		err := rows.Scan(&result, &count)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch result {
		case "win":
			stats.Wins = count
		case "lose":
			stats.Losses = count
		case "draw":
			stats.Draws = count
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Initialize database connection
	initDB()
	defer db.Close()

	// Set up HTTP routes
	http.HandleFunc("/play", playGame)
	http.HandleFunc("/stats", getStats)

	// Start the server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
