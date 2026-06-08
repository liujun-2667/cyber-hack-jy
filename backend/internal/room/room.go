package room

import (
	"log"
	"sync"
	"time"

	"cyberhack/internal/database"
	"cyberhack/internal/game"
	"cyberhack/internal/leaderboard"
	"cyberhack/internal/rank"
	"cyberhack/internal/replay"
	"cyberhack/internal/season"
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

	if r.gameStarted {
		r.mu.Unlock()
		return
	}

	r.readyPlayers[playerID] = true

	allReady := len(r.players) >= 2 && len(r.readyPlayers) == len(r.players)
	if !allReady {
		r.mu.Unlock()
		return
	}

	r.gameStarted = true
	r.Game.StartGame()
	r.autoPlaceNodes()
	r.mu.Unlock()

	r.Broadcast("game_started", map[string]interface{}{
		"phase":     "placement",
		"playerIds": r.playerOrder,
	})

	r.startFirstTurn()
}

func (r *Room) ForceStartGame() {
	r.mu.Lock()

	if r.gameStarted {
		r.mu.Unlock()
		return
	}

	r.gameStarted = true
	r.Game.StartGame()
	r.autoPlaceNodes()
	r.mu.Unlock()

	r.Broadcast("game_started", map[string]interface{}{
		"phase":     "placement",
		"playerIds": r.playerOrder,
	})

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
	r.mu.Lock()
	currentTurn := r.Game.CurrentTurn
	programmingTime := r.Game.Config.ProgrammingTime
	if r.phaseTimer != nil {
		r.phaseTimer.Stop()
	}
	r.phaseTimer = time.AfterFunc(time.Duration(programmingTime)*time.Second, func() {
		r.ExecutePhase()
	})
	r.mu.Unlock()

	r.Broadcast("phase_change", map[string]interface{}{
		"phase": "programming",
		"turn":  currentTurn,
		"time":  programmingTime,
	})

	r.sendAllPlayerStates()
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

	allReady := true
	for id := range r.players {
		if !r.readyPlayers[id] {
			allReady = false
			break
		}
	}

	isProgrammingPhase := r.Game.Phase == game.PhaseProgramming
	if allReady && isProgrammingPhase {
		if r.phaseTimer != nil {
			r.phaseTimer.Stop()
			r.phaseTimer = nil
		}
	}
	r.mu.Unlock()

	if allReady && isProgrammingPhase {
		r.ExecutePhase()
	}
}

func (r *Room) ExecutePhase() {
	r.mu.Lock()

	if r.Game.Phase != game.PhaseProgramming {
		r.mu.Unlock()
		return
	}

	r.readyPlayers = make(map[string]bool)
	r.Game.ExecutePhase()

	isGameOver := r.Game.Phase == game.PhaseGameOver
	winnerID := r.Game.WinnerID
	turns := r.Game.CurrentTurn
	actions := r.Game.TurnActions
	gameLog := r.Game.GameLog

	r.mu.Unlock()

	r.Broadcast("execution_result", map[string]interface{}{
		"actions": actions,
		"gameLog": gameLog,
	})

	r.sendAllPlayerStates()

	if isGameOver {
		replayData := replay.CreateReplay(r.ID, r.Game)
		replay.GetStore().AddReplay(r.ID, replayData)

		rankResults := r.calculateRankResults(winnerID)

		r.Broadcast("game_over", map[string]interface{}{
			"winnerId":    winnerID,
			"turns":       turns,
			"replayId":    replayData.ID,
			"rankResults": rankResults,
		})
		return
	}

	time.AfterFunc(3*time.Second, func() {
		r.startProgrammingPhase()
	})
}

func (r *Room) sendAllPlayerStates() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for playerID := range r.players {
		state := r.getPlayerStateLocked(playerID)
		if state != nil {
			r.hubBroadcast(playerID, "game_state", state)
		}
	}
}

func (r *Room) getPlayerStateLocked(playerID string) map[string]interface{} {
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

func (r *Room) GetPlayerState(playerID string) map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getPlayerStateLocked(playerID)
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

type PlayerRankResult struct {
	PlayerID     string `json:"playerId"`
	Username     string `json:"username"`
	IsWinner     bool   `json:"isWinner"`
	EloChange    int    `json:"eloChange"`
	OldElo       int    `json:"oldElo"`
	NewElo       int    `json:"newElo"`
	OldRank      string `json:"oldRank"`
	NewRank      string `json:"newRank"`
	RankChange   string `json:"rankChange"`
	BoundaryBonus bool  `json:"boundaryBonus"`
}

func (r *Room) calculateRankResults(winnerID string) map[string]*PlayerRankResult {
	results := make(map[string]*PlayerRankResult)
	
	if len(r.playerOrder) < 2 {
		return results
	}
	
	playerIDs := make([]string, 0, len(r.playerOrder))
	playerElos := make(map[string]int)
	playerRanks := make(map[string]string)
	playerProtections := make(map[string]int)
	
	for _, playerID := range r.playerOrder {
		player, err := database.GetPlayerByID(playerID)
		if err != nil {
			log.Printf("Error getting player %s: %v", playerID, err)
			playerElos[playerID] = 1200
			playerRanks[playerID] = "bronze"
			playerProtections[playerID] = 0
		} else {
			playerElos[playerID] = player.EloRating
			playerRanks[playerID] = player.CurrentRank
			playerProtections[playerID] = player.RankProtectionGames
		}
		playerIDs = append(playerIDs, playerID)
	}
	
	if len(playerIDs) < 2 {
		return results
	}
	
	p1ID := playerIDs[0]
	p2ID := playerIDs[1]
	
	var winnerElo, loserElo int
	var winnerIDResult, loserID string
	var winnerRank, loserRank string
	var winnerProt, loserProt int
	
	if p1ID == winnerID {
		winnerElo = playerElos[p1ID]
		loserElo = playerElos[p2ID]
		winnerIDResult = p1ID
		loserID = p2ID
		winnerRank = playerRanks[p1ID]
		loserRank = playerRanks[p2ID]
		winnerProt = playerProtections[p1ID]
		loserProt = playerProtections[p2ID]
	} else {
		winnerElo = playerElos[p2ID]
		loserElo = playerElos[p1ID]
		winnerIDResult = p2ID
		loserID = p1ID
		winnerRank = playerRanks[p2ID]
		loserRank = playerRanks[p1ID]
		winnerProt = playerProtections[p2ID]
		loserProt = playerProtections[p1ID]
	}
	
	eloResult := rank.CalculateEloChange(winnerElo, loserElo)
	
	winnerNewElo := eloResult.WinnerNewElo
	winnerNewRank := rank.Rank(winnerRank)
	winnerNewProt := winnerProt
	if winnerIDResult == p1ID || winnerIDResult == p2ID {
		winnerNewElo, winnerNewRank, winnerNewProt = rank.ApplyRankProtection(
			winnerElo, eloResult.WinnerNewElo, winnerProt, rank.Rank(winnerRank))
	}
	
	loserNewElo, loserNewRank, loserNewProt := rank.ApplyRankProtection(
		loserElo, eloResult.LoserNewElo, loserProt, rank.Rank(loserRank))
	
	rankChangeWinner := "none"
	if rank.GetRankIndex(rank.Rank(winnerNewRank)) > rank.GetRankIndex(rank.Rank(winnerRank)) {
		rankChangeWinner = "promote"
	}
	
	rankChangeLoser := "none"
	if rank.GetRankIndex(loserNewRank) < rank.GetRankIndex(rank.Rank(loserRank)) {
		rankChangeLoser = "demote"
	}
	
	winnerGamePlayer := r.Game.Players[winnerIDResult]
	loserGamePlayer := r.Game.Players[loserID]
	
	winnerNodesDestroyed := countDestroyedNodes(loserGamePlayer)
	loserNodesDestroyed := countDestroyedNodes(winnerGamePlayer)
	
	currentSeason := season.GetManager().GetCurrentSeason()
	seasonID := 1
	if currentSeason != nil {
		seasonID = currentSeason.ID
	}
	
	gameID := uuid.New().String()
	turns := r.Game.CurrentTurn
	
	go func() {
		database.RecordGameResult(gameID, r.ID, []string{winnerIDResult, loserID}, 
			winnerID, r.Game.Config.GameMode, turns, 0, seasonID)
		
		database.RecordPlayerGameStat(gameID, winnerIDResult, winnerNodesDestroyed, 
			countCardsPlayed(winnerGamePlayer), 0, 0, 
			getCoreHp(winnerGamePlayer), "win",
			int(winnerNewElo)-winnerElo, int(winnerNewElo), 
			string(winnerNewRank), rankChangeWinner)
		
		database.RecordPlayerGameStat(gameID, loserID, loserNodesDestroyed,
			countCardsPlayed(loserGamePlayer), 0, 0,
			getCoreHp(loserGamePlayer), "loss",
			int(loserNewElo)-loserElo, int(loserNewElo),
			string(loserNewRank), rankChangeLoser)
		
		cardsPlayedWinner := getPlayedCardTypes(winnerGamePlayer)
		cardsPlayedLoser := getPlayedCardTypes(loserGamePlayer)
		
		database.RecordGameCardStats(gameID, winnerIDResult, cardsPlayedWinner)
		database.RecordGameCardStats(gameID, loserID, cardsPlayedLoser)
		
		database.UpdatePlayerStats(winnerIDResult, true, winnerNodesDestroyed, turns)
		database.UpdatePlayerStats(loserID, false, loserNodesDestroyed, turns)
		
		database.UpdatePlayerElo(winnerIDResult, int(winnerNewElo), string(winnerNewRank), winnerNewProt)
		database.UpdatePlayerElo(loserID, int(loserNewElo), string(loserNewRank), loserNewProt)
		
		if rank.GetRankIndex(rank.Rank(winnerNewRank)) > rank.GetRankIndex(rank.Rank(winnerRank)) {
			database.UpdateBestRank(winnerIDResult, string(winnerNewRank))
		}
		
		for cardType, count := range cardsPlayedWinner {
			database.UpdateCardUsage(winnerIDResult, cardType, count)
		}
		
		for cardType, count := range cardsPlayedLoser {
			database.UpdateCardUsage(loserID, cardType, count)
		}
		
		leaderboard.InvalidateCache()
	}()
	
	results[winnerIDResult] = &PlayerRankResult{
		PlayerID:   winnerIDResult,
		Username:   r.players[winnerIDResult].Username,
		IsWinner:   true,
		EloChange:  int(winnerNewElo) - winnerElo,
		OldElo:     winnerElo,
		NewElo:     int(winnerNewElo),
		OldRank:    winnerRank,
		NewRank:    string(winnerNewRank),
		RankChange: rankChangeWinner,
		BoundaryBonus: eloResult.WinnerBoundaryBonus,
	}
	
	results[loserID] = &PlayerRankResult{
		PlayerID:   loserID,
		Username:   r.players[loserID].Username,
		IsWinner:   false,
		EloChange:  int(loserNewElo) - loserElo,
		OldElo:     loserElo,
		NewElo:     int(loserNewElo),
		OldRank:    loserRank,
		NewRank:    string(loserNewRank),
		RankChange: rankChangeLoser,
		BoundaryBonus: eloResult.LoserBoundaryBonus,
	}
	
	return results
}

func countDestroyedNodes(player *game.Player) int {
	if player == nil {
		return 0
	}
	count := 0
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			node := player.Grid[x][y]
			if node != nil && !node.IsAlive() && node.Type != game.NodeTypeCore {
				count++
			}
		}
	}
	return count
}

func countCardsPlayed(player *game.Player) int {
	if player == nil {
		return 0
	}
	return len(player.PlayedCards)
}

func getCoreHp(player *game.Player) int {
	if player == nil || player.CoreNode == nil {
		return 0
	}
	return player.CoreNode.HP
}

func getPlayedCardTypes(player *game.Player) map[string]int {
	result := make(map[string]int)
	if player == nil {
		return result
	}
	for _, played := range player.PlayedCards {
		if played.Card != nil {
			result[string(played.Card.Type)]++
		}
	}
	return result
}
