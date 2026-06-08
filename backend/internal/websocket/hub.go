package websocket

import (
	"encoding/json"
	"log"
	"sync"

	"cyberhack/internal/room"
	"cyberhack/internal/matchmaking"
)

type Hub struct {
	clients    map[string]*Client
	rooms      map[string]*room.Room
	matchmaker *matchmaking.Matchmaker
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]*room.Room),
		matchmaker: matchmaking.NewMatchmaker(),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (%s)", client.ID, client.Username)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				h.handleDisconnect(client)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.ID)
		}
	}
}

func (h *Hub) handleMessage(client *Client, msg *Message) {
	switch msg.Type {
	case "quick_match":
		h.handleQuickMatch(client)
	case "create_room":
		h.handleCreateRoom(client, msg)
	case "join_room":
		h.handleJoinRoom(client, msg)
	case "leave_room":
		h.handleLeaveRoom(client)
	case "place_node":
		h.handlePlaceNode(client, msg)
	case "start_game":
		h.handleStartGame(client)
	case "play_card":
		h.handlePlayCard(client, msg)
	case "end_turn":
		h.handleEndTurn(client)
	case "chat":
		h.handleChat(client, msg)
	case "game_state":
		h.handleGameStateRequest(client)
	}
}

func (h *Hub) sendToClient(playerID, msgType string, payload interface{}) {
	h.mu.RLock()
	client, exists := h.clients[playerID]
	h.mu.RUnlock()

	if !exists {
		return
	}
	client.SendMessage(msgType, payload)
}

func (h *Hub) handleQuickMatch(client *Client) {
	matchRequest := &matchmaking.MatchRequest{
		PlayerID:   client.ID,
		Username:   client.Username,
		Client:     client,
		GameMode:   "quick",
		MaxPlayers: 2,
	}

	match, err := h.matchmaker.AddToQueue(matchRequest)
	if err != nil {
		client.SendMessage("error", map[string]string{"message": err.Error()})
		return
	}

	if match != nil {
		h.createMatchedRoom(match)
	} else {
		client.SendMessage("matchmaking_queued", map[string]interface{}{
			"position": 1,
		})
	}
}

func (h *Hub) createMatchedRoom(match *matchmaking.MatchResult) {
	r := room.NewRoom("", 2)
	r.SetBroadcastFunc(h.sendToClient)
	
	h.mu.Lock()
	h.rooms[r.ID] = r
	h.mu.Unlock()

	for _, player := range match.Players {
		r.AddPlayer(player.PlayerID, player.Username)
		h.sendToClient(player.PlayerID, "matched", map[string]interface{}{
			"roomId":  r.ID,
			"players": r.GetPlayerList(),
		})
	}

	r.StartGame(match.Players[0].PlayerID)
}

func (h *Hub) handleCreateRoom(client *Client, msg *Message) {
	var config struct {
		MaxPlayers int    `json:"maxPlayers"`
		GameMode   string `json:"gameMode"`
	}
	json.Unmarshal(msg.Payload, &config)

	if config.MaxPlayers < 2 || config.MaxPlayers > 4 {
		config.MaxPlayers = 2
	}

	r := room.NewRoom("", config.MaxPlayers)
	r.SetBroadcastFunc(h.sendToClient)
	
	h.mu.Lock()
	h.rooms[r.ID] = r
	h.mu.Unlock()

	r.AddPlayer(client.ID, client.Username)

	client.SendMessage("room_created", map[string]interface{}{
		"roomId":     r.ID,
		"maxPlayers": config.MaxPlayers,
		"players":    r.GetPlayerList(),
	})
}

func (h *Hub) handleJoinRoom(client *Client, msg *Message) {
	var payload struct {
		RoomID string `json:"roomId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	h.mu.RLock()
	r, exists := h.rooms[payload.RoomID]
	h.mu.RUnlock()

	if !exists {
		client.SendMessage("error", map[string]string{"message": "房间不存在"})
		return
	}

	if !r.AddPlayer(client.ID, client.Username) {
		client.SendMessage("error", map[string]string{"message": "房间已满"})
		return
	}

	client.SendMessage("room_joined", map[string]interface{}{
		"roomId":  r.ID,
		"players": r.GetPlayerList(),
	})

	r.Broadcast("player_joined", map[string]interface{}{
		"playerId": client.ID,
		"username": client.Username,
	})
}

func (h *Hub) handleLeaveRoom(client *Client) {
	h.mu.RLock()
	var targetRoom *room.Room
	for _, r := range h.rooms {
		if r.HasPlayer(client.ID) {
			targetRoom = r
			break
		}
	}
	h.mu.RUnlock()

	if targetRoom != nil {
		targetRoom.RemovePlayer(client.ID)
		targetRoom.Broadcast("player_left", map[string]string{
			"playerId": client.ID,
		})
	}
}

func (h *Hub) findRoomByPlayer(playerID string) *room.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, r := range h.rooms {
		if r.HasPlayer(playerID) {
			return r
		}
	}
	return nil
}

func (h *Hub) handlePlaceNode(client *Client, msg *Message) {
	r := h.findRoomByPlayer(client.ID)
	if r == nil {
		return
	}

	var payload struct {
		NodeType string `json:"nodeType"`
		X        int    `json:"x"`
		Y        int    `json:"y"`
	}
	json.Unmarshal(msg.Payload, &payload)

	r.PlaceNode(client.ID, payload.NodeType, payload.X, payload.Y)
}

func (h *Hub) handleStartGame(client *Client) {
	r := h.findRoomByPlayer(client.ID)
	if r == nil {
		return
	}

	r.StartGame(client.ID)
}

func (h *Hub) handlePlayCard(client *Client, msg *Message) {
	r := h.findRoomByPlayer(client.ID)
	if r == nil {
		return
	}

	var payload struct {
		CardID         string `json:"cardId"`
		TargetNodeID   string `json:"targetNodeId"`
		TargetPlayerID string `json:"targetPlayerId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	r.PlayCard(client.ID, payload.CardID, payload.TargetNodeID, payload.TargetPlayerID)
}

func (h *Hub) handleEndTurn(client *Client) {
	r := h.findRoomByPlayer(client.ID)
	if r == nil {
		return
	}

	r.PlayerReady(client.ID)
}

func (h *Hub) handleChat(client *Client, msg *Message) {
	r := h.findRoomByPlayer(client.ID)
	if r == nil {
		return
	}

	var payload struct {
		Message string `json:"message"`
	}
	json.Unmarshal(msg.Payload, &payload)

	r.Broadcast("chat", map[string]interface{}{
		"playerId": client.ID,
		"username": client.Username,
		"message":  payload.Message,
	})
}

func (h *Hub) handleGameStateRequest(client *Client) {
	r := h.findRoomByPlayer(client.ID)
	if r == nil {
		return
	}

	state := r.GetPlayerState(client.ID)
	client.SendMessage("game_state", state)
}

func (h *Hub) handleDisconnect(client *Client) {
	h.matchmaker.RemoveFromQueue(client.ID)

	for _, r := range h.rooms {
		if r.HasPlayer(client.ID) {
			r.RemovePlayer(client.ID)
			r.Broadcast("player_disconnected", map[string]string{
				"playerId": client.ID,
			})
		}
	}
}

func (h *Hub) GetClient(id string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	client, exists := h.clients[id]
	return client, exists
}
