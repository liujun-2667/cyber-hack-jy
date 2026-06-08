package matchmaking

import (
	"errors"
	"sync"
)

type MatchRequest struct {
	PlayerID   string
	Username   string
	Client     interface{}
	GameMode   string
	MaxPlayers int
	EloRating  int
}

type MatchResult struct {
	Players []*MatchRequest
	RoomID  string
}

type Matchmaker struct {
	queues map[string][]*MatchRequest
	mu     sync.Mutex
}

func NewMatchmaker() *Matchmaker {
	return &Matchmaker{
		queues: make(map[string][]*MatchRequest),
	}
}

func (m *Matchmaker) AddToQueue(req *MatchRequest) (*MatchResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	queueKey := req.GameMode + "_" + string(rune(req.MaxPlayers+'0'))

	for _, existing := range m.queues[queueKey] {
		if existing.PlayerID == req.PlayerID {
			return nil, errors.New("already in queue")
		}
	}

	queue := m.queues[queueKey]

	if len(queue)+1 >= req.MaxPlayers {
		players := append(queue, req)
		m.queues[queueKey] = make([]*MatchRequest, 0)

		return &MatchResult{
			Players: players[:req.MaxPlayers],
		}, nil
	}

	m.queues[queueKey] = append(queue, req)
	return nil, nil
}

func (m *Matchmaker) RemoveFromQueue(playerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for key, queue := range m.queues {
		for i, req := range queue {
			if req.PlayerID == playerID {
				m.queues[key] = append(queue[:i], queue[i+1:]...)
				return
			}
		}
	}
}

func (m *Matchmaker) GetQueuePosition(playerID, gameMode string, maxPlayers int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	queueKey := gameMode + "_" + string(rune(maxPlayers+'0'))
	queue := m.queues[queueKey]

	for i, req := range queue {
		if req.PlayerID == playerID {
			return i + 1
		}
	}
	return -1
}
