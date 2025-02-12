package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"rps-game/internal/config"
	"rps-game/internal/game"

	_ "github.com/lib/pq"
)

type Database struct {
	db *sql.DB
}

func NewPostgres(cfg config.DBConfig) (*Database, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	var db *sql.DB
	var err error

	// Try to connect with retries
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
		return nil, fmt.Errorf("failed to connect to database after 5 attempts: %v", err)
	}

	// Create tables
	if err := initTables(db); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	log.Println("Successfully connected to database")
	return &Database{db: db}, nil
}

func initTables(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS game_stats (
			id SERIAL PRIMARY KEY,
			result VARCHAR(10),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func (d *Database) SaveGameResult(result string) error {
	_, err := d.db.Exec("INSERT INTO game_stats (result) VALUES ($1)", result)
	return err
}

func (d *Database) GetStats() (game.Stats, error) {
	var stats game.Stats
	rows, err := d.db.Query(`
		SELECT result, COUNT(*) as count 
		FROM game_stats 
		GROUP BY result
	`)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var result string
		var count int
		if err := rows.Scan(&result, &count); err != nil {
			return stats, err
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

	return stats, nil
}

func (d *Database) Close() {
	if d.db != nil {
		d.db.Close()
	}
}
