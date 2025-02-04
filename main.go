package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/go-gorm/h2"
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

func initDB() {
	var err error
	db, err = sql.Open("h2", "file:./gamedb;mode=rwc;")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS game_stats (
            id INT PRIMARY KEY AUTO_INCREMENT,
            result VARCHAR(10),
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		log.Fatal(err)
	}
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
	if r.Method != "POST" {
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
	rand.Seed(time.Now().UnixNano())
	game.ComputerChoice = choices[rand.Intn(len(choices))]

	// Determine winner
	game.Result = determineWinner(game.PlayerChoice, game.ComputerChoice)

	// Save result to database
	_, err = db.Exec("INSERT INTO game_stats (result) VALUES (?)", game.Result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

func getStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
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
	initDB()
	defer db.Close()

	http.HandleFunc("/play", playGame)
	http.HandleFunc("/stats", getStats)

	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
