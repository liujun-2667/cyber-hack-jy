package game

import (
	"sort"
	"math/rand"
	"strconv"
)

type GamePhase string

const (
	PhaseSetup      GamePhase = "setup"
	PhasePlacement  GamePhase = "placement"
	PhaseProgramming GamePhase = "programming"
	PhaseExecution  GamePhase = "execution"
	PhaseGameOver   GamePhase = "gameover"
)

type GameConfig struct {
	MaxPlayers      int        `json:"maxPlayers"`
	MaxTurns        int        `json:"maxTurns"`
	InitialHandSize int        `json:"initialHandSize"`
	ProgrammingTime int        `json:"programmingTime"`
	BannedCards     []CardType `json:"bannedCards"`
	GameMode        string     `json:"gameMode"`
}

type Game struct {
	ID                string            `json:"id"`
	Config            *GameConfig       `json:"config"`
	Players           map[string]*Player `json:"players"`
	PlayerOrder       []string          `json:"playerOrder"`
	CurrentTurn       int               `json:"currentTurn"`
	Phase             GamePhase         `json:"phase"`
	CurrentPlayerIndex int              `json:"-"`
	WinnerID          string            `json:"winnerId"`
	GameLog           []string          `json:"gameLog"`
	TurnActions       []*TurnAction     `json:"turnActions"`
	ReplayTurns       []*ReplayTurn     `json:"-"`
	lastReplayLogIndex int              `json:"-"`
}

type TurnAction struct {
	PlayerID       string  `json:"playerId"`
	Card           *Card   `json:"card"`
	Target         string  `json:"target"`
	TargetPlayerID string  `json:"targetPlayerId"`
	Result         string  `json:"result"`
	Damage         int     `json:"damage"`
}

type ReplayTurn struct {
	TurnNumber int                  `json:"turnNumber"`
	Actions    []*ReplayAction      `json:"actions"`
	NodeStates map[string]*NodeStateSnapshot `json:"nodeStates"`
	Log        []string             `json:"log"`
}

type ReplayAction struct {
	PlayerID       string `json:"playerId"`
	Username       string `json:"username"`
	Card           *Card  `json:"card"`
	TargetNodeID   string `json:"targetNodeId"`
	TargetPlayerID string `json:"targetPlayerId"`
	Result         string `json:"result"`
	Damage         int    `json:"damage"`
}

type NodeStateSnapshot struct {
	ID        string      `json:"id"`
	Type      NodeType    `json:"type"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	HP        int         `json:"hp"`
	MaxHP     int         `json:"maxHp"`
	Defense   int         `json:"defense"`
	Bandwidth int         `json:"bandwidth"`
	IsAlive   bool        `json:"isAlive"`
	Status    NodeStatus  `json:"status"`
	OwnerID   string      `json:"ownerId"`
}

func DefaultGameConfig() *GameConfig {
	return &GameConfig{
		MaxPlayers:      2,
		MaxTurns:        30,
		InitialHandSize: 5,
		ProgrammingTime: 20,
		BannedCards:     []CardType{},
		GameMode:        "quick",
	}
}

func NewGame(id string, config *GameConfig) *Game {
	if config == nil {
		config = DefaultGameConfig()
	}
	return &Game{
		ID:            id,
		Config:        config,
		Players:       make(map[string]*Player),
		PlayerOrder:   make([]string, 0),
		CurrentTurn:   0,
		Phase:         PhaseSetup,
		GameLog:       make([]string, 0),
		TurnActions:   make([]*TurnAction, 0),
		ReplayTurns:   make([]*ReplayTurn, 0),
	}
}

func (g *Game) recordTurnSnapshot() {
	nodeStates := make(map[string]*NodeStateSnapshot)
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		if player == nil {
			continue
		}
		for x := 0; x < 5; x++ {
			for y := 0; y < 5; y++ {
				node := player.Grid[x][y]
				if node != nil {
					snapshot := &NodeStateSnapshot{
						ID:        node.ID,
						Type:      node.Type,
						X:         node.X,
						Y:         node.Y,
						HP:        node.HP,
						MaxHP:     node.MaxHP,
						Defense:   node.Defense,
						Bandwidth: node.Bandwidth,
						IsAlive:   node.IsAlive(),
						Status:    node.Status,
						OwnerID:   node.OwnerID,
					}
					nodeStates[node.ID] = snapshot
				}
			}
		}
	}

	actions := make([]*ReplayAction, 0, len(g.TurnActions))
	for _, action := range g.TurnActions {
		player := g.Players[action.PlayerID]
		username := ""
		if player != nil {
			username = player.Username
		}
		replayAction := &ReplayAction{
			PlayerID:       action.PlayerID,
			Username:       username,
			Card:           action.Card,
			TargetNodeID:   action.Target,
			TargetPlayerID: action.TargetPlayerID,
			Result:         action.Result,
			Damage:         action.Damage,
		}
		actions = append(actions, replayAction)
	}

	currentLogLen := len(g.GameLog)
	turnLog := make([]string, 0)
	if g.lastReplayLogIndex < currentLogLen {
		turnLog = g.GameLog[g.lastReplayLogIndex:currentLogLen]
	}
	g.lastReplayLogIndex = currentLogLen

	turn := &ReplayTurn{
		TurnNumber: g.CurrentTurn,
		Actions:    actions,
		NodeStates: nodeStates,
		Log:        turnLog,
	}
	g.ReplayTurns = append(g.ReplayTurns, turn)
}

func (g *Game) AddPlayer(playerID, username string) bool {
	if len(g.Players) >= g.Config.MaxPlayers {
		return false
	}
	if g.Phase != PhaseSetup {
		return false
	}
	player := NewPlayer(playerID, username)
	g.Players[playerID] = player
	g.PlayerOrder = append(g.PlayerOrder, playerID)
	g.addLog("玩家 " + username + " 加入游戏")
	return true
}

func (g *Game) StartGame() bool {
	if len(g.Players) < 2 {
		return false
	}
	g.Phase = PhasePlacement
	g.addLog("游戏开始！进入布阵阶段")
	
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		player.InitializeGrid()
	}
	
	return true
}

func (g *Game) PlaceNode(playerID string, nodeType NodeType, x, y int) bool {
	if g.Phase != PhasePlacement {
		return false
	}
	player := g.Players[playerID]
	if player == nil {
		return false
	}
	return player.PlaceNode(nodeType, x, y)
}

func (g *Game) StartFirstTurn() {
	g.lastReplayLogIndex = len(g.GameLog)
	g.CurrentTurn = 1
	g.Phase = PhaseProgramming
	
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		player.DrawCards(g.Config.InitialHandSize, playerID+"-init")
	}
	
	g.addLog("第 1 回合开始 - 编程阶段")
}

func (g *Game) StartProgrammingPhase() {
	g.Phase = PhaseProgramming
	g.TurnActions = make([]*TurnAction, 0)
	
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		if player.IsAlive {
			player.PlayedCards = make([]*PlayedCard, 0)
			player.UpdateCooldowns()
			g.resetTurnStatus(player)
			player.DrawCards(2, playerID+"-turn"+strconv.Itoa(g.CurrentTurn))
		}
	}
	
	g.addLog("第 " + strconv.Itoa(g.CurrentTurn) + " 回合 - 编程阶段")
}

func (g *Game) resetTurnStatus(player *Player) {
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			node := player.Grid[x][y]
			if node != nil {
				node.Status.HasIDS = false
				node.Status.HasDDoS = false
				node.Bandwidth = node.MaxBandwidth
				if node.Status.FirewallTurns > 0 {
					node.Status.FirewallTurns--
				}
			}
		}
	}
}

func (g *Game) PlayCard(playerID, cardID, targetNodeID, targetPlayerID string) bool {
	if g.Phase != PhaseProgramming {
		return false
	}
	player := g.Players[playerID]
	if player == nil || !player.IsAlive {
		return false
	}
	if len(player.PlayedCards) >= 3 {
		return false
	}
	
	card := player.RemoveCard(cardID)
	if card == nil {
		return false
	}
	
	if cooldown, ok := player.Cooldowns[card.Type]; ok && cooldown > 0 {
		player.Hand = append(player.Hand, card)
		return false
	}
	
	played := &PlayedCard{
		Card:           card,
		PlayerID:       playerID,
		TargetNodeID:   targetNodeID,
		TargetPlayerID: targetPlayerID,
	}
	player.PlayedCards = append(player.PlayedCards, played)
	player.Cooldowns[card.Type] = card.Cooldown
	
	return true
}

func (g *Game) ExecutePhase() {
	g.Phase = PhaseExecution
	g.addLog("第 " + strconv.Itoa(g.CurrentTurn) + " 回合 - 执行阶段")
	
	allPlayedCards := make([]*PlayedCard, 0)
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		if player.IsAlive {
			allPlayedCards = append(allPlayedCards, player.PlayedCards...)
		}
	}
	
	sort.Slice(allPlayedCards, func(i, j int) bool {
		return allPlayedCards[i].Card.Speed > allPlayedCards[j].Card.Speed
	})
	
	for _, played := range allPlayedCards {
		if g.Phase == PhaseGameOver {
			break
		}
		g.executeCard(played)
	}
	
	g.checkGameEnd()
	
	g.recordTurnSnapshot()
	
	if g.Phase != PhaseGameOver {
		g.CurrentTurn++
		g.StartProgrammingPhase()
	}
}

func (g *Game) executeCard(played *PlayedCard) {
	player := g.Players[played.PlayerID]
	if player == nil || !player.IsAlive {
		return
	}

	targetPlayer := g.Players[played.TargetPlayerID]
	if targetPlayer == nil {
		targetPlayer = player
	}

	targetNode := targetPlayer.GetNodeByID(played.TargetNodeID)
	
	action := &TurnAction{
		PlayerID:       played.PlayerID,
		Card:           played.Card,
		Target:         played.TargetNodeID,
		TargetPlayerID: targetPlayer.ID,
	}

	switch played.Card.Type {
	case CardPortScan:
		g.executePortScan(player, targetPlayer, targetNode, action)
	case CardBruteForce:
		g.executeBruteForce(player, targetPlayer, targetNode, played.Card, action)
	case CardSQLInjection:
		g.executeSQLInjection(player, targetPlayer, targetNode, played.Card, action)
	case CardDDoS:
		g.executeDDoS(player, targetPlayer, targetNode, action)
	case CardTrojan:
		g.executeTrojan(player, targetPlayer, targetNode, action)
	case CardFirewall:
		g.executeFirewall(player, targetNode, action)
	case CardIDS:
		g.executeIDS(player, targetNode, action)
	case CardEncryption:
		g.executeEncryption(player, targetNode, action)
	case CardHoneypot:
		g.executeHoneypot(player, targetNode, action)
	case CardTrafficClean:
		g.executeTrafficClean(player, targetNode, action)
	case CardBandwidthUpgrade:
		g.executeBandwidthUpgrade(player, targetNode, action)
	case CardNodeRepair:
		g.executeNodeRepair(player, targetNode, action)
	case CardDataTheft:
		g.executeDataTheft(player, targetPlayer, targetNode, action)
	case CardBackdoor:
		g.executeBackdoor(player, targetPlayer, targetNode, action)
	case CardSniff:
		g.executeSniff(player, targetPlayer, targetNode, action)
	}

	g.TurnActions = append(g.TurnActions, action)
}

func (g *Game) executePortScan(attacker, defender *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	target.Status.ScanRevealed = true
	attacker.RevealNode(defender.ID, target.ID)
	action.Result = "扫描成功"
	g.addLog(attacker.Username + " 使用端口扫描揭示了 " + target.ID + " 的信息")
}

func (g *Game) executeBruteForce(attacker, defender *Player, target *Node, card *Card, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	if !g.isTargetReachable(attacker, defender, target) {
		action.Result = "目标不可达"
		return
	}
	
	damage := target.TakeDamage(card.Damage)
	action.Damage = damage
	
	if target.Status.HasIDS {
		counterDamage := 4
		edgeNodes := attacker.GetEdgeNodes()
		if len(edgeNodes) > 0 {
			edgeNodes[0].TakeDamage(counterDamage)
			g.addLog(target.ID + " 的入侵检测系统反击，造成 " + strconv.Itoa(counterDamage) + " 点伤害")
		}
	}
	
	if !target.IsAlive() {
		action.Result = "节点被摧毁"
		attacker.AddCard(DrawRandomCard(attacker.ID+"-reward", ""))
		if target.Type == NodeTypeDatabase {
			attacker.AddCard(DrawRandomCard(attacker.ID+"-rare", "rare"))
		}
		if target.Status.HasHoneypot {
			edgeNodes := attacker.GetEdgeNodes()
			if len(edgeNodes) > 0 {
				honeypotTarget := edgeNodes[rand.Intn(len(edgeNodes))]
				honeypotTarget.TakeDamage(5)
				defender.RevealNode(attacker.ID, honeypotTarget.ID)
				action.Result += "，触发蜜罐"
			}
		}
		g.addLog(attacker.Username + " 暴力破解摧毁了 " + target.ID)
	} else {
		action.Result = "造成 " + strconv.Itoa(damage) + " 点伤害"
		g.addLog(attacker.Username + " 对 " + target.ID + " 造成 " + strconv.Itoa(damage) + " 点伤害")
	}
	
	defender.CheckDefeat()
}

func (g *Game) executeSQLInjection(attacker, defender *Player, target *Node, card *Card, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if target.Type != NodeTypeDatabase {
		action.Result = "只能对数据库使用"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	if !g.isTargetReachable(attacker, defender, target) {
		action.Result = "目标不可达"
		return
	}
	
	damage := card.Damage * 2
	target.HP -= damage
	if target.HP < 0 {
		target.HP = 0
	}
	action.Damage = damage
	
	if !target.IsAlive() {
		action.Result = "数据库被摧毁"
		attacker.AddCard(DrawRandomCard(attacker.ID+"-reward", ""))
		attacker.AddCard(DrawRandomCard(attacker.ID+"-rare", "rare"))
		g.addLog(attacker.Username + " SQL注入摧毁了数据库 " + target.ID)
	} else {
		action.Result = "造成 " + strconv.Itoa(damage) + " 点伤害"
		g.addLog(attacker.Username + " SQL注入对 " + target.ID + " 造成 " + strconv.Itoa(damage) + " 点伤害")
	}
	
	defender.CheckDefeat()
}

func (g *Game) executeDDoS(attacker, defender *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	if !g.isTargetReachable(attacker, defender, target) {
		action.Result = "目标不可达"
		return
	}
	
	target.Status.HasDDoS = true
	target.Bandwidth = 0
	action.Result = "DDoS攻击成功"
	g.addLog(attacker.Username + " 对 " + target.ID + " 发动DDoS攻击")
}

func (g *Game) executeTrojan(attacker, defender *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	if !g.isTargetReachable(attacker, defender, target) {
		action.Result = "目标不可达"
		return
	}
	
	target.Status.HasTrojan = true
	action.Result = "木马植入成功"
	g.addLog(attacker.Username + " 在 " + target.ID + " 植入了木马")
}

func (g *Game) executeFirewall(player *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	
	target.Status.FirewallTurns = 3
	action.Result = "防火墙部署成功"
	g.addLog(player.Username + " 在 " + target.ID + " 部署了防火墙")
}

func (g *Game) executeIDS(player *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	
	target.Status.HasIDS = true
	action.Result = "入侵检测部署成功"
	g.addLog(player.Username + " 在 " + target.ID + " 部署了入侵检测系统")
}

func (g *Game) executeEncryption(player *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	
	target.Defense += 2
	action.Result = "数据加密成功"
	g.addLog(player.Username + " 为 " + target.ID + " 启用了数据加密")
}

func (g *Game) executeHoneypot(player *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	
	target.Status.HasHoneypot = true
	action.Result = "蜜罐布置成功"
	g.addLog(player.Username + " 在 " + target.ID + " 布置了蜜罐")
}

func (g *Game) executeTrafficClean(player *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	
	target.Status.HasDDoS = false
	target.Status.HasTrojan = false
	target.Bandwidth = target.MaxBandwidth
	action.Result = "流量清洗完成"
	g.addLog(player.Username + " 对 " + target.ID + " 进行了流量清洗")
}

func (g *Game) executeBandwidthUpgrade(player *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	
	target.MaxBandwidth += 2
	target.Bandwidth += 2
	action.Result = "带宽升级成功"
	g.addLog(player.Username + " 升级了 " + target.ID + " 的带宽")
}

func (g *Game) executeNodeRepair(player *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	
	target.Heal(10)
	action.Result = "节点修复成功"
	g.addLog(player.Username + " 修复了 " + target.ID + "，恢复 10 点生命")
}

func (g *Game) executeDataTheft(attacker, defender *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if target.Type != NodeTypeDatabase {
		action.Result = "只能对数据库使用"
		return
	}
	if target.IsAlive() {
		action.Result = "数据库尚未被攻破"
		return
	}
	
	card := DrawRandomCard(attacker.ID+"-stolen", "rare")
	if card != nil {
		attacker.AddCard(card)
	}
	action.Result = "窃取了稀有卡牌"
	g.addLog(attacker.Username + " 从已攻破的数据库中窃取了稀有卡牌")
}

func (g *Game) executeBackdoor(attacker, defender *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	if !target.IsAlive() {
		action.Result = "目标已被摧毁"
		return
	}
	if !g.isTargetReachable(attacker, defender, target) {
		action.Result = "目标不可达"
		return
	}
	
	if !contains(target.Programs, "backdoor") {
		target.Programs = append(target.Programs, "backdoor")
	}
	action.Result = "后门安装成功"
	g.addLog(attacker.Username + " 在 " + target.ID + " 安装了后门")
}

func (g *Game) executeSniff(attacker, defender *Player, target *Node, action *TurnAction) {
	if target == nil {
		action.Result = "目标无效"
		return
	}
	
	y := target.Y
	for x := 0; x < 5; x++ {
		node := defender.GetNodeAt(x, y)
		if node != nil {
			node.Status.ScanRevealed = true
			attacker.RevealNode(defender.ID, node.ID)
		}
	}
	action.Result = "嗅探成功"
	g.addLog(attacker.Username + " 嗅探了 " + defender.Username + " 的第 " + strconv.Itoa(y+1) + " 行节点")
}

func (g *Game) isTargetReachable(attacker, defender *Player, target *Node) bool {
	knownNodes := make([]*Node, 0)
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			node := defender.GetNodeAt(x, y)
			if node != nil && node.IsAlive() && defender.IsEdgeNode(x, y) {
				knownNodes = append(knownNodes, node)
			}
		}
	}
	
	for _, known := range knownNodes {
		if known.ID == target.ID {
			return true
		}
		if defender.IsNodeReachable(known.X, known.Y, target.X, target.Y) {
			return true
		}
	}
	
	return false
}

func (g *Game) checkGameEnd() {
	alivePlayers := make([]string, 0)
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		if player.IsAlive {
			alivePlayers = append(alivePlayers, playerID)
		}
	}
	
	if len(alivePlayers) <= 1 {
		g.Phase = PhaseGameOver
		if len(alivePlayers) == 1 {
			g.WinnerID = alivePlayers[0]
			g.addLog("游戏结束！胜者：" + g.Players[g.WinnerID].Username)
		} else {
			maxHP := -1
			var winner string
			for _, playerID := range g.PlayerOrder {
				player := g.Players[playerID]
				if player.CoreNode != nil && player.CoreNode.HP > maxHP {
					maxHP = player.CoreNode.HP
					winner = playerID
				}
			}
			g.WinnerID = winner
			g.addLog("游戏结束！核心血量最高者获胜：" + g.Players[winner].Username)
		}
		return
	}
	
	if g.CurrentTurn >= g.Config.MaxTurns {
		g.Phase = PhaseGameOver
		maxHP := -1
		var winner string
		for _, playerID := range g.PlayerOrder {
			player := g.Players[playerID]
			if player.CoreNode != nil && player.CoreNode.HP > maxHP {
				maxHP = player.CoreNode.HP
				winner = playerID
			}
		}
		g.WinnerID = winner
		g.addLog("回合数用尽！核心血量最高者获胜：" + g.Players[winner].Username)
	}
}

func (g *Game) addLog(message string) {
	g.GameLog = append(g.GameLog, message)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
