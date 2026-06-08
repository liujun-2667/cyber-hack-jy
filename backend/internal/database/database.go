package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/lib/pq"
)

var DB *sql.DB

func Init() error {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "cyberhack"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "cyberhack2077"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "cyberhack"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(25)
	DB.SetConnMaxLifetime(5 * time.Minute)

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

func GetPlayerByID(playerID string) (*Player, error) {
	player := &Player{}
	query := `SELECT id, username, elo_rating, wins, losses, created_at, last_login 
	          FROM players WHERE id = $1`
	
	err := DB.QueryRow(query, playerID).Scan(
		&player.ID,
		&player.Username,
		&player.EloRating,
		&player.Wins,
		&player.Losses,
		&player.CreatedAt,
		&player.LastLogin,
	)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func GetPlayerByUsername(username string) (*Player, error) {
	player := &Player{}
	query := `SELECT id, username, elo_rating, wins, losses, created_at, last_login 
	          FROM players WHERE username = $1`
	
	err := DB.QueryRow(query, username).Scan(
		&player.ID,
		&player.Username,
		&player.EloRating,
		&player.Wins,
		&player.Losses,
		&player.CreatedAt,
		&player.LastLogin,
	)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func CreatePlayer(player *Player) error {
	query := `INSERT INTO players (id, username, password_hash, elo_rating, wins, losses) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	
	_, err := DB.Exec(query,
		player.ID,
		player.Username,
		player.PasswordHash,
		player.EloRating,
		player.Wins,
		player.Losses,
	)
	return err
}

func UpdatePlayerElo(playerID string, newElo int) error {
	query := `UPDATE players SET elo_rating = $1, last_login = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := DB.Exec(query, newElo, playerID)
	return err
}

func RecordGameResult(gameID string, roomID string, playerIDs []string, winnerID string, gameMode string, turns int, duration int) error {
	query := `INSERT INTO game_records (id, room_id, player_ids, winner_id, game_mode, turns, duration) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`
	
	_, err := DB.Exec(query, gameID, roomID, pq.Array(playerIDs), winnerID, gameMode, turns, duration)
	return err
}

type Player struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	EloRating    int       `json:"eloRating"`
	Wins         int       `json:"wins"`
	Losses       int       `json:"losses"`
	CreatedAt    time.Time `json:"createdAt"`
	LastLogin    time.Time `json:"lastLogin"`
}

var _ = pq.Array
