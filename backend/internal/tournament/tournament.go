package tournament

import (
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"cyberhack/internal/database"
	"cyberhack/internal/rank"
	"cyberhack/internal/room"
	"github.com/google/uuid"
)

type TournamentManager struct {
	tournaments  map[string]*TournamentState
	rooms        map[string]*TournamentRoomInfo
	broadcastFn  func(tournamentID, msgType string, payload interface{})
	sendToPlayerFn func(playerID, msgType string, payload interface{})
	mu           sync.RWMutex
	ticker       *time.Ticker
	stopChan     chan struct{}
}

type TournamentState struct {
	Tournament *database.Tournament
	Players    []*database.TournamentPlayer
	Matches    []*database.TournamentMatch
}

type TournamentRoomInfo struct {
	RoomID       string
	MatchID      string
	TournamentID string
	Player1ID    string
	Player2ID    string
	StartedAt    time.Time
}

var manager *TournamentManager
var once sync.Once

func GetManager() *TournamentManager {
	once.Do(func() {
		manager = &TournamentManager{
			tournaments: make(map[string]*TournamentState),
			rooms:       make(map[string]*TournamentRoomInfo),
			stopChan:    make(chan struct{}),
		}
	})
	return manager
}

func (tm *TournamentManager) SetBroadcastFunc(fn func(tournamentID, msgType string, payload interface{})) {
	tm.broadcastFn = fn
}

func (tm *TournamentManager) SetSendToPlayerFunc(fn func(playerID, msgType string, payload interface{})) {
	tm.sendToPlayerFn = fn
}

func (tm *TournamentManager) Start() {
	tm.ticker = time.NewTicker(1 * time.Second)
	go tm.tick()
	log.Println("Tournament manager started")
}

func (tm *TournamentManager) Stop() {
	if tm.ticker != nil {
		tm.ticker.Stop()
	}
	close(tm.stopChan)
	log.Println("Tournament manager stopped")
}

func (tm *TournamentManager) tick() {
	for range tm.ticker.C {
		tm.checkRegistrationDeadlines()
		tm.checkMatchTimeouts()
	}
}

func (tm *TournamentManager) checkRegistrationDeadlines() {
	tm.mu.RLock()
	tournamentIDs := make([]string, 0, len(tm.tournaments))
	for id, state := range tm.tournaments {
		if state.Tournament.Status == "registering" {
			tournamentIDs = append(tournamentIDs, id)
		}
	}
	tm.mu.RUnlock()

	for _, id := range tournamentIDs {
		tm.mu.RLock()
		state, exists := tm.tournaments[id]
		tm.mu.RUnlock()
		if !exists {
			continue
		}

		if time.Now().After(state.Tournament.RegistrationDeadline) {
			go tm.startTournament(id)
		}
	}
}

func (tm *TournamentManager) checkMatchTimeouts() {
	tm.mu.RLock()
	roomInfos := make([]*TournamentRoomInfo, 0, len(tm.rooms))
	for _, info := range tm.rooms {
		roomInfos = append(roomInfos, info)
	}
	tm.mu.RUnlock()

	for _, info := range roomInfos {
		if time.Since(info.StartedAt) > 60*time.Second {
			go tm.handleMatchTimeout(info)
		}
	}
}

func (tm *TournamentManager) handleMatchTimeout(info *TournamentRoomInfo) {
	tm.mu.Lock()
	_, exists := tm.rooms[info.RoomID]
	tm.mu.Unlock()
	if !exists {
		return
	}

	winnerID := info.Player1ID
	if tm.isPlayerConnected(info.Player2ID) {
		winnerID = info.Player2ID
	} else if !tm.isPlayerConnected(info.Player1ID) {
		winnerID = info.Player1ID
	}

	tm.HandleMatchResult(info.MatchID, winnerID)
}

func (tm *TournamentManager) isPlayerConnected(playerID string) bool {
	return true
}

func (tm *TournamentManager) CreateTournament(creatorID, creatorName, name string, maxPlayers int, minRank string, durationMinutes int) (*database.Tournament, error) {
	id := uuid.New().String()
	totalRounds := int(math.Ceil(math.Log2(float64(maxPlayers))))

	t := &database.Tournament{
		ID:                   id,
		Name:                 name,
		MaxPlayers:           maxPlayers,
		MinRank:              minRank,
		CreatorID:            creatorID,
		Status:               "registering",
		RegistrationDeadline: time.Now().Add(time.Duration(durationMinutes) * time.Minute),
		TotalRounds:          totalRounds,
	}

	err := database.CreateTournament(t)
	if err != nil {
		return nil, err
	}

	state := &TournamentState{
		Tournament: t,
		Players:    make([]*database.TournamentPlayer, 0),
		Matches:    make([]*database.TournamentMatch, 0),
	}

	tm.mu.Lock()
	tm.tournaments[id] = state
	tm.mu.Unlock()

	tm.broadcastTournamentListUpdate()

	return t, nil
}

func (tm *TournamentManager) RegisterPlayer(tournamentID, playerID, username string) error {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		var err error
		state, err = tm.loadTournamentState(tournamentID)
		if err != nil {
			return err
		}
		tm.mu.Lock()
		tm.tournaments[tournamentID] = state
		tm.mu.Unlock()
	}

	if state.Tournament.Status != "registering" {
		return &TournamentError{Message: "锦标赛报名已结束"}
	}

	playerCount := len(state.Players)
	if playerCount >= state.Tournament.MaxPlayers {
		return &TournamentError{Message: "锦标赛已满员"}
	}

	player, err := database.GetPlayerByID(playerID)
	if err != nil {
		player = &database.Player{
			ID:          playerID,
			Username:    username,
			EloRating:   1200,
			CurrentRank: "bronze",
		}
	}

	if !tm.checkRankRequirement(player.CurrentRank, state.Tournament.MinRank) {
		return &TournamentError{Message: "段位不满足要求"}
	}

	isRegistered, _ := database.IsPlayerInTournament(tournamentID, playerID)
	if isRegistered {
		return &TournamentError{Message: "你已经报名了该锦标赛"}
	}

	tp := &database.TournamentPlayer{
		TournamentID: tournamentID,
		PlayerID:     playerID,
		Username:     username,
		EloRating:    player.EloRating,
		CurrentRank:  player.CurrentRank,
		Seed:         playerCount + 1,
	}

	err = database.AddTournamentPlayer(tp)
	if err != nil {
		return err
	}

	tm.mu.Lock()
	state.Players = append(state.Players, tp)
	tm.mu.Unlock()

	tm.broadcastTournamentUpdate(tournamentID)

	database.AddTournamentChat(tournamentID, playerID, username, "加入了锦标赛", true)
	tm.broadcastChatUpdate(tournamentID)

	if len(state.Players) >= state.Tournament.MaxPlayers {
		go tm.startTournament(tournamentID)
	}

	return nil
}

func (tm *TournamentManager) RemovePlayer(tournamentID, playerID string) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	if exists {
		newPlayers := make([]*database.TournamentPlayer, 0)
		for _, p := range state.Players {
			if p.PlayerID != playerID {
				newPlayers = append(newPlayers, p)
			}
		}
		state.Players = newPlayers
	}
	tm.mu.Unlock()

	if exists {
		tm.broadcastTournamentUpdate(tournamentID)
		tm.broadcastChatUpdate(tournamentID)
	}
}

func (tm *TournamentManager) KickPlayer(tournamentID, operatorID, targetPlayerID string) error {
	tm.mu.RLock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.RUnlock()

	if !exists {
		return &TournamentError{Message: "锦标赛不存在"}
	}

	if state.Tournament.Status != "registering" {
		return &TournamentError{Message: "锦标赛已开始，无法踢人"}
	}

	if state.Tournament.CreatorID != operatorID {
		return &TournamentError{Message: "只有锦标赛创建者可以踢人"}
	}

	if operatorID == targetPlayerID {
		return &TournamentError{Message: "不能踢出自己"}
	}

	var targetPlayer *database.TournamentPlayer
	for _, p := range state.Players {
		if p.PlayerID == targetPlayerID {
			targetPlayer = p
			break
		}
	}

	if targetPlayer == nil {
		return &TournamentError{Message: "该玩家不在锦标赛中"}
	}

	database.RemoveTournamentPlayer(tournamentID, targetPlayerID)

	tm.mu.Lock()
	newPlayers := make([]*database.TournamentPlayer, 0)
	for _, p := range state.Players {
		if p.PlayerID != targetPlayerID {
			newPlayers = append(newPlayers, p)
		}
	}
	state.Players = newPlayers
	tm.mu.Unlock()

	tm.broadcastTournamentUpdate(tournamentID)

	database.AddTournamentChat(tournamentID, operatorID, targetPlayer.Username, "被踢出了锦标赛", true)
	tm.broadcastChatUpdate(tournamentID)

	if tm.sendToPlayerFn != nil {
		tm.sendToPlayerFn(targetPlayerID, "tournament_kicked", map[string]interface{}{
			"tournamentId":   tournamentID,
			"tournamentName": state.Tournament.Name,
		})
	}

	return nil
}

func (tm *TournamentManager) checkRankRequirement(playerRank, minRank string) bool {
	if minRank == "none" || minRank == "" {
		return true
	}

	playerIdx := rank.GetRankIndex(rank.Rank(playerRank))
	minIdx := rank.GetRankIndex(rank.Rank(minRank))

	return playerIdx >= minIdx
}

func (tm *TournamentManager) startTournament(tournamentID string) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists || state.Tournament.Status != "registering" {
		return
	}

	playerCount := len(state.Players)
	if playerCount < 4 {
		tm.cancelTournament(tournamentID, "报名人数不足4人，锦标赛取消")
		return
	}

	tm.generateSeeds(tournamentID)
	tm.generateBracket(tournamentID)

	database.StartTournament(tournamentID)
	state.Tournament.Status = "in_progress"
	state.Tournament.CurrentRound = 1
	state.Tournament.StartedAt = time.Now()

	tm.broadcastTournamentUpdate(tournamentID)
	tm.broadcastBracketUpdate(tournamentID)

	database.AddTournamentChat(tournamentID, "", "系统", "锦标赛正式开始！", true)
	tm.broadcastChatUpdate(tournamentID)

	tm.startRound(tournamentID, 1)
}

func (tm *TournamentManager) cancelTournament(tournamentID, reason string) {
	database.UpdateTournamentStatus(tournamentID, "cancelled")

	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	if exists {
		state.Tournament.Status = "cancelled"
	}
	tm.mu.Unlock()

	tm.broadcastTournamentUpdate(tournamentID)

	if exists {
		for _, player := range state.Players {
			if tm.sendToPlayerFn != nil {
				tm.sendToPlayerFn(player.PlayerID, "tournament_cancelled", map[string]interface{}{
					"tournamentId": tournamentID,
					"tournamentName": state.Tournament.Name,
					"reason": reason,
				})
			}
		}
	}

	database.AddTournamentChat(tournamentID, "", "系统", reason, true)
	tm.broadcastChatUpdate(tournamentID)
}

func (tm *TournamentManager) generateSeeds(tournamentID string) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		return
	}

	sort.Slice(state.Players, func(i, j int) bool {
		return state.Players[i].EloRating > state.Players[j].EloRating
	})

	for i, player := range state.Players {
		player.Seed = i + 1
		database.UpdatePlayerSeed(tournamentID, player.PlayerID, i+1)
	}
}

func (tm *TournamentManager) generateBracket(tournamentID string) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		return
	}

	players := state.Players
	numPlayers := len(players)
	nextPowerOf2 := int(math.Pow(2, math.Ceil(math.Log2(float64(numPlayers)))))

	numFirstRoundMatches := nextPowerOf2 / 2
	firstRoundPlayers := make([]*database.TournamentPlayer, nextPowerOf2)

	for i := 0; i < numPlayers; i++ {
		firstRoundPlayers[i] = players[i]
	}

	tm.snakeSeed(firstRoundPlayers, numPlayers)

	round := 1
	matchIndex := 0
	for i := 0; i < numFirstRoundMatches; i++ {
		p1 := firstRoundPlayers[i*2]
		p2 := firstRoundPlayers[i*2+1]

		matchID := uuid.New().String()
		match := &database.TournamentMatch{
			ID:          matchID,
			TournamentID: tournamentID,
			RoundNumber: round,
			MatchIndex:  matchIndex,
			Status:      "pending",
		}

		if p1 != nil {
			match.Player1ID = p1.PlayerID
			match.Player1Name = p1.Username
		}
		if p2 != nil {
			match.Player2ID = p2.PlayerID
			match.Player2Name = p2.Username
		}

		if p1 != nil && p2 == nil {
			match.Status = "finished"
			match.WinnerID = p1.PlayerID
		} else if p1 == nil && p2 != nil {
			match.Status = "finished"
			match.WinnerID = p2.PlayerID
		}

		database.CreateTournamentMatch(match)
		state.Matches = append(state.Matches, match)
		matchIndex++
	}

	totalRounds := int(math.Log2(float64(nextPowerOf2)))
	state.Tournament.TotalRounds = totalRounds
}

func (tm *TournamentManager) snakeSeed(players []*database.TournamentPlayer, numPlayers int) {
	positions := make([]int, len(players))
	for i := range positions {
		positions[i] = i
	}

	result := make([]*database.TournamentPlayer, len(players))
	playerIdx := 0

	roundSize := len(players)
	for roundSize > 1 {
		half := roundSize / 2
		top := make([]int, half)
		bottom := make([]int, half)

		for i := 0; i < half; i++ {
			top[i] = positions[i*2]
			bottom[half-1-i] = positions[i*2+1]
		}

		newPositions := make([]int, 0)
		newPositions = append(newPositions, top...)
		newPositions = append(newPositions, bottom...)
		positions = newPositions
		roundSize = half
	}

	for _, pos := range positions {
		if playerIdx < numPlayers {
			result[pos] = players[playerIdx]
			playerIdx++
		}
	}

	for i := range players {
		players[i] = result[i]
	}
}

func (tm *TournamentManager) startRound(tournamentID string, roundNumber int) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		return
	}

	roundMatches := make([]*database.TournamentMatch, 0)
	for _, m := range state.Matches {
		if m.RoundNumber == roundNumber {
			roundMatches = append(roundMatches, m)
		}
	}

	allReady := true
	for _, m := range roundMatches {
		if m.Player1ID == "" || m.Player2ID == "" {
			allReady = false
			break
		}
	}

	if !allReady {
		return
	}

	database.UpdateTournamentRound(tournamentID, roundNumber)
	state.Tournament.CurrentRound = roundNumber

	tm.broadcastTournamentUpdate(tournamentID)

	for _, match := range roundMatches {
		if match.Status == "pending" {
			go tm.startMatch(tournamentID, match.ID)
		}
	}
}

func (tm *TournamentManager) startMatch(tournamentID, matchID string) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		return
	}

	var match *database.TournamentMatch
	for _, m := range state.Matches {
		if m.ID == matchID {
			match = m
			break
		}
	}

	if match == nil || match.Status != "pending" {
		return
	}

	r := room.NewRoom("", 2)
	
	gameOverHandled := false
	var gameOverMu sync.Mutex
	
	r.SetBroadcastFunc(func(playerID, msgType string, payload interface{}) {
		if msgType == "game_over" {
			gameOverMu.Lock()
			if !gameOverHandled {
				gameOverHandled = true
				gameOverMu.Unlock()
				go func() {
					if payloadMap, ok := payload.(map[string]interface{}); ok {
						if winnerID, ok := payloadMap["winnerId"].(string); ok {
							tm.HandleMatchResult(matchID, winnerID)
						}
					}
				}()
			} else {
				gameOverMu.Unlock()
			}
		}
		if msgType == "phase_change" {
			if payloadMap, ok := payload.(map[string]interface{}); ok {
				if turn, ok := payloadMap["turn"].(int); ok {
					tm.broadcastMatchProgress(tournamentID, matchID, turn, r.Game.Config.MaxTurns)
				}
			}
		}
		if tm.sendToPlayerFn != nil {
			tm.sendToPlayerFn(playerID, msgType, payload)
		}
	})

	r.AddPlayer(match.Player1ID, match.Player1Name)
	r.AddPlayer(match.Player2ID, match.Player2Name)

	database.UpdateTournamentMatchRoom(matchID, r.ID)
	match.RoomID = r.ID
	match.Status = "in_progress"
	match.StartedAt = time.Now()

	tm.mu.Lock()
	tm.rooms[r.ID] = &TournamentRoomInfo{
		RoomID:       r.ID,
		MatchID:      matchID,
		TournamentID: tournamentID,
		Player1ID:    match.Player1ID,
		Player2ID:    match.Player2ID,
		StartedAt:    time.Now(),
	}
	tm.mu.Unlock()

	r.Broadcast("tournament_match_start", map[string]interface{}{
		"tournamentId": tournamentID,
		"matchId":      matchID,
		"roomId":       r.ID,
		"round":        match.RoundNumber,
		"opponents": []map[string]string{
			{"id": match.Player1ID, "username": match.Player1Name},
			{"id": match.Player2ID, "username": match.Player2Name},
		},
	})

	r.ForceStartGame()

	tm.broadcastBracketUpdate(tournamentID)
}

func (tm *TournamentManager) HandleMatchResult(matchID, winnerID string) {
	match, err := database.GetTournamentMatchByID(matchID)
	if err != nil {
		return
	}

	if match.Status == "finished" {
		return
	}

	database.UpdateTournamentMatchResult(matchID, winnerID)
	match.Status = "finished"
	match.WinnerID = winnerID
	match.EndedAt = time.Now()

	tm.mu.Lock()
	state, exists := tm.tournaments[match.TournamentID]
	if exists {
		for _, m := range state.Matches {
			if m.ID == matchID {
				m.Status = "finished"
				m.WinnerID = winnerID
				m.EndedAt = time.Now()
				break
			}
		}
	}
	delete(tm.rooms, match.RoomID)
	tm.mu.Unlock()

	if !exists {
		return
	}

	tm.broadcastBracketUpdate(match.TournamentID)

	tm.advanceWinner(match.TournamentID, match.RoundNumber, match.MatchIndex, winnerID)

	tm.checkRoundComplete(match.TournamentID, match.RoundNumber)
}

func (tm *TournamentManager) advanceWinner(tournamentID string, currentRound, matchIndex int, winnerID string) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		return
	}

	nextRound := currentRound + 1
	if nextRound > state.Tournament.TotalRounds {
		return
	}

	nextMatchIndex := matchIndex / 2
	isFirstPlayer := matchIndex%2 == 0

	var winnerPlayer *database.TournamentPlayer
	for _, p := range state.Players {
		if p.PlayerID == winnerID {
			winnerPlayer = p
			break
		}
	}

	if winnerPlayer == nil {
		return
	}

	var nextMatch *database.TournamentMatch
	for _, m := range state.Matches {
		if m.RoundNumber == nextRound && m.MatchIndex == nextMatchIndex {
			nextMatch = m
			break
		}
	}

	if nextMatch == nil {
		matchID := uuid.New().String()
		nextMatch = &database.TournamentMatch{
			ID:          matchID,
			TournamentID: tournamentID,
			RoundNumber: nextRound,
			MatchIndex:  nextMatchIndex,
			Status:      "pending",
		}
		database.CreateTournamentMatch(nextMatch)
		state.Matches = append(state.Matches, nextMatch)
	}

	if isFirstPlayer {
		nextMatch.Player1ID = winnerPlayer.PlayerID
		nextMatch.Player1Name = winnerPlayer.Username
	} else {
		nextMatch.Player2ID = winnerPlayer.PlayerID
		nextMatch.Player2Name = winnerPlayer.Username
	}
}

func (tm *TournamentManager) checkRoundComplete(tournamentID string, roundNumber int) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		return
	}

	roundMatches := make([]*database.TournamentMatch, 0)
	for _, m := range state.Matches {
		if m.RoundNumber == roundNumber {
			roundMatches = append(roundMatches, m)
		}
	}

	allFinished := true
	for _, m := range roundMatches {
		if m.Status != "finished" {
			allFinished = false
			break
		}
	}

	if !allFinished {
		return
	}

	if roundNumber >= state.Tournament.TotalRounds {
		tm.finishTournament(tournamentID)
		return
	}

	time.AfterFunc(3*time.Second, func() {
		tm.startRound(tournamentID, roundNumber+1)
	})
}

func (tm *TournamentManager) finishTournament(tournamentID string) {
	tm.mu.Lock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.Unlock()

	if !exists {
		return
	}

	var winnerID string
	for _, m := range state.Matches {
		if m.RoundNumber == state.Tournament.TotalRounds {
			winnerID = m.WinnerID
			break
		}
	}

	database.EndTournament(tournamentID, winnerID)
	state.Tournament.Status = "finished"
	state.Tournament.WinnerID = winnerID
	state.Tournament.EndedAt = time.Now()

	tm.calculateAndAwardPrizes(tournamentID)

	tm.broadcastTournamentUpdate(tournamentID)
	tm.broadcastBracketUpdate(tournamentID)

	var winnerName string
	for _, p := range state.Players {
		if p.PlayerID == winnerID {
			winnerName = p.Username
			break
		}
	}

	database.AddTournamentChat(tournamentID, "", "系统", 
		"锦标赛结束！冠军是："+winnerName, true)
	tm.broadcastChatUpdate(tournamentID)
}

func (tm *TournamentManager) calculateAndAwardPrizes(tournamentID string) {
	tm.mu.RLock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.RUnlock()

	if !exists {
		return
	}

	numPlayers := len(state.Players)

	type playerStats struct {
		playerID string
		wins     int
		losses   int
		position int
	}

	statsMap := make(map[string]*playerStats)
	for _, p := range state.Players {
		statsMap[p.PlayerID] = &playerStats{
			playerID: p.PlayerID,
			wins:     0,
			losses:   0,
			position: 0,
		}
	}

	for _, m := range state.Matches {
		if m.Status != "finished" {
			continue
		}
		if m.Player1ID != "" {
			if m.WinnerID == m.Player1ID {
				statsMap[m.Player1ID].wins++
			} else {
				statsMap[m.Player1ID].losses++
			}
		}
		if m.Player2ID != "" {
			if m.WinnerID == m.Player2ID {
				statsMap[m.Player2ID].wins++
			} else {
				statsMap[m.Player2ID].losses++
			}
		}
	}

	playersByRound := make([][]string, state.Tournament.TotalRounds+1)
	for round := 1; round <= state.Tournament.TotalRounds; round++ {
		for _, m := range state.Matches {
			if m.RoundNumber == round && m.Status == "finished" {
				if m.WinnerID != m.Player1ID && m.Player1ID != "" {
					playersByRound[round] = append(playersByRound[round], m.Player1ID)
				}
				if m.WinnerID != m.Player2ID && m.Player2ID != "" {
					playersByRound[round] = append(playersByRound[round], m.Player2ID)
				}
			}
		}
	}

	var championID string
	for _, m := range state.Matches {
		if m.RoundNumber == state.Tournament.TotalRounds && m.Status == "finished" {
			championID = m.WinnerID
			break
		}
	}

	position := 1
	for round := state.Tournament.TotalRounds; round >= 1; round-- {
		if round == state.Tournament.TotalRounds {
			if statsMap[championID] != nil {
				statsMap[championID].position = 1
			}
			position = 2
		}

		eliminated := playersByRound[round]
		eliminatedCount := len(eliminated)
		if eliminatedCount > 0 {
			for _, pid := range eliminated {
				if statsMap[pid] != nil {
					statsMap[pid].position = position
				}
			}
			position += eliminatedCount
		}
	}

	for pid, stats := range statsMap {
		var player *database.TournamentPlayer
		for _, p := range state.Players {
			if p.PlayerID == pid {
				player = p
				break
			}
		}
		if player == nil {
			continue
		}

		eloBonus := 0
		hasTop4Badge := false

		if stats.position == 1 {
			eloBonus = numPlayers * 5
		} else if stats.position == 2 {
			eloBonus = numPlayers * 2
		}

		if stats.position <= 4 {
			hasTop4Badge = true
		}

		database.UpdatePlayerFinalPosition(tournamentID, pid, stats.position)

		record := &database.TournamentRecord{
			TournamentID:   tournamentID,
			PlayerID:       pid,
			TournamentName: state.Tournament.Name,
			FinalPosition:  stats.position,
			TotalMatches:   stats.wins + stats.losses,
			Wins:           stats.wins,
			Losses:         stats.losses,
			EloBonus:       eloBonus,
			HasTop4Badge:   hasTop4Badge,
		}
		database.CreateTournamentRecord(record)

		if eloBonus > 0 {
			p, err := database.GetPlayerByID(pid)
			if err == nil {
				newElo := p.EloRating + eloBonus
				newRank := rank.GetRank(newElo)
				database.UpdatePlayerElo(pid, newElo, string(newRank), p.RankProtectionGames)
			}
		}
	}
}

func (tm *TournamentManager) GetTournament(tournamentID string) (*TournamentState, error) {
	tm.mu.RLock()
	state, exists := tm.tournaments[tournamentID]
	tm.mu.RUnlock()

	if exists {
		return state, nil
	}

	return tm.loadTournamentState(tournamentID)
}

func (tm *TournamentManager) loadTournamentState(tournamentID string) (*TournamentState, error) {
	t, err := database.GetTournamentByID(tournamentID)
	if err != nil {
		return nil, err
	}

	players, err := database.GetTournamentPlayers(tournamentID)
	if err != nil {
		players = make([]*database.TournamentPlayer, 0)
	}

	matches, err := database.GetTournamentMatches(tournamentID)
	if err != nil {
		matches = make([]*database.TournamentMatch, 0)
	}

	state := &TournamentState{
		Tournament: t,
		Players:    players,
		Matches:    matches,
	}

	return state, nil
}

func (tm *TournamentManager) GetActiveTournaments() []*database.Tournament {
	tm.mu.RLock()
	tournaments := make([]*database.Tournament, 0, len(tm.tournaments))
	for _, state := range tm.tournaments {
		if state.Tournament.Status == "registering" || state.Tournament.Status == "in_progress" {
			tournaments = append(tournaments, state.Tournament)
		}
	}
	tm.mu.RUnlock()

	if len(tournaments) == 0 {
		dbTournaments, err := database.GetActiveTournaments()
		if err == nil {
			tournaments = dbTournaments
		}
	}

	return tournaments
}

func (tm *TournamentManager) GetTournamentWithPlayerCount(tournamentID string) (map[string]interface{}, error) {
	state, err := tm.GetTournament(tournamentID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"tournament": state.Tournament,
		"playerCount": len(state.Players),
		"players":     state.Players,
	}, nil
}

func (tm *TournamentManager) GetBracket(tournamentID string) ([]*database.TournamentMatch, error) {
	state, err := tm.GetTournament(tournamentID)
	if err != nil {
		return nil, err
	}
	return state.Matches, nil
}

func (tm *TournamentManager) GetChatMessages(tournamentID string, limit int) ([]*database.TournamentChatMessage, error) {
	return database.GetTournamentChat(tournamentID, limit)
}

func (tm *TournamentManager) SendChatMessage(tournamentID, playerID, username, message string) error {
	err := database.AddTournamentChat(tournamentID, playerID, username, message, false)
	if err != nil {
		return err
	}
	tm.broadcastChatUpdate(tournamentID)
	return nil
}

func (tm *TournamentManager) broadcastTournamentListUpdate() {
	tournaments := tm.GetActiveTournaments()
	tournamentList := make([]map[string]interface{}, 0)
	
	for _, t := range tournaments {
		playerCount, _ := database.GetTournamentPlayerCount(t.ID)
		tournamentList = append(tournamentList, map[string]interface{}{
			"id":              t.ID,
			"name":            t.Name,
			"maxPlayers":      t.MaxPlayers,
			"minRank":         t.MinRank,
			"status":          t.Status,
			"registrationDeadline": t.RegistrationDeadline,
			"currentRound":    t.CurrentRound,
			"totalRounds":     t.TotalRounds,
			"playerCount":     playerCount,
		})
	}

	if tm.broadcastFn != nil {
		tm.broadcastFn("", "tournament_list_update", map[string]interface{}{
			"tournaments": tournamentList,
		})
	}
}

func (tm *TournamentManager) broadcastTournamentUpdate(tournamentID string) {
	if tm.broadcastFn != nil {
		state, _ := tm.GetTournament(tournamentID)
		if state != nil {
			tm.broadcastFn(tournamentID, "tournament_update", map[string]interface{}{
				"tournament":  state.Tournament,
				"playerCount": len(state.Players),
				"players":     state.Players,
			})
		}
	}
	tm.broadcastTournamentListUpdate()
}

func (tm *TournamentManager) broadcastBracketUpdate(tournamentID string) {
	if tm.broadcastFn != nil {
		matches, _ := tm.GetBracket(tournamentID)
		tm.broadcastFn(tournamentID, "bracket_update", map[string]interface{}{
			"matches": matches,
		})
	}
}

func (tm *TournamentManager) broadcastChatUpdate(tournamentID string) {
	if tm.broadcastFn != nil {
		messages, _ := tm.GetChatMessages(tournamentID, 50)
		tm.broadcastFn(tournamentID, "chat_update", map[string]interface{}{
			"messages": messages,
		})
	}
}

func (tm *TournamentManager) broadcastMatchProgress(tournamentID, matchID string, currentTurn, maxTurns int) {
	if tm.broadcastFn != nil {
		tm.broadcastFn(tournamentID, "match_progress", map[string]interface{}{
			"matchId":     matchID,
			"currentTurn": currentTurn,
			"maxTurns":    maxTurns,
		})
	}
}

type TournamentError struct {
	Message string
}

func (e *TournamentError) Error() string {
	return e.Message
}
