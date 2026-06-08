package leaderboard

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cyberhack/internal/database"
	redisClient "cyberhack/internal/redis"
)

const (
	leaderboardKey = "leaderboard:global"
	cacheTTL       = 5 * time.Minute
)

type LeaderboardEntry struct {
	Rank      int     `json:"rank"`
	PlayerID  string  `json:"playerId"`
	Username  string  `json:"username"`
	EloRating int     `json:"eloRating"`
	Wins      int     `json:"wins"`
	Losses    int     `json:"losses"`
	RankName  string  `json:"rankName"`
	WinRate   float64 `json:"winRate"`
}

func GetTopPlayers(limit int) ([]*LeaderboardEntry, error) {
	cached, err := getFromCache(limit)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	entries, err := loadFromDB(limit)
	if err != nil {
		log.Printf("Error loading leaderboard from DB: %v", err)
		return make([]*LeaderboardEntry, 0), nil
	}

	if len(entries) > 0 {
		go saveToCache(entries)
	}

	return entries, nil
}

func getFromCache(limit int) ([]*LeaderboardEntry, error) {
	if redisClient.Client == nil {
		return nil, fmt.Errorf("redis not available")
	}

	ctx := context.Background()
	data, err := redisClient.Client.Get(ctx, leaderboardKey).Bytes()
	if err != nil {
		return nil, err
	}

	var allEntries []*LeaderboardEntry
	err = json.Unmarshal(data, &allEntries)
	if err != nil {
		return nil, err
	}

	if limit > len(allEntries) {
		limit = len(allEntries)
	}

	return allEntries[:limit], nil
}

func saveToCache(entries []*LeaderboardEntry) {
	if redisClient.Client == nil {
		return
	}

	ctx := context.Background()
	data, err := json.Marshal(entries)
	if err != nil {
		return
	}

	redisClient.Client.Set(ctx, leaderboardKey, data, cacheTTL)
}

func loadFromDB(limit int) ([]*LeaderboardEntry, error) {
	dbEntries, err := database.GetLeaderboard(limit)
	if err != nil {
		return nil, err
	}

	entries := make([]*LeaderboardEntry, len(dbEntries))
	for i, dbEntry := range dbEntries {
		entries[i] = &LeaderboardEntry{
			Rank:      dbEntry.Rank,
			PlayerID:  dbEntry.PlayerID,
			Username:  dbEntry.Username,
			EloRating: dbEntry.EloRating,
			Wins:      dbEntry.Wins,
			Losses:    dbEntry.Losses,
			RankName:  dbEntry.RankName,
			WinRate:   dbEntry.WinRate,
		}
	}

	return entries, nil
}

func InvalidateCache() {
	if redisClient.Client == nil {
		return
	}

	ctx := context.Background()
	redisClient.Client.Del(ctx, leaderboardKey)
}

func GetPlayerRankAndEntry(playerID string) (*LeaderboardEntry, int, error) {
	rank, err := database.GetPlayerRank(playerID)
	if err != nil {
		return nil, 0, err
	}

	player, err := database.GetPlayerByID(playerID)
	if err != nil {
		return nil, 0, err
	}

	var winRate float64
	totalGames := player.Wins + player.Losses
	if totalGames > 0 {
		winRate = float64(player.Wins) / float64(totalGames) * 100
	}

	entry := &LeaderboardEntry{
		Rank:      rank,
		PlayerID:  player.ID,
		Username:  player.Username,
		EloRating: player.EloRating,
		Wins:      player.Wins,
		Losses:    player.Losses,
		RankName:  player.CurrentRank,
		WinRate:   winRate,
	}

	return entry, rank, nil
}
