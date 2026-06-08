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
	if DB == nil {
		return make([]*LeaderboardEntry, 0), nil
	}

	query := `SELECT id, username, elo_rating, wins, losses, current_rank,
	          CASE WHEN (wins + losses) > 0 THEN ROUND((wins::float / (wins + losses)) * 100, 1) ELSE 0 END as win_rate
	          FROM players 
	          ORDER BY elo_rating DESC, wins DESC 
	          LIMIT $1`
	
	rows, err := DB.Query(query, limit)
	if err != nil {
		return make([]*LeaderboardEntry, 0), nil
	}
	defer rows.Close()
	
	entries := make([]*LeaderboardEntry, 0, limit)
	rank := 1
	for rows.Next() {
		entry := &LeaderboardEntry{}
		err := rows.Scan(&entry.PlayerID, &entry.Username, &entry.EloRating, 
			&entry.Wins, &entry.Losses, &entry.RankName, &entry.WinRate)
		if err != nil {
			continue
		}
		entry.Rank = rank
		rank++
		entries = append(entries, entry)
	}
	
	return entries, nil
}

func GetPlayerRank(playerID string) (int, error) {
	if DB == nil {
		return 9999, nil
	}

	var rank int
	query := `SELECT COUNT(*) + 1 FROM players WHERE elo_rating > (
	          SELECT elo_rating FROM players WHERE id = $1)`
	
	err := DB.QueryRow(query, playerID).Scan(&rank)
	if err != nil {
		return 9999, nil
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

func RecordGameCardStats(gameID string, playerID string, cardStats map[string]int) error {
	if DB == nil {
		return nil
	}
	
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	stmt, err := tx.Prepare(`INSERT INTO game_card_stats (game_id, player_id, card_type, usage_count) 
	                          VALUES ($1, $2, $3, $4)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	for cardType, count := range cardStats {
		_, err = stmt.Exec(gameID, playerID, cardType, count)
		if err != nil {
			return err
		}
	}
	
	return tx.Commit()
}

func UpdateCardUsage(playerID string, cardType string, count int) error {
	if DB == nil {
		return nil
	}

	query := `INSERT INTO card_usage_stats (player_id, card_type, usage_count) 
	          VALUES ($1, $2, $3)
	          ON CONFLICT (player_id, card_type) 
	          DO UPDATE SET usage_count = card_usage_stats.usage_count + EXCLUDED.usage_count`
	
	_, err := DB.Exec(query, playerID, cardType, count)
	return err
}

type RecentGameRecord struct {
	GameID        string    `json:"gameId"`
	OpponentID    string    `json:"opponentId"`
	OpponentName  string    `json:"opponentName"`
	Result        string    `json:"result"`
	EloChange     int       `json:"eloChange"`
	TopCard       string    `json:"topCard"`
	TopCardCount  int       `json:"topCardCount"`
	CreatedAt     time.Time `json:"createdAt"`
}

func GetRecentGames(playerID string, limit int) ([]*RecentGameRecord, error) {
	if DB == nil {
		return make([]*RecentGameRecord, 0), nil
	}

	query := `
		SELECT 
			gr.id as game_id,
			gr.created_at,
			pgs.result,
			pgs.elo_change,
			opp.username as opponent_name,
			opp.id as opponent_id,
			gcs.card_type as top_card,
			gcs.usage_count as top_card_count
		FROM game_records gr
		JOIN player_game_stats pgs ON pgs.game_id = gr.id AND pgs.player_id = $1
		JOIN player_game_stats opp_pgs ON opp_pgs.game_id = gr.id AND opp_pgs.player_id != $1
		JOIN players opp ON opp.id = opp_pgs.player_id
		LEFT JOIN game_card_stats gcs ON gcs.game_id = gr.id AND gcs.player_id = $1
			AND gcs.usage_count = (
				SELECT MAX(usage_count) 
				FROM game_card_stats 
				WHERE game_id = gr.id AND player_id = $1
			)
		ORDER BY gr.created_at DESC
		LIMIT $2
	`

	rows, err := DB.Query(query, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*RecentGameRecord, 0)
	for rows.Next() {
		record := &RecentGameRecord{}
		var opponentName, opponentID, topCard sql.NullString
		var topCardCount sql.NullInt32
		
		err := rows.Scan(
			&record.GameID,
			&record.CreatedAt,
			&record.Result,
			&record.EloChange,
			&opponentName,
			&opponentID,
			&topCard,
			&topCardCount,
		)
		if err != nil {
			continue
		}
		
		if opponentName.Valid {
			record.OpponentName = opponentName.String
		} else {
			record.OpponentName = "未知对手"
		}
		if opponentID.Valid {
			record.OpponentID = opponentID.String
		}
		if topCard.Valid {
			record.TopCard = topCard.String
		} else {
			record.TopCard = ""
		}
		if topCardCount.Valid {
			record.TopCardCount = int(topCardCount.Int32)
		}
		
		records = append(records, record)
	}

	return records, nil
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
	player, err := GetPlayerByID(playerID)
	if err != nil {
		return nil, fmt.Errorf("player not found")
	}

	totalGames := player.Wins + player.Losses
	winRate := 0.0
	if totalGames > 0 {
		winRate = float64(player.Wins) / float64(totalGames) * 100
	}

	avgNodes := 0.0
	avgTurns := 0.0
	if totalGames > 0 {
		avgNodes = float64(player.TotalNodesDestroyed) / float64(totalGames)
		avgTurns = float64(player.TotalTurnsSurvived) / float64(totalGames)
	}

	stats := &PlayerStatsSummary{
		TotalGames:        totalGames,
		Wins:              player.Wins,
		Losses:            player.Losses,
		WinRate:           roundFloat(winRate, 1),
		CurrentStreak:     player.CurrentStreak,
		BestStreak:        player.BestStreak,
		AvgNodesDestroyed: roundFloat(avgNodes, 1),
		AvgTurnsSurvived:  roundFloat(avgTurns, 1),
		TopCards:          make([]*CardUsageStat, 0),
	}

	topCards, err := GetTopCards(playerID, 3)
	if err == nil && len(topCards) > 0 {
		stats.TopCards = topCards
	}

	return stats, nil
}

func roundFloat(val float64, precision int) float64 {
	ratio := 1.0
	for i := 0; i < precision; i++ {
		ratio *= 10
	}
	return float64(int(val*ratio+0.5)) / ratio
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

type Tournament struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	MaxPlayers         int       `json:"maxPlayers"`
	MinRank            string    `json:"minRank"`
	CreatorID          string    `json:"creatorId"`
	Status             string    `json:"status"`
	RegistrationDeadline time.Time `json:"registrationDeadline"`
	StartedAt          time.Time `json:"startedAt"`
	EndedAt            time.Time `json:"endedAt"`
	WinnerID           string    `json:"winnerId"`
	CurrentRound       int       `json:"currentRound"`
	TotalRounds        int       `json:"totalRounds"`
	CreatedAt          time.Time `json:"createdAt"`
}

type TournamentPlayer struct {
	ID           int       `json:"id"`
	TournamentID string    `json:"tournamentId"`
	PlayerID     string    `json:"playerId"`
	Username     string    `json:"username"`
	EloRating    int       `json:"eloRating"`
	CurrentRank  string    `json:"currentRank"`
	Seed         int       `json:"seed"`
	FinalPosition int      `json:"finalPosition"`
	JoinedAt     time.Time `json:"joinedAt"`
}

type TournamentMatch struct {
	ID          string    `json:"id"`
	TournamentID string   `json:"tournamentId"`
	RoundNumber int       `json:"roundNumber"`
	MatchIndex  int       `json:"matchIndex"`
	Player1ID   string    `json:"player1Id"`
	Player2ID   string    `json:"player2Id"`
	Player1Name string    `json:"player1Name"`
	Player2Name string    `json:"player2Name"`
	WinnerID    string    `json:"winnerId"`
	RoomID      string    `json:"roomId"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"startedAt"`
	EndedAt     time.Time `json:"endedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TournamentChatMessage struct {
	ID         int       `json:"id"`
	TournamentID string   `json:"tournamentId"`
	PlayerID   string    `json:"playerId"`
	Username   string    `json:"username"`
	Message    string    `json:"message"`
	IsSystem   bool      `json:"isSystem"`
	CreatedAt  time.Time `json:"createdAt"`
}

type TournamentRecord struct {
	ID              int       `json:"id"`
	TournamentID    string    `json:"tournamentId"`
	PlayerID        string    `json:"playerId"`
	TournamentName  string    `json:"tournamentName"`
	FinalPosition   int       `json:"finalPosition"`
	TotalMatches    int       `json:"totalMatches"`
	Wins            int       `json:"wins"`
	Losses          int       `json:"losses"`
	EloBonus        int       `json:"eloBonus"`
	HasTop4Badge    bool      `json:"hasTop4Badge"`
	CreatedAt       time.Time `json:"createdAt"`
}

func CreateTournament(t *Tournament) error {
	query := `INSERT INTO tournaments (id, name, max_players, min_rank, creator_id, status, registration_deadline, total_rounds)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := DB.Exec(query, t.ID, t.Name, t.MaxPlayers, t.MinRank, t.CreatorID, t.Status, t.RegistrationDeadline, t.TotalRounds)
	return err
}

func GetTournamentByID(tournamentID string) (*Tournament, error) {
	t := &Tournament{}
	query := `SELECT id, name, max_players, min_rank, creator_id, status, registration_deadline, 
	          started_at, ended_at, winner_id, current_round, total_rounds, created_at
	          FROM tournaments WHERE id = $1`
	err := DB.QueryRow(query, tournamentID).Scan(
		&t.ID, &t.Name, &t.MaxPlayers, &t.MinRank, &t.CreatorID, &t.Status,
		&t.RegistrationDeadline, &t.StartedAt, &t.EndedAt, &t.WinnerID,
		&t.CurrentRound, &t.TotalRounds, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func GetTournamentsByStatus(status string) ([]*Tournament, error) {
	query := `SELECT id, name, max_players, min_rank, creator_id, status, registration_deadline,
	          started_at, ended_at, winner_id, current_round, total_rounds, created_at
	          FROM tournaments WHERE status = $1 ORDER BY created_at DESC`
	rows, err := DB.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tournaments := make([]*Tournament, 0)
	for rows.Next() {
		t := &Tournament{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.MaxPlayers, &t.MinRank, &t.CreatorID, &t.Status,
			&t.RegistrationDeadline, &t.StartedAt, &t.EndedAt, &t.WinnerID,
			&t.CurrentRound, &t.TotalRounds, &t.CreatedAt,
		)
		if err != nil {
			continue
		}
		tournaments = append(tournaments, t)
	}
	return tournaments, nil
}

func GetActiveTournaments() ([]*Tournament, error) {
	query := `SELECT id, name, max_players, min_rank, creator_id, status, registration_deadline,
	          started_at, ended_at, winner_id, current_round, total_rounds, created_at
	          FROM tournaments WHERE status IN ('registering', 'in_progress') ORDER BY created_at DESC`
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tournaments := make([]*Tournament, 0)
	for rows.Next() {
		t := &Tournament{}
		err := rows.Scan(
			&t.ID, &t.Name, &t.MaxPlayers, &t.MinRank, &t.CreatorID, &t.Status,
			&t.RegistrationDeadline, &t.StartedAt, &t.EndedAt, &t.WinnerID,
			&t.CurrentRound, &t.TotalRounds, &t.CreatedAt,
		)
		if err != nil {
			continue
		}
		tournaments = append(tournaments, t)
	}
	return tournaments, nil
}

func UpdateTournamentStatus(tournamentID, status string) error {
	query := `UPDATE tournaments SET status = $1 WHERE id = $2`
	_, err := DB.Exec(query, status, tournamentID)
	return err
}

func UpdateTournamentRound(tournamentID string, currentRound int) error {
	query := `UPDATE tournaments SET current_round = $1 WHERE id = $2`
	_, err := DB.Exec(query, currentRound, tournamentID)
	return err
}

func StartTournament(tournamentID string) error {
	query := `UPDATE tournaments SET status = 'in_progress', started_at = CURRENT_TIMESTAMP, current_round = 1 WHERE id = $1`
	_, err := DB.Exec(query, tournamentID)
	return err
}

func EndTournament(tournamentID, winnerID string) error {
	query := `UPDATE tournaments SET status = 'finished', ended_at = CURRENT_TIMESTAMP, winner_id = $1 WHERE id = $2`
	_, err := DB.Exec(query, winnerID, tournamentID)
	return err
}

func AddTournamentPlayer(tp *TournamentPlayer) error {
	query := `INSERT INTO tournament_players (tournament_id, player_id, username, elo_rating, current_rank, seed)
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := DB.Exec(query, tp.TournamentID, tp.PlayerID, tp.Username, tp.EloRating, tp.CurrentRank, tp.Seed)
	return err
}

func GetTournamentPlayers(tournamentID string) ([]*TournamentPlayer, error) {
	query := `SELECT id, tournament_id, player_id, username, elo_rating, current_rank, seed, final_position, joined_at
	          FROM tournament_players WHERE tournament_id = $1 ORDER BY seed ASC, joined_at ASC`
	rows, err := DB.Query(query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]*TournamentPlayer, 0)
	for rows.Next() {
		p := &TournamentPlayer{}
		err := rows.Scan(&p.ID, &p.TournamentID, &p.PlayerID, &p.Username,
			&p.EloRating, &p.CurrentRank, &p.Seed, &p.FinalPosition, &p.JoinedAt)
		if err != nil {
			continue
		}
		players = append(players, p)
	}
	return players, nil
}

func GetTournamentPlayerCount(tournamentID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM tournament_players WHERE tournament_id = $1`
	err := DB.QueryRow(query, tournamentID).Scan(&count)
	return count, err
}

func IsPlayerInTournament(tournamentID, playerID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM tournament_players WHERE tournament_id = $1 AND player_id = $2)`
	err := DB.QueryRow(query, tournamentID, playerID).Scan(&exists)
	return exists, err
}

func RemoveTournamentPlayer(tournamentID, playerID string) error {
	query := `DELETE FROM tournament_players WHERE tournament_id = $1 AND player_id = $2`
	_, err := DB.Exec(query, tournamentID, playerID)
	return err
}

func UpdatePlayerSeed(tournamentID, playerID string, seed int) error {
	query := `UPDATE tournament_players SET seed = $1 WHERE tournament_id = $2 AND player_id = $3`
	_, err := DB.Exec(query, seed, tournamentID, playerID)
	return err
}

func UpdatePlayerFinalPosition(tournamentID, playerID string, position int) error {
	query := `UPDATE tournament_players SET final_position = $1 WHERE tournament_id = $2 AND player_id = $3`
	_, err := DB.Exec(query, position, tournamentID, playerID)
	return err
}

func CreateTournamentMatch(m *TournamentMatch) error {
	query := `INSERT INTO tournament_matches (id, tournament_id, round_number, match_index, 
	          player1_id, player2_id, player1_name, player2_name, status)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := DB.Exec(query, m.ID, m.TournamentID, m.RoundNumber, m.MatchIndex,
		m.Player1ID, m.Player2ID, m.Player1Name, m.Player2Name, m.Status)
	return err
}

func GetTournamentMatches(tournamentID string) ([]*TournamentMatch, error) {
	query := `SELECT id, tournament_id, round_number, match_index, player1_id, player2_id,
	          player1_name, player2_name, winner_id, room_id, status, started_at, ended_at, created_at
	          FROM tournament_matches WHERE tournament_id = $1 ORDER BY round_number ASC, match_index ASC`
	rows, err := DB.Query(query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]*TournamentMatch, 0)
	for rows.Next() {
		m := &TournamentMatch{}
		err := rows.Scan(&m.ID, &m.TournamentID, &m.RoundNumber, &m.MatchIndex,
			&m.Player1ID, &m.Player2ID, &m.Player1Name, &m.Player2Name,
			&m.WinnerID, &m.RoomID, &m.Status, &m.StartedAt, &m.EndedAt, &m.CreatedAt)
		if err != nil {
			continue
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func GetTournamentMatchesByRound(tournamentID string, round int) ([]*TournamentMatch, error) {
	query := `SELECT id, tournament_id, round_number, match_index, player1_id, player2_id,
	          player1_name, player2_name, winner_id, room_id, status, started_at, ended_at, created_at
	          FROM tournament_matches WHERE tournament_id = $1 AND round_number = $2 ORDER BY match_index ASC`
	rows, err := DB.Query(query, tournamentID, round)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	matches := make([]*TournamentMatch, 0)
	for rows.Next() {
		m := &TournamentMatch{}
		err := rows.Scan(&m.ID, &m.TournamentID, &m.RoundNumber, &m.MatchIndex,
			&m.Player1ID, &m.Player2ID, &m.Player1Name, &m.Player2Name,
			&m.WinnerID, &m.RoomID, &m.Status, &m.StartedAt, &m.EndedAt, &m.CreatedAt)
		if err != nil {
			continue
		}
		matches = append(matches, m)
	}
	return matches, nil
}

func UpdateTournamentMatchRoom(matchID, roomID string) error {
	query := `UPDATE tournament_matches SET room_id = $1, status = 'in_progress', started_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := DB.Exec(query, roomID, matchID)
	return err
}

func UpdateTournamentMatchResult(matchID, winnerID string) error {
	query := `UPDATE tournament_matches SET winner_id = $1, status = 'finished', ended_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := DB.Exec(query, winnerID, matchID)
	return err
}

func GetTournamentMatchByID(matchID string) (*TournamentMatch, error) {
	m := &TournamentMatch{}
	query := `SELECT id, tournament_id, round_number, match_index, player1_id, player2_id,
	          player1_name, player2_name, winner_id, room_id, status, started_at, ended_at, created_at
	          FROM tournament_matches WHERE id = $1`
	err := DB.QueryRow(query, matchID).Scan(&m.ID, &m.TournamentID, &m.RoundNumber, &m.MatchIndex,
		&m.Player1ID, &m.Player2ID, &m.Player1Name, &m.Player2Name,
		&m.WinnerID, &m.RoomID, &m.Status, &m.StartedAt, &m.EndedAt, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func AddTournamentChat(tournamentID, playerID, username, message string, isSystem bool) error {
	query := `INSERT INTO tournament_chat (tournament_id, player_id, username, message, is_system)
	          VALUES ($1, $2, $3, $4, $5)`
	_, err := DB.Exec(query, tournamentID, playerID, username, message, isSystem)
	return err
}

func GetTournamentChat(tournamentID string, limit int) ([]*TournamentChatMessage, error) {
	query := `SELECT id, tournament_id, player_id, username, message, is_system, created_at
	          FROM tournament_chat WHERE tournament_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := DB.Query(query, tournamentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*TournamentChatMessage, 0)
	for rows.Next() {
		msg := &TournamentChatMessage{}
		err := rows.Scan(&msg.ID, &msg.TournamentID, &msg.PlayerID, &msg.Username,
			&msg.Message, &msg.IsSystem, &msg.CreatedAt)
		if err != nil {
			continue
		}
		messages = append([]*TournamentChatMessage{msg}, messages...)
	}
	return messages, nil
}

func CreateTournamentRecord(tr *TournamentRecord) error {
	query := `INSERT INTO tournament_records (tournament_id, player_id, tournament_name, final_position, 
	          total_matches, wins, losses, elo_bonus, has_top4_badge)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err := DB.Exec(query, tr.TournamentID, tr.PlayerID, tr.TournamentName, tr.FinalPosition,
		tr.TotalMatches, tr.Wins, tr.Losses, tr.EloBonus, tr.HasTop4Badge)
	return err
}

func GetPlayerTournamentRecords(playerID string, limit int) ([]*TournamentRecord, error) {
	query := `SELECT id, tournament_id, player_id, tournament_name, final_position, 
	          total_matches, wins, losses, elo_bonus, has_top4_badge, created_at
	          FROM tournament_records WHERE player_id = $1 ORDER BY created_at DESC LIMIT $2`
	rows, err := DB.Query(query, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := make([]*TournamentRecord, 0)
	for rows.Next() {
		r := &TournamentRecord{}
		err := rows.Scan(&r.ID, &r.TournamentID, &r.PlayerID, &r.TournamentName, &r.FinalPosition,
			&r.TotalMatches, &r.Wins, &r.Losses, &r.EloBonus, &r.HasTop4Badge, &r.CreatedAt)
		if err != nil {
			continue
		}
		records = append(records, r)
	}
	return records, nil
}

var _ = pq.Array
