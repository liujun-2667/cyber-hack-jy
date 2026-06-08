package game

type NodeType string

const (
	NodeTypeRouter   NodeType = "router"
	NodeTypeServer   NodeType = "server"
	NodeTypeDatabase NodeType = "database"
	NodeTypeFirewall NodeType = "firewall"
	NodeTypeCore     NodeType = "core"
)

type NodeStatus struct {
	ScanRevealed bool
	HasTrojan    bool
	HasDDoS      bool
	HasHoneypot  bool
	HasIDS       bool
	FirewallTurns int
}

type Node struct {
	ID         string     `json:"id"`
	Type       NodeType   `json:"type"`
	X          int        `json:"x"`
	Y          int        `json:"y"`
	MaxHP      int        `json:"maxHp"`
	HP         int        `json:"hp"`
	Defense    int        `json:"defense"`
	Bandwidth  int        `json:"bandwidth"`
	MaxBandwidth int     `json:"maxBandwidth"`
	Programs   []string   `json:"programs"`
	Status     NodeStatus `json:"status"`
	OwnerID    string     `json:"ownerId"`
}

type GridPosition struct {
	X int
	Y int
}

func NewNode(nodeType NodeType, x, y int, ownerID string) *Node {
	node := &Node{
		ID:      ownerID + "-" + string(rune(x+'0')) + string(rune(y+'0')),
		Type:    nodeType,
		X:       x,
		Y:       y,
		OwnerID: ownerID,
		Status:  NodeStatus{},
	}

	switch nodeType {
	case NodeTypeCore:
		node.MaxHP = 30
		node.HP = 30
		node.Defense = 3
		node.Bandwidth = 5
		node.MaxBandwidth = 5
	case NodeTypeServer:
		node.MaxHP = 20
		node.HP = 20
		node.Defense = 2
		node.Bandwidth = 4
		node.MaxBandwidth = 4
	case NodeTypeDatabase:
		node.MaxHP = 15
		node.HP = 15
		node.Defense = 1
		node.Bandwidth = 3
		node.MaxBandwidth = 3
	case NodeTypeFirewall:
		node.MaxHP = 25
		node.HP = 25
		node.Defense = 5
		node.Bandwidth = 2
		node.MaxBandwidth = 2
	case NodeTypeRouter:
		node.MaxHP = 10
		node.HP = 10
		node.Defense = 0
		node.Bandwidth = 6
		node.MaxBandwidth = 6
	}

	return node
}

func (n *Node) IsAlive() bool {
	return n.HP > 0
}

func (n *Node) TakeDamage(damage int) int {
	actualDamage := damage - n.Defense
	if actualDamage < 0 {
		actualDamage = 0
	}
	if n.Status.FirewallTurns > 0 {
		actualDamage = actualDamage - 3
		if actualDamage < 0 {
			actualDamage = 0
		}
	}
	n.HP -= actualDamage
	if n.HP < 0 {
		n.HP = 0
	}
	return actualDamage
}

func (n *Node) Heal(amount int) {
	n.HP += amount
	if n.HP > n.MaxHP {
		n.HP = n.MaxHP
	}
}
