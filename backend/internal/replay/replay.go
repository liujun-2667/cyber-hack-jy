package replay

import (
	"sync"
	"time"

	"cyberhack/internal/game"
	"github.com/google/uuid"
)

const MaxReplaysPerRoom = 20

type Replay struct {
	ID         string              `json:"id"`
	RoomID     string              `json:"roomId"`
	CreatedAt  time.Time           `json:"createdAt"`
	Players    []ReplayPlayer      `json:"players"`
	TotalTurns int                 `json:"totalTurns"`
	WinnerID   string              `json:"winnerId"`
	Turns      []*game.ReplayTurn  `json:"turns"`
}

type ReplayPlayer struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ReplayStore struct {
	replays map[string][]*Replay
	mu      sync.RWMutex
}

var store = &ReplayStore{
	replays: make(map[string][]*Replay),
}

func GetStore() *ReplayStore {
	return store
}

func CreateReplay(roomID string, g *game.Game) *Replay {
	players := make([]ReplayPlayer, 0, len(g.PlayerOrder))
	for _, playerID := range g.PlayerOrder {
		player := g.Players[playerID]
		if player != nil {
			players = append(players, ReplayPlayer{
				ID:       player.ID,
				Username: player.Username,
			})
		}
	}

	replay := &Replay{
		ID:         uuid.New().String(),
		RoomID:     roomID,
		CreatedAt:  time.Now(),
		Players:    players,
		TotalTurns: g.CurrentTurn,
		WinnerID:   g.WinnerID,
		Turns:      g.ReplayTurns,
	}

	return replay
}

func (s *ReplayStore) AddReplay(roomID string, replay *Replay) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.replays[roomID]; !exists {
		s.replays[roomID] = make([]*Replay, 0, MaxReplaysPerRoom)
	}

	s.replays[roomID] = append([]*Replay{replay}, s.replays[roomID]...)

	if len(s.replays[roomID]) > MaxReplaysPerRoom {
		s.replays[roomID] = s.replays[roomID][:MaxReplaysPerRoom]
	}
}

func (s *ReplayStore) GetReplays(roomID string) []*Replay {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if replays, exists := s.replays[roomID]; exists {
		result := make([]*Replay, len(replays))
		copy(result, replays)
		return result
	}
	return []*Replay{}
}

func (s *ReplayStore) GetReplay(roomID, replayID string) *Replay {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if replays, exists := s.replays[roomID]; exists {
		for _, r := range replays {
			if r.ID == replayID {
				return r
			}
		}
	}
	return nil
}
