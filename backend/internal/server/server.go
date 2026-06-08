package server

import (
	"log"
	"net/http"
	"os"

	ws "cyberhack/internal/websocket"
	"cyberhack/internal/database"
	redisClient "cyberhack/internal/redis"

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
