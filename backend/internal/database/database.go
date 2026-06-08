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

type Player struct {
	ID                string    `json:"id"`
	Username          string    `json:"username"`
	PasswordHash      string    `json:"-"`
	EloRating         int       `json:"eloRating"`
	Wins              int       `json:"wins"`
	Losses            int       `json:"losses"`
	CurrentStreak     int       `json:"currentStreak"`
	BestStreak        int       `json:"bestStreak"`
	RankProtectionGames int    `json:"rankProtectionGames"`
	CurrentRank       string    `json:"currentRank"`
	BestRank          string    `json:"bestRank"`
	TotalNodesDestroyed int     `json:"totalNodesDestroyed"`
	TotalTurnsSurvived int      `json:"totalTurnsSurvived"`
	CreatedAt         time.Time `json:"createdAt"`
	LastLogin         time.Time `json:"lastLogin"`
}

func GetPlayerByID(playerID string) (*Player, error) {
	player := &Player{}
	query := `SELECT id, username, elo_rating, wins, losses, current_streak, best_streak,
	          rank_protection_games, current_rank, best_rank, total_nodes_destroyed,
	          total_turns_survived, created_at, last_login 
	          FROM players WHERE id = $1`
	
	err := DB.QueryRow(query, playerID).Scan(
		&player.ID,
		&player.Username,
		&player.EloRating,
		&player.Wins,
		&player.Losses,
		&player.CurrentStreak,
		&player.BestStreak,
		&player.RankProtectionGames,
		&player.CurrentRank,
		&player.BestRank,
		&player.TotalNodesDestroyed,
		&player.TotalTurnsSurvived,
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
	query := `SELECT id, username, elo_rating, wins, losses, current_streak, best_streak,
	          rank_protection_games, current_rank, best_rank, total_nodes_destroyed,
	          total_turns_survived, created_at, last_login 
	          FROM players WHERE username = $1`
	
	err := DB.QueryRow(query, username).Scan(
		&player.ID,
		&player.Username,
		&player.EloRating,
		&player.Wins,
		&player.Losses,
		&player.CurrentStreak,
		&player.BestStreak,
		&player.RankProtectionGames,
		&player.CurrentRank,
		&player.BestRank,
		&player.TotalNodesDestroyed,
		&player.TotalTurnsSurvived,
		&player.CreatedAt,
		&player.LastLogin,
	)
	if err != nil {
		return nil, err
	}
	return player, nil
}

func CreatePlayer(player *Player) error {
	query := `INSERT INTO players (id, username, password_hash, elo_rating, wins, losses, 
	          current_streak, best_streak, rank_protection_games, current_rank, best_rank,
	          total_nodes_destroyed, total_turns_survived) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	
	_, err := DB.Exec(query,
		player.ID,
		player.Username,
		player.PasswordHash,
		player.EloRating,
		player.Wins,
		player.Losses,
		player.CurrentStreak,
		player.BestStreak,
		player.RankProtectionGames,
		player.CurrentRank,
		player.BestRank,
		player.TotalNodesDestroyed,
		player.TotalTurnsSurvived,
	)
	return err
}

func UpdatePlayerElo(playerID string, newElo int, newRank string, protectionGames int) error {
	query := `UPDATE players SET elo_rating = $1, current_rank = $2, rank_protection_games = $3, 
	          last_login = CURRENT_TIMESTAMP WHERE id = $4`
	_, err := DB.Exec(query, newElo, newRank, protectionGames, playerID)
	return err
}

func UpdatePlayerStats(playerID string, isWin bool, nodesDestroyed int, turnsSurvived int) error {
	var winChange, lossChange int
	var streakChange int
	
	if isWin {
		winChange = 1
		streakChange = 1
	} else {
		lossChange = 1
		streakChange = 0
	}
	
	query := `UPDATE players SET 
	          wins = wins + $1, 
	          losses = losses + $2,
	          current_streak = CASE WHEN $3 = 1 THEN current_streak + 1 ELSE 0 END,
	          best_streak = CASE WHEN (CASE WHEN $3 = 1 THEN current_streak + 1 ELSE 0 END) > best_streak 
	                        THEN (CASE WHEN $3 = 1 THEN current_streak + 1 ELSE 0 END) ELSE best_streak END,
	          total_nodes_destroyed = total_nodes_destroyed + $4,
	          total_turns_survived = total_turns_survived + $5,
	          last_login = CURRENT_TIMESTAMP
	          WHERE id = $6`
	
	_, err := DB.Exec(query, winChange, lossChange, streakChange, nodesDestroyed, turnsSurvived, playerID)
	return err
}

func UpdateBestRank(playerID string, newBestRank string) error {
	query := `UPDATE players SET best_rank = $1 WHERE id = $2 AND best_rank < $3`
	_, err := DB.Exec(query, newBestRank, playerID, newBestRank)
	return err
}

type LeaderboardEntry struct {
	Rank      int    `json:"rank"`
	PlayerID  string `json:"playerId"`
	Username  string `json:"username"`
	EloRating int    `json:"eloRating"`
	Wins      int    `json:"wins"`
	Losses    int    `json:"losses"`
	RankName  string `json:"rankName"`
	WinRate   float64 `json:"winRate"`
}

func GetLeaderboard(limit int) ([]*LeaderboardEntry, error) {
	query := `SELECT id, username, elo_rating, wins, losses, current_rank,
	          CASE WHEN (wins + losses) > 0 THEN ROUND((wins::float / (wins + losses)) * 100, 1) ELSE 0 END as win_rate
	          FROM players 
	          ORDER BY elo_rating DESC, wins DESC 
	          LIMIT $1`
	
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	entries := make([]*LeaderboardEntry, 0, limit)
	rank := 1
	for rows.Next() {
		entry := &LeaderboardEntry{}
		err := rows.Scan(&entry.PlayerID, &entry.Username, &entry.EloRating, 
			&entry.Wins, &entry.Losses, &entry.RankName, &entry.WinRate)
		if err != nil {
			return nil, err
		}
		entry.Rank = rank
		rank++
		entries = append(entries, entry)
	}
	
	return entries, nil
}

func GetPlayerRank(playerID string) (int, error) {
	var rank int
	query := `SELECT COUNT(*) + 1 FROM players WHERE elo_rating > (
	          SELECT elo_rating FROM players WHERE id = $1)`
	
	err := DB.QueryRow(query, playerID).Scan(&rank)
	if err != nil {
		return 0, err
	}
	return rank, nil
}

type CardUsageStat struct {
	CardType    string `json:"cardType"`
	UsageCount  int    `json:"usageCount"`
	CardName    string `json:"cardName"`
}

func GetTopCards(playerID string, limit int) ([]*CardUsageStat, error) {
	query := `SELECT card_type, usage_count 
	          FROM card_usage_stats 
	          WHERE player_id = $1 
	          ORDER BY usage_count DESC 
	          LIMIT $2`
	
	rows, err := DB.Query(query, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	stats := make([]*CardUsageStat, 0)
	for rows.Next() {
		stat := &CardUsageStat{}
		err := rows.Scan(&stat.CardType, &stat.UsageCount)
		if err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}
	
	return stats, nil
}

func UpdateCardUsage(playerID string, cardType string, count int) error {
	query := `INSERT INTO card_usage_stats (player_id, card_type, usage_count) 
	          VALUES ($1, $2, $3)
	          ON CONFLICT (player_id, card_type) 
	          DO UPDATE SET usage_count = card_usage_stats.usage_count + EXCLUDED.usage_count`
	
	_, err := DB.Exec(query, playerID, cardType, count)
	return err
}

type PlayerStatsSummary struct {
	TotalGames      int     `json:"totalGames"`
	Wins            int     `json:"wins"`
	Losses          int     `json:"losses"`
	WinRate         float64 `json:"winRate"`
	CurrentStreak   int     `json:"currentStreak"`
	BestStreak      int     `json:"bestStreak"`
	AvgNodesDestroyed float64 `json:"avgNodesDestroyed"`
	AvgTurnsSurvived float64 `json:"avgTurnsSurvived"`
	TopCards        []*CardUsageStat `json:"topCards"`
}

func GetPlayerStatsSummary(playerID string) (*PlayerStatsSummary, error) {
	query := `SELECT 
	          wins + losses as total_games,
	          wins,
	          losses,
	          CASE WHEN (wins + losses) > 0 THEN ROUND((wins::float / (wins + losses)) * 100, 1) ELSE 0 END as win_rate,
	          current_streak,
	          best_streak,
	          CASE WHEN (wins + losses) > 0 THEN ROUND((total_nodes_destroyed::float / (wins + losses)), 1) ELSE 0 END as avg_nodes_destroyed,
	          CASE WHEN (wins + losses) > 0 THEN ROUND((total_turns_survived::float / (wins + losses)), 1) ELSE 0 END as avg_turns_survived
	          FROM players WHERE id = $1`
	
	stats := &PlayerStatsSummary{}
	err := DB.QueryRow(query, playerID).Scan(
		&stats.TotalGames,
		&stats.Wins,
		&stats.Losses,
		&stats.WinRate,
		&stats.CurrentStreak,
		&stats.BestStreak,
		&stats.AvgNodesDestroyed,
		&stats.AvgTurnsSurvived,
	)
	if err != nil {
		return nil, err
	}
	
	topCards, err := GetTopCards(playerID, 3)
	if err != nil {
		stats.TopCards = make([]*CardUsageStat, 0)
	} else {
		stats.TopCards = topCards
	}
	
	return stats, nil
}

func RecordGameResult(gameID string, roomID string, playerIDs []string, winnerID string, gameMode string, turns int, duration int, seasonID int) error {
	query := `INSERT INTO game_records (id, room_id, player_ids, winner_id, game_mode, turns, duration, season_id) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	
	_, err := DB.Exec(query, gameID, roomID, pq.Array(playerIDs), winnerID, gameMode, turns, duration, seasonID)
	return err
}

func RecordPlayerGameStat(gameID string, playerID string, nodesDestroyed int, cardsPlayed int, 
	damageDealt int, damageTaken int, coreHpRemaining int, result string, 
	eloChange int, eloAfter int, rankAfter string, rankChange string) error {
	
	query := `INSERT INTO player_game_stats 
	          (game_id, player_id, nodes_destroyed, cards_played, damage_dealt, damage_taken, 
	           core_hp_remaining, result, elo_change, elo_after, rank_after, rank_change) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	
	_, err := DB.Exec(query, gameID, playerID, nodesDestroyed, cardsPlayed, 
		damageDealt, damageTaken, coreHpRemaining, result, eloChange, eloAfter, rankAfter, rankChange)
	return err
}

type Season struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
	IsActive  bool      `json:"isActive"`
}

func GetActiveSeason() (*Season, error) {
	season := &Season{}
	query := `SELECT id, name, start_date, end_date, is_active FROM seasons WHERE is_active = true LIMIT 1`
	
	err := DB.QueryRow(query).Scan(&season.ID, &season.Name, &season.StartDate, &season.EndDate, &season.IsActive)
	if err != nil {
		return nil, err
	}
	return season, nil
}

func GetAllSeasons() ([]*Season, error) {
	query := `SELECT id, name, start_date, end_date, is_active FROM seasons ORDER BY id DESC`
	
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	seasons := make([]*Season, 0)
	for rows.Next() {
		season := &Season{}
		err := rows.Scan(&season.ID, &season.Name, &season.StartDate, &season.EndDate, &season.IsActive)
		if err != nil {
			return nil, err
		}
		seasons = append(seasons, season)
	}
	
	return seasons, nil
}

func CreateNewSeason(name string, durationDays int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	
	_, err = tx.Exec(`UPDATE seasons SET is_active = false WHERE is_active = true`)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	query := `INSERT INTO seasons (name, start_date, end_date, is_active) 
	          VALUES ($1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + ($2 || ' days')::INTERVAL, true)`
	_, err = tx.Exec(query, name, durationDays)
	if err != nil {
		tx.Rollback()
		return err
	}
	
	return tx.Commit()
}

func ResetAllPlayersElo() error {
	query := `UPDATE players SET 
	          elo_rating = 1200 + (elo_rating - 1200) / 2,
	          current_rank = CASE 
	            WHEN 1200 + (elo_rating - 1200) / 2 >= 2000 THEN 'diamond'
	            WHEN 1200 + (elo_rating - 1200) / 2 >= 1700 THEN 'platinum'
	            WHEN 1200 + (elo_rating - 1200) / 2 >= 1400 THEN 'gold'
	            WHEN 1200 + (elo_rating - 1200) / 2 >= 1100 THEN 'silver'
	            ELSE 'bronze'
	          END,
	          rank_protection_games = 0,
	          current_streak = 0`
	_, err := DB.Exec(query)
	return err
}

var _ = pq.Array
