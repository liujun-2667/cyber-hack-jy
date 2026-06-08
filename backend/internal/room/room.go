package room

import (
	"sync"
	"time"

	"cyberhack/internal/game"
	"github.com/google/uuid"
)

type Room struct {
	ID           string
	Game         *game.Game
	players      map[string]*RoomPlayer
	playerOrder  []string
	maxPlayers   int
	mu           sync.RWMutex
	phaseTimer   *time.Timer
	readyPlayers map[string]bool
	gameStarted  bool
	hubBroadcast func(playerID string, msgType string, payload interface{})
}

type RoomPlayer struct {
	ID       string
	Username string
}

func NewRoom(id string, maxPlayers int) *Room {
	if id == "" {
		id = uuid.New().String()[:8]
	}
	config := game.DefaultGameConfig()
	config.MaxPlayers = maxPlayers

	return &Room{
		ID:           id,
		Game:         game.NewGame(id, config),
		players:      make(map[string]*RoomPlayer),
		playerOrder:  make([]string, 0),
		maxPlayers:   maxPlayers,
		readyPlayers: make(map[string]bool),
		gameStarted:  false,
	}
}

func (r *Room) SetBroadcastFunc(fn func(string, string, interface{})) {
	r.hubBroadcast = fn
}

func (r *Room) AddPlayer(playerID, username string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.players) >= r.maxPlayers {
		return false
	}
	if r.gameStarted {
		return false
	}

	r.players[playerID] = &RoomPlayer{
		ID:       playerID,
		Username: username,
	}
	r.playerOrder = append(r.playerOrder, playerID)
	r.Game.AddPlayer(playerID, username)

	return true
}

func (r *Room) RemovePlayer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.players, playerID)
	for i, id := range r.playerOrder {
		if id == playerID {
			r.playerOrder = append(r.playerOrder[:i], r.playerOrder[i+1:]...)
			break
		}
	}
	delete(r.readyPlayers, playerID)
}

func (r *Room) HasPlayer(playerID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.players[playerID]
	return exists
}

func (r *Room) GetPlayerList() []map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	players := make([]map[string]string, 0, len(r.players))
	for _, player := range r.players {
		players = append(players, map[string]string{
			"id":       player.ID,
			"username": player.Username,
		})
	}
	return players
}

func (r *Room) PlaceNode(playerID, nodeType string, x, y int) {
	if !r.gameStarted || r.Game.Phase != game.PhasePlacement {
		return
	}

	r.Game.PlaceNode(playerID, game.NodeType(nodeType), x, y)
	r.Broadcast("node_placed", map[string]interface{}{
		"playerId": playerID,
		"nodeType": nodeType,
		"x":        x,
		"y":        y,
	})
}

func (r *Room) StartGame(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.gameStarted {
		return
	}

	r.readyPlayers[playerID] = true

	allReady := len(r.players) >= 2 && len(r.readyPlayers) == len(r.players)
	if !allReady {
		return
	}

	r.gameStarted = true
	r.Game.StartGame()

	r.Broadcast("game_started", map[string]interface{}{
		"phase":     "placement",
		"playerIds": r.playerOrder,
	})

	r.autoPlaceNodes()
	r.startFirstTurn()
}

func (r *Room) autoPlaceNodes() {
	for _, playerID := range r.playerOrder {
		player := r.Game.Players[playerID]
		if player == nil {
			continue
		}
		
		nodes := []game.NodeType{
			game.NodeTypeRouter,
			game.NodeTypeServer,
			game.NodeTypeDatabase,
			game.NodeTypeFirewall,
			game.NodeTypeRouter,
			game.NodeTypeServer,
			game.NodeTypeFirewall,
			game.NodeTypeDatabase,
			game.NodeTypeRouter,
			game.NodeTypeServer,
			game.NodeTypeRouter,
			game.NodeTypeFirewall,
		}
		
		positions := [][2]int{
			{0, 0}, {0, 2}, {0, 4},
			{1, 1}, {1, 3},
			{2, 0}, {2, 4},
			{3, 1}, {3, 3},
			{4, 0}, {4, 2}, {4, 4},
		}
		
		for i, pos := range positions {
			if i >= len(nodes) {
				break
			}
			player.PlaceNode(nodes[i], pos[0], pos[1])
		}
	}
}

func (r *Room) startFirstTurn() {
	r.Game.StartFirstTurn()
	r.startProgrammingPhase()
}

func (r *Room) startProgrammingPhase() {
	r.Broadcast("phase_change", map[string]interface{}{
		"phase": "programming",
		"turn":  r.Game.CurrentTurn,
		"time":  r.Game.Config.ProgrammingTime,
	})

	r.sendAllPlayerStates()

	if r.phaseTimer != nil {
		r.phaseTimer.Stop()
	}
	r.phaseTimer = time.AfterFunc(time.Duration(r.Game.Config.ProgrammingTime)*time.Second, func() {
		r.ExecutePhase()
	})
}

func (r *Room) PlayCard(playerID, cardID, targetNodeID, targetPlayerID string) bool {
	if r.Game.Phase != game.PhaseProgramming {
		return false
	}

	return r.Game.PlayCard(playerID, cardID, targetNodeID, targetPlayerID)
}

func (r *Room) PlayerReady(playerID string) {
	r.mu.Lock()
	r.readyPlayers[playerID] = true
	r.mu.Unlock()

	allReady := true
	r.mu.RLock()
	for id := range r.players {
		if !r.readyPlayers[id] {
			allReady = false
			break
		}
	}
	r.mu.RUnlock()

	if allReady && r.Game.Phase == game.PhaseProgramming {
		if r.phaseTimer != nil {
			r.phaseTimer.Stop()
		}
		r.ExecutePhase()
	}
}

func (r *Room) ExecutePhase() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game.Phase != game.PhaseProgramming {
		return
	}

	r.readyPlayers = make(map[string]bool)

	r.Game.ExecutePhase()

	r.Broadcast("execution_result", map[string]interface{}{
		"actions": r.Game.TurnActions,
		"gameLog": r.Game.GameLog,
	})

	r.sendAllPlayerStates()

	if r.Game.Phase == game.PhaseGameOver {
		r.Broadcast("game_over", map[string]interface{}{
			"winnerId": r.Game.WinnerID,
			"turns":    r.Game.CurrentTurn,
		})
		return
	}

	time.AfterFunc(3*time.Second, func() {
		r.startProgrammingPhase()
	})
}

func (r *Room) sendAllPlayerStates() {
	for playerID := range r.players {
		state := r.GetPlayerState(playerID)
		r.SendToPlayer(playerID, "game_state", state)
	}
}

func (r *Room) GetPlayerState(playerID string) map[string]interface{} {
	player, exists := r.players[playerID]
	if !exists {
		return nil
	}

	gamePlayer := r.Game.Players[playerID]
	if gamePlayer == nil {
		return nil
	}

	state := map[string]interface{}{
		"phase":       string(r.Game.Phase),
		"currentTurn": r.Game.CurrentTurn,
		"myPlayerId":  playerID,
		"myUsername":  player.Username,
		"hand":        gamePlayer.Hand,
		"playedCards": gamePlayer.PlayedCards,
		"cooldowns":   gamePlayer.Cooldowns,
		"myGrid":      gridToArray(gamePlayer.Grid),
		"coreHp":      gamePlayer.CoreNode.HP,
		"coreMaxHp":   gamePlayer.CoreNode.MaxHP,
		"opponents":   make(map[string]interface{}),
		"gameLog":     r.Game.GameLog,
	}

	opponents := make(map[string]interface{})
	for _, oppID := range r.playerOrder {
		if oppID == playerID {
			continue
		}
		opp := r.Game.Players[oppID]
		if opp == nil {
			continue
		}

		oppGrid := make([][]interface{}, 5)
		for x := 0; x < 5; x++ {
			oppGrid[x] = make([]interface{}, 5)
			for y := 0; y < 5; y++ {
				node := opp.Grid[x][y]
				if node == nil {
					oppGrid[x][y] = nil
					continue
				}
				if gamePlayer.IsNodeRevealed(oppID, node.ID) || node.Status.ScanRevealed {
					oppGrid[x][y] = node
				} else {
					oppGrid[x][y] = map[string]interface{}{
						"id":      node.ID,
						"x":       node.X,
						"y":       node.Y,
						"unknown": true,
						"alive":   node.IsAlive(),
					}
				}
			}
		}

		coreHp := 0
		coreMaxHp := 0
		if opp.CoreNode != nil {
			coreHp = opp.CoreNode.HP
			coreMaxHp = opp.CoreNode.MaxHP
		}

		opponents[oppID] = map[string]interface{}{
			"id":         oppID,
			"username":   opp.Username,
			"grid":       oppGrid,
			"coreHp":     coreHp,
			"coreMaxHp":  coreMaxHp,
			"isAlive":    opp.IsAlive,
		}
	}
	state["opponents"] = opponents

	return state
}

func gridToArray(grid [5][5]*game.Node) [][]interface{} {
	result := make([][]interface{}, 5)
	for x := 0; x < 5; x++ {
		result[x] = make([]interface{}, 5)
		for y := 0; y < 5; y++ {
			if grid[x][y] != nil {
				result[x][y] = grid[x][y]
			} else {
				result[x][y] = nil
			}
		}
	}
	return result
}

func (r *Room) Broadcast(msgType string, payload interface{}) {
	if r.hubBroadcast == nil {
		return
	}
	r.mu.RLock()
	defer r.mu.RUnlock()

	for playerID := range r.players {
		r.hubBroadcast(playerID, msgType, payload)
	}
}

func (r *Room) SendToPlayer(playerID, msgType string, payload interface{}) {
	if r.hubBroadcast == nil {
		return
	}
	r.hubBroadcast(playerID, msgType, payload)
}
