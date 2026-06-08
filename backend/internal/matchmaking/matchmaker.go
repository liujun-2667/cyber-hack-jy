package matchmaking

import (
	"errors"
	"sync"
	"time"
)

type MatchRequest struct {
	PlayerID    string
	Username    string
	Client      interface{}
	GameMode    string
	MaxPlayers  int
	EloRating   int
	QueuedAt    time.Time
}

type MatchResult struct {
	Players []*MatchRequest
	RoomID  string
}

type MatchRangeInfo struct {
	CurrentRange   int
	WaitTime       int
	EstimatedRange string
}

const (
	RangeStage1 = 200
	RangeStage2 = 400
	RangeStage3 = 9999

	TimeStage1 = 10
	TimeStage2 = 20
)

type Matchmaker struct {
	queues     map[string][]*MatchRequest
	mu         sync.Mutex
	ticker     *time.Ticker
	stopChan   chan struct{}
	matchCb    func(string, *MatchResult)
}

func NewMatchmaker() *Matchmaker {
	return &Matchmaker{
		queues:   make(map[string][]*MatchRequest),
		stopChan: make(chan struct{}),
	}
}

func (m *Matchmaker) SetMatchCallback(cb func(string, *MatchResult)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.matchCb = cb
}

func (m *Matchmaker) Start() {
	m.ticker = time.NewTicker(1 * time.Second)
	go m.run()
}

func (m *Matchmaker) Stop() {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	close(m.stopChan)
}

func (m *Matchmaker) run() {
	for {
		select {
		case <-m.ticker.C:
			m.tryMatchAllQueues()
		case <-m.stopChan:
			return
		}
	}
}

func (m *Matchmaker) tryMatchAllQueues() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for queueKey, queue := range m.queues {
		if len(queue) < 2 {
			continue
		}

		matched := m.tryMatchQueue(queue)
		if len(matched) >= 2 && m.matchCb != nil {
			newQueue := make([]*MatchRequest, 0)
			matchedMap := make(map[string]bool)
			for _, p := range matched {
				matchedMap[p.PlayerID] = true
			}
			for _, p := range queue {
				if !matchedMap[p.PlayerID] {
					newQueue = append(newQueue, p)
				}
			}
			m.queues[queueKey] = newQueue

			result := &MatchResult{
				Players: matched,
			}

			go m.matchCb(queueKey, result)
		}
	}
}

func (m *Matchmaker) tryMatchQueue(queue []*MatchRequest) []*MatchRequest {
	if len(queue) < 2 {
		return nil
	}

	now := time.Now()

	for i := 0; i < len(queue); i++ {
		for j := i + 1; j < len(queue); j++ {
			p1 := queue[i]
			p2 := queue[j]

			maxWait := 0
			wait1 := int(now.Sub(p1.QueuedAt).Seconds())
			wait2 := int(now.Sub(p2.QueuedAt).Seconds())
			if wait1 > maxWait {
				maxWait = wait1
			}
			if wait2 > maxWait {
				maxWait = wait2
			}

			rangeLimit := m.getCurrentRange(maxWait)

			eloDiff := abs(p1.EloRating - p2.EloRating)
			if eloDiff <= rangeLimit {
				return []*MatchRequest{p1, p2}
			}
		}
	}

	return nil
}

func (m *Matchmaker) getCurrentRange(waitSeconds int) int {
	if waitSeconds >= TimeStage2 {
		return RangeStage3
	} else if waitSeconds >= TimeStage1 {
		return RangeStage2
	}
	return RangeStage1
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

	req.QueuedAt = time.Now()

	queue := m.queues[queueKey]

	if len(queue)+1 >= req.MaxPlayers {
		matched := false
		for i, existing := range queue {
			eloDiff := abs(existing.EloRating - req.EloRating)
			if eloDiff <= RangeStage1 {
				players := []*MatchRequest{existing, req}
				newQueue := make([]*MatchRequest, 0)
				for j, p := range queue {
					if j != i {
						newQueue = append(newQueue, p)
					}
				}
				m.queues[queueKey] = newQueue

				return &MatchResult{
					Players: players[:req.MaxPlayers],
				}, nil
			}
		}
		_ = matched
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

func (m *Matchmaker) GetMatchRangeInfo(playerID, gameMode string, maxPlayers int) *MatchRangeInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	queueKey := gameMode + "_" + string(rune(maxPlayers+'0'))
	queue := m.queues[queueKey]

	var req *MatchRequest
	for _, r := range queue {
		if r.PlayerID == playerID {
			req = r
			break
		}
	}

	if req == nil {
		return nil
	}

	waitSeconds := int(time.Since(req.QueuedAt).Seconds())
	currentRange := m.getCurrentRange(waitSeconds)

	estimatedRange := ""
	if currentRange == RangeStage1 {
		estimatedRange = "±200"
	} else if currentRange == RangeStage2 {
		estimatedRange = "±400"
	} else {
		estimatedRange = "不限"
	}

	return &MatchRangeInfo{
		CurrentRange:   currentRange,
		WaitTime:       waitSeconds,
		EstimatedRange: estimatedRange,
	}
}

func (m *Matchmaker) GetWaitTime(playerID, gameMode string, maxPlayers int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	queueKey := gameMode + "_" + string(rune(maxPlayers+'0'))
	queue := m.queues[queueKey]

	for _, req := range queue {
		if req.PlayerID == playerID {
			return int(time.Since(req.QueuedAt).Seconds())
		}
	}
	return -1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
