package websocket

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"cyberhack/internal/database"
	"cyberhack/internal/matchmaking"
	"cyberhack/internal/room"
	"cyberhack/internal/season"
	"cyberhack/internal/tournament"
)

type Hub struct {
	clients         map[string]*Client
	rooms           map[string]*room.Room
	matchmaker      *matchmaking.Matchmaker
	tournamentMgr   *tournament.TournamentManager
	tournamentViewers map[string]map[string]bool
	Register        chan *Client
	Unregister      chan *Client
	mu              sync.RWMutex
	ticker          *time.Ticker
	stopChan        chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients:          make(map[string]*Client),
		rooms:            make(map[string]*room.Room),
		matchmaker:       matchmaking.NewMatchmaker(),
		tournamentMgr:    tournament.GetManager(),
		tournamentViewers: make(map[string]map[string]bool),
		Register:         make(chan *Client),
		Unregister:       make(chan *Client),
		stopChan:         make(chan struct{}),
	}
}

func (h *Hub) Run() {
	h.matchmaker.SetMatchCallback(h.onMatchFound)
	h.matchmaker.Start()

	h.tournamentMgr.SetBroadcastFunc(h.broadcastToTournament)
	h.tournamentMgr.SetSendToPlayerFunc(h.sendToClient)
	h.tournamentMgr.Start()

	h.ticker = time.NewTicker(1 * time.Second)
	go h.tick()

	season.GetManager().SetBroadcastFunc(h.broadcastToAll)
	season.GetManager().Start()

	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client registered: %s (%s)", client.ID, client.Username)

			go h.sendPlayerInfo(client)
			go h.broadcastOnlineCount()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				h.handleDisconnect(client)
			}
			h.mu.Unlock()
			log.Printf("Client unregistered: %s", client.ID)

			go h.broadcastOnlineCount()

		case <-h.stopChan:
			return
		}
	}
}

func (h *Hub) Stop() {
	if h.ticker != nil {
		h.ticker.Stop()
	}
	h.matchmaker.Stop()
	h.tournamentMgr.Stop()
	season.GetManager().Stop()
	close(h.stopChan)
}

func (h *Hub) tick() {
	for range h.ticker.C {
		h.updateMatchmakingStatus()
		h.broadcastOnlineCount()
	}
}

func (h *Hub) broadcastOnlineCount() {
	h.mu.RLock()
	count := len(h.clients)
	h.mu.RUnlock()

	h.broadcastToAll("online_count", map[string]interface{}{
		"count": count,
	})
}

func (h *Hub) updateMatchmakingStatus() {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		info := h.matchmaker.GetMatchRangeInfo(client.ID, "quick", 2)
		if info != nil {
			client.SendMessage("matchmaking_status", map[string]interface{}{
				"waitTime":       info.WaitTime,
				"estimatedRange": info.EstimatedRange,
				"currentRange":   info.CurrentRange,
			})
		}
	}
}

func (h *Hub) onMatchFound(queueKey string, result *matchmaking.MatchResult) {
	h.createMatchedRoom(result)
}

func (h *Hub) sendPlayerInfo(client *Client) {
	player, err := database.GetPlayerByID(client.ID)
	if err != nil {
		newPlayer := &database.Player{
			ID:                client.ID,
			Username:          client.Username,
			EloRating:         1200,
			CurrentRank:       "bronze",
			BestRank:          "bronze",
			RankProtectionGames: 0,
		}
		database.CreatePlayer(newPlayer)
		player = newPlayer
	}

	seasonInfo := season.GetManager().GetCurrentSeason()
	seasonData := map[string]interface{}{}
	if seasonInfo != nil {
		seasonData = map[string]interface{}{
			"id":          seasonInfo.ID,
			"name":        seasonInfo.Name,
			"startDate":   seasonInfo.StartDate,
			"endDate":     seasonInfo.EndDate,
			"daysRemaining": int(time.Until(seasonInfo.EndDate).Hours() / 24),
		}
	}

	client.SendMessage("player_info", map[string]interface{}{
		"playerId":          player.ID,
		"username":          player.Username,
		"eloRating":         player.EloRating,
		"currentRank":       player.CurrentRank,
		"bestRank":          player.BestRank,
		"wins":              player.Wins,
		"losses":            player.Losses,
		"currentStreak":     player.CurrentStreak,
		"bestStreak":        player.BestStreak,
		"rankProtectionGames": player.RankProtectionGames,
		"season":            seasonData,
	})
}

func (h *Hub) handleMessage(client *Client, msg *Message) {
	switch msg.Type {
	case "quick_match":
		h.handleQuickMatch(client)
	case "cancel_match":
		h.handleCancelMatch(client)
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
	case "get_player_info":
		h.sendPlayerInfo(client)
	case "tournament_create":
		h.handleTournamentCreate(client, msg)
	case "tournament_list":
		h.handleTournamentList(client)
	case "tournament_join":
		h.handleTournamentJoin(client, msg)
	case "tournament_leave":
		h.handleTournamentLeave(client, msg)
	case "tournament_watch":
		h.handleTournamentWatch(client, msg)
	case "tournament_unwatch":
		h.handleTournamentUnwatch(client, msg)
	case "tournament_detail":
		h.handleTournamentDetail(client, msg)
	case "tournament_bracket":
		h.handleTournamentBracket(client, msg)
	case "tournament_chat":
		h.handleTournamentChat(client, msg)
	case "tournament_chat_history":
		h.handleTournamentChatHistory(client, msg)
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

func (h *Hub) broadcastToAll(msgType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		client.SendMessage(msgType, payload)
	}
}

func (h *Hub) handleQuickMatch(client *Client) {
	player, err := database.GetPlayerByID(client.ID)
	eloRating := 1200
	if err == nil {
		eloRating = player.EloRating
	}

	matchRequest := &matchmaking.MatchRequest{
		PlayerID:   client.ID,
		Username:   client.Username,
		Client:     client,
		GameMode:   "quick",
		MaxPlayers: 2,
		EloRating:  eloRating,
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
			"position":       1,
			"eloRating":      eloRating,
			"estimatedRange": "±200",
			"waitTime":       0,
		})
	}
}

func (h *Hub) handleCancelMatch(client *Client) {
	h.matchmaker.RemoveFromQueue(client.ID)
	client.SendMessage("matchmaking_cancelled", map[string]string{
		"message": "已取消匹配",
	})
}

func (h *Hub) createMatchedRoom(match *matchmaking.MatchResult) {
	r := room.NewRoom("", 2)
	r.SetBroadcastFunc(h.sendToClient)
	
	h.mu.Lock()
	h.rooms[r.ID] = r
	h.mu.Unlock()

	for _, player := range match.Players {
		r.AddPlayer(player.PlayerID, player.Username)
	}

	r.Broadcast("matched", map[string]interface{}{
		"roomId":  r.ID,
		"players": r.GetPlayerList(),
	})

	r.ForceStartGame()
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

func (h *Hub) broadcastToTournament(tournamentID, msgType string, payload interface{}) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	viewers, exists := h.tournamentViewers[tournamentID]
	if !exists {
		if tournamentID == "" {
			for _, client := range h.clients {
				client.SendMessage(msgType, payload)
			}
		}
		return
	}

	for playerID := range viewers {
		if client, ok := h.clients[playerID]; ok {
			client.SendMessage(msgType, payload)
		}
	}
}

func (h *Hub) handleTournamentCreate(client *Client, msg *Message) {
	var config struct {
		Name             string `json:"name"`
		MaxPlayers       int    `json:"maxPlayers"`
		MinRank          string `json:"minRank"`
		DurationMinutes  int    `json:"durationMinutes"`
	}
	json.Unmarshal(msg.Payload, &config)

	if config.Name == "" {
		client.SendMessage("error", map[string]string{"message": "锦标赛名称不能为空"})
		return
	}

	if config.MaxPlayers != 8 && config.MaxPlayers != 16 && config.MaxPlayers != 32 {
		config.MaxPlayers = 8
	}

	if config.DurationMinutes != 5 && config.DurationMinutes != 10 && config.DurationMinutes != 15 {
		config.DurationMinutes = 10
	}

	if config.MinRank == "" {
		config.MinRank = "none"
	}

	t, err := h.tournamentMgr.CreateTournament(client.ID, client.Username, config.Name, 
		config.MaxPlayers, config.MinRank, config.DurationMinutes)
	if err != nil {
		client.SendMessage("error", map[string]string{"message": "创建锦标赛失败: " + err.Error()})
		return
	}

	client.SendMessage("tournament_created", map[string]interface{}{
		"tournamentId": t.ID,
		"tournament":   t,
	})

	err = h.tournamentMgr.RegisterPlayer(t.ID, client.ID, client.Username)
	if err != nil {
		client.SendMessage("error", map[string]string{"message": "自动报名失败: " + err.Error()})
	}
}

func (h *Hub) handleTournamentList(client *Client) {
	tournaments := h.tournamentMgr.GetActiveTournaments()
	tournamentList := make([]map[string]interface{}, 0)

	for _, t := range tournaments {
		playerCount, _ := database.GetTournamentPlayerCount(t.ID)
		tournamentList = append(tournamentList, map[string]interface{}{
			"id":                   t.ID,
			"name":                 t.Name,
			"maxPlayers":           t.MaxPlayers,
			"minRank":              t.MinRank,
			"status":               t.Status,
			"registrationDeadline": t.RegistrationDeadline,
			"currentRound":         t.CurrentRound,
			"totalRounds":          t.TotalRounds,
			"playerCount":          playerCount,
		})
	}

	client.SendMessage("tournament_list", map[string]interface{}{
		"tournaments": tournamentList,
	})
}

func (h *Hub) handleTournamentJoin(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" {
		client.SendMessage("error", map[string]string{"message": "锦标赛ID不能为空"})
		return
	}

	err := h.tournamentMgr.RegisterPlayer(payload.TournamentID, client.ID, client.Username)
	if err != nil {
		client.SendMessage("error", map[string]string{"message": err.Error()})
		return
	}

	client.SendMessage("tournament_joined", map[string]interface{}{
		"tournamentId": payload.TournamentID,
	})
}

func (h *Hub) handleTournamentLeave(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" {
		return
	}

	database.RemoveTournamentPlayer(payload.TournamentID, client.ID)
	client.SendMessage("tournament_left", map[string]interface{}{
		"tournamentId": payload.TournamentID,
	})
}

func (h *Hub) handleTournamentWatch(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" {
		client.SendMessage("error", map[string]string{"message": "锦标赛ID不能为空"})
		return
	}

	h.mu.Lock()
	if h.tournamentViewers[payload.TournamentID] == nil {
		h.tournamentViewers[payload.TournamentID] = make(map[string]bool)
	}
	h.tournamentViewers[payload.TournamentID][client.ID] = true
	h.mu.Unlock()

	state, err := h.tournamentMgr.GetTournament(payload.TournamentID)
	if err != nil {
		client.SendMessage("error", map[string]string{"message": "锦标赛不存在"})
		return
	}

	matches, _ := h.tournamentMgr.GetBracket(payload.TournamentID)
	messages, _ := h.tournamentMgr.GetChatMessages(payload.TournamentID, 50)

	client.SendMessage("tournament_watching", map[string]interface{}{
		"tournamentId": payload.TournamentID,
		"tournament":   state.Tournament,
		"players":      state.Players,
		"bracket":      matches,
		"chatMessages": messages,
		"playerCount":  len(state.Players),
	})
}

func (h *Hub) handleTournamentUnwatch(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" {
		return
	}

	h.mu.Lock()
	if viewers, ok := h.tournamentViewers[payload.TournamentID]; ok {
		delete(viewers, client.ID)
		if len(viewers) == 0 {
			delete(h.tournamentViewers, payload.TournamentID)
		}
	}
	h.mu.Unlock()

	client.SendMessage("tournament_unwatched", map[string]interface{}{
		"tournamentId": payload.TournamentID,
	})
}

func (h *Hub) handleTournamentDetail(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" {
		client.SendMessage("error", map[string]string{"message": "锦标赛ID不能为空"})
		return
	}

	data, err := h.tournamentMgr.GetTournamentWithPlayerCount(payload.TournamentID)
	if err != nil {
		client.SendMessage("error", map[string]string{"message": "锦标赛不存在"})
		return
	}

	client.SendMessage("tournament_detail", data)
}

func (h *Hub) handleTournamentBracket(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" {
		client.SendMessage("error", map[string]string{"message": "锦标赛ID不能为空"})
		return
	}

	matches, err := h.tournamentMgr.GetBracket(payload.TournamentID)
	if err != nil {
		client.SendMessage("error", map[string]string{"message": "获取对阵表失败"})
		return
	}

	client.SendMessage("tournament_bracket", map[string]interface{}{
		"tournamentId": payload.TournamentID,
		"matches":      matches,
	})
}

func (h *Hub) handleTournamentChat(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
		Message      string `json:"message"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" || payload.Message == "" {
		return
	}

	h.tournamentMgr.SendChatMessage(payload.TournamentID, client.ID, client.Username, payload.Message)
}

func (h *Hub) handleTournamentChatHistory(client *Client, msg *Message) {
	var payload struct {
		TournamentID string `json:"tournamentId"`
		Limit        int    `json:"limit"`
	}
	json.Unmarshal(msg.Payload, &payload)

	if payload.TournamentID == "" {
		return
	}

	if payload.Limit <= 0 || payload.Limit > 100 {
		payload.Limit = 50
	}

	messages, err := h.tournamentMgr.GetChatMessages(payload.TournamentID, payload.Limit)
	if err != nil {
		return
	}

	client.SendMessage("tournament_chat_history", map[string]interface{}{
		"tournamentId": payload.TournamentID,
		"messages":     messages,
	})
}

func (h *Hub) findRoomByID(roomID string) *room.Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.rooms[roomID]
}
