package server

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	ws "cyberhack/internal/websocket"
	"cyberhack/internal/database"
	"cyberhack/internal/leaderboard"
	"cyberhack/internal/replay"
	redisClient "cyberhack/internal/redis"
	"cyberhack/internal/season"

	"github.com/google/uuid"
	gorillaWs "github.com/gorilla/websocket"
)

var upgrader = gorillaWs.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	hub *ws.Hub
}

func NewServer(hub *ws.Hub) *Server {
	return &Server{hub: hub}
}

func (s *Server) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		playerID = uuid.New().String()
	}

	username := r.URL.Query().Get("username")
	if username == "" {
		username = "Hacker-" + playerID[:6]
	}

	client := ws.NewClient(playerID, username, conn, s.hub)
	s.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}

func (s *Server) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/ws", s.ServeWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/api/replays", s.handleGetReplays)
	mux.HandleFunc("/api/replay", s.handleGetReplay)
	mux.HandleFunc("/api/player", s.handleGetPlayer)
	mux.HandleFunc("/api/player/stats", s.handleGetPlayerStats)
	mux.HandleFunc("/api/leaderboard", s.handleGetLeaderboard)
	mux.HandleFunc("/api/season", s.handleGetSeason)
}

func (s *Server) handleGetReplays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	roomID := r.URL.Query().Get("roomId")
	if roomID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "roomId is required"})
		return
	}

	replays := replay.GetStore().GetReplays(roomID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"replays": replays,
	})
}

func (s *Server) handleGetReplay(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	roomID := r.URL.Query().Get("roomId")
	replayID := r.URL.Query().Get("replayId")

	if roomID == "" || replayID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "roomId and replayId are required"})
		return
	}

	replayData := replay.GetStore().GetReplay(roomID, replayID)
	if replayData == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "replay not found"})
		return
	}

	json.NewEncoder(w).Encode(replayData)
}

func (s *Server) handleGetPlayer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	playerID := r.URL.Query().Get("playerId")
	username := r.URL.Query().Get("username")

	var player *database.Player
	var err error

	if playerID != "" {
		player, err = database.GetPlayerByID(playerID)
	} else if username != "" {
		player, err = database.GetPlayerByUsername(username)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "playerId or username is required"})
		return
	}

	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "player not found"})
		return
	}

	rank, err := database.GetPlayerRank(player.ID)
	if err != nil {
		rank = 0
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"player": player,
		"rank":   rank,
	})
}

func (s *Server) handleGetPlayerStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	playerID := r.URL.Query().Get("playerId")
	if playerID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "playerId is required"})
		return
	}

	stats, err := database.GetPlayerStatsSummary(playerID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "player stats not found"})
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleGetLeaderboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	entries, err := leaderboard.GetTopPlayers(limit)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to get leaderboard"})
		return
	}

	playerID := r.URL.Query().Get("playerId")
	var playerEntry *leaderboard.LeaderboardEntry
	var playerRank int

	if playerID != "" {
		playerEntry, playerRank, _ = leaderboard.GetPlayerRankAndEntry(playerID)
	}

	response := map[string]interface{}{
		"leaderboard": entries,
	}

	if playerEntry != nil {
		response["playerRank"] = map[string]interface{}{
			"rank":     playerRank,
			"player":   playerEntry,
			"isInTop":  playerRank <= limit,
		}
	}

	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleGetSeason(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
		return
	}

	currentSeason := season.GetManager().GetCurrentSeason()
	if currentSeason == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "no active season"})
		return
	}

	timeRemaining := season.GetManager().GetTimeRemaining()
	totalSeconds := int(timeRemaining.Seconds())
	daysRemaining := int(timeRemaining.Hours() / 24)
	hoursRemaining := int(timeRemaining.Hours()) % 24

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            currentSeason.ID,
		"name":          currentSeason.Name,
		"startDate":     currentSeason.StartDate,
		"endDate":       currentSeason.EndDate,
		"isActive":      currentSeason.IsActive,
		"timeRemaining": totalSeconds,
		"daysRemaining": daysRemaining,
		"hoursRemaining": hoursRemaining,
		"totalSeconds":  totalSeconds,
	})
}

func Start() {
	if err := redisClient.Init(); err != nil {
		log.Printf("Warning: Redis connection failed: %v", err)
		log.Println("Running without Redis - using in-memory storage")
	} else {
		log.Println("Redis connected successfully")
	}

	if err := database.Init(); err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Running without database")
	} else {
		log.Println("Database connected successfully")
	}

	hub := ws.NewHub()
	go hub.Run()

	server := NewServer(hub)

	mux := http.NewServeMux()
	server.SetupRoutes(mux)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("CyberHack server starting on port %s", port)
	log.Printf("WebSocket endpoint: ws://localhost:%s/ws", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
