package game

import (
	"math/rand"
)

type Player struct {
	ID            string     `json:"id"`
	Username      string     `json:"username"`
	Grid          [5][5]*Node `json:"grid"`
	Hand          []*Card    `json:"hand"`
	PlayedCards   []*PlayedCard `json:"playedCards"`
	MaxHandSize   int        `json:"maxHandSize"`
	Cooldowns     map[CardType]int `json:"cooldowns"`
	IsAlive       bool       `json:"isAlive"`
	CoreNode      *Node      `json:"-"`
	KnownOpponents map[string]map[string]bool `json:"-"`
}

func NewPlayer(id, username string) *Player {
	player := &Player{
		ID:            id,
		Username:      username,
		MaxHandSize:   7,
		Cooldowns:     make(map[CardType]int),
		IsAlive:       true,
		Hand:          make([]*Card, 0),
		PlayedCards:   make([]*PlayedCard, 0),
		KnownOpponents: make(map[string]map[string]bool),
	}
	return player
}

func (p *Player) InitializeGrid() {
	coreNode := NewNode(NodeTypeCore, 2, 2, p.ID)
	p.Grid[2][2] = coreNode
	p.CoreNode = coreNode
}

func (p *Player) PlaceNode(nodeType NodeType, x, y int) bool {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return false
	}
	if x == 2 && y == 2 {
		return false
	}
	if p.Grid[x][y] != nil {
		return false
	}
	p.Grid[x][y] = NewNode(nodeType, x, y, p.ID)
	return true
}

func (p *Player) GetNodeAt(x, y int) *Node {
	if x < 0 || x >= 5 || y < 0 || y >= 5 {
		return nil
	}
	return p.Grid[x][y]
}

func (p *Player) GetNodeByID(nodeID string) *Node {
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			if p.Grid[x][y] != nil && p.Grid[x][y].ID == nodeID {
				return p.Grid[x][y]
			}
		}
	}
	return nil
}

func (p *Player) GetAdjacentNodes(x, y int) []*Node {
	adjacent := make([]*Node, 0)
	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}
	for _, dir := range directions {
		nx, ny := x+dir[0], y+dir[1]
		if node := p.GetNodeAt(nx, ny); node != nil && node.IsAlive() {
			adjacent = append(adjacent, node)
		}
	}
	return adjacent
}

func (p *Player) IsNodeReachable(fromX, fromY, toX, toY int) bool {
	if fromX == toX && fromY == toY {
		return true
	}
	visited := make(map[string]bool)
	queue := [][2]int{{fromX, fromY}}
	visited[key(fromX, fromY)] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		cx, cy := current[0], current[1]

		for _, adj := range p.GetAdjacentNodes(cx, cy) {
			k := key(adj.X, adj.Y)
			if !visited[k] {
				if adj.X == toX && adj.Y == toY {
					return true
				}
				visited[k] = true
				queue = append(queue, [2]int{adj.X, adj.Y})
			}
		}
	}
	return false
}

func (p *Player) GetEdgeNodes() []*Node {
	edges := make([]*Node, 0)
	for x := 0; x < 5; x++ {
		for y := 0; y < 5; y++ {
			node := p.Grid[x][y]
			if node != nil && node.IsAlive() && p.IsEdgeNode(x, y) {
				edges = append(edges, node)
			}
		}
	}
	return edges
}

func (p *Player) IsEdgeNode(x, y int) bool {
	if x == 0 || x == 4 || y == 0 || y == 4 {
		return true
	}
	return false
}

func (p *Player) AddCard(card *Card) bool {
	if len(p.Hand) >= p.MaxHandSize {
		return false
	}
	p.Hand = append(p.Hand, card)
	return true
}

func (p *Player) RemoveCard(cardID string) *Card {
	for i, card := range p.Hand {
		if card.ID == cardID {
			p.Hand = append(p.Hand[:i], p.Hand[i+1:]...)
			return card
		}
	}
	return nil
}

func (p *Player) DrawCards(count int, cardIDPrefix string) []*Card {
	drawn := make([]*Card, 0)
	for i := 0; i < count; i++ {
		if len(p.Hand) >= p.MaxHandSize {
			break
		}
		rarity := "common"
		if rand.Float64() < 0.15 {
			rarity = "rare"
		} else if rand.Float64() < 0.3 {
			rarity = "uncommon"
		}
		card := DrawRandomCard(cardIDPrefix+"-"+string(rune(i+'0')), rarity)
		if card != nil && p.AddCard(card) {
			drawn = append(drawn, card)
		}
	}
	return drawn
}

func (p *Player) PlayCard(cardID, targetNodeID, targetPlayerID string) (*PlayedCard, bool) {
	card := p.RemoveCard(cardID)
	if card == nil {
		return nil, false
	}
	if cooldown, ok := p.Cooldowns[card.Type]; ok && cooldown > 0 {
		p.Hand = append(p.Hand, card)
		return nil, false
	}
	played := &PlayedCard{
		Card:           card,
		PlayerID:       p.ID,
		TargetNodeID:   targetNodeID,
		TargetPlayerID: targetPlayerID,
	}
	p.PlayedCards = append(p.PlayedCards, played)
	return played, true
}

func (p *Player) UpdateCooldowns() {
	for cardType := range p.Cooldowns {
		if p.Cooldowns[cardType] > 0 {
			p.Cooldowns[cardType]--
		}
	}
}

func (p *Player) CheckDefeat() bool {
	if p.CoreNode == nil {
		return true
	}
	if !p.CoreNode.IsAlive() {
		p.IsAlive = false
		return true
	}
	return false
}

func (p *Player) RevealNode(opponentID string, nodeID string) {
	if _, ok := p.KnownOpponents[opponentID]; !ok {
		p.KnownOpponents[opponentID] = make(map[string]bool)
	}
	p.KnownOpponents[opponentID][nodeID] = true
}

func (p *Player) IsNodeRevealed(opponentID string, nodeID string) bool {
	if known, ok := p.KnownOpponents[opponentID]; ok {
		return known[nodeID]
	}
	return false
}

func key(x, y int) string {
	return string(rune(x+'0')) + "," + string(rune(y+'0'))
}
