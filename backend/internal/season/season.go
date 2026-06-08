package season

import (
	"log"
	"sync"
	"time"

	"cyberhack/internal/database"
)

type SeasonManager struct {
	currentSeason *database.Season
	mu            sync.RWMutex
	ticker        *time.Ticker
	stopChan      chan struct{}
	broadcastFunc func(msgType string, payload interface{})
}

var instance *SeasonManager
var once sync.Once

func GetManager() *SeasonManager {
	once.Do(func() {
		instance = &SeasonManager{
			stopChan: make(chan struct{}),
		}
	})
	return instance
}

func (sm *SeasonManager) SetBroadcastFunc(fn func(msgType string, payload interface{})) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.broadcastFunc = fn
}

func (sm *SeasonManager) Start() {
	sm.loadCurrentSeason()
	sm.ticker = time.NewTicker(1 * time.Hour)
	
	go func() {
		for {
			select {
			case <-sm.ticker.C:
				sm.checkSeasonEnd()
			case <-sm.stopChan:
				return
			}
		}
	}()
	
	log.Println("Season manager started")
}

func (sm *SeasonManager) Stop() {
	if sm.ticker != nil {
		sm.ticker.Stop()
	}
	close(sm.stopChan)
}

func (sm *SeasonManager) loadCurrentSeason() {
	season, err := database.GetActiveSeason()
	if err != nil {
		log.Printf("Error loading active season: %v", err)
		return
	}
	
	sm.mu.Lock()
	sm.currentSeason = season
	sm.mu.Unlock()
	
	log.Printf("Current season: %s (ends: %s)", season.Name, season.EndDate.Format("2006-01-02"))
}

func (sm *SeasonManager) checkSeasonEnd() {
	sm.mu.RLock()
	season := sm.currentSeason
	sm.mu.RUnlock()
	
	if season == nil {
		return
	}
	
	if time.Now().After(season.EndDate) {
		log.Println("Season ended! Starting new season...")
		sm.endSeason()
	}
}

func (sm *SeasonManager) endSeason() {
	sm.mu.Lock()
	oldSeason := sm.currentSeason
	sm.mu.Unlock()
	
	if oldSeason == nil {
		return
	}
	
	err := database.ResetAllPlayersElo()
	if err != nil {
		log.Printf("Error resetting player ELO: %v", err)
		return
	}
	
	newSeasonNum := oldSeason.ID + 1
	newSeasonName := "第" + itoa(newSeasonNum) + "赛季"
	
	err = database.CreateNewSeason(newSeasonName, 30)
	if err != nil {
		log.Printf("Error creating new season: %v", err)
		return
	}
	
	sm.loadCurrentSeason()
	
	sm.mu.RLock()
	broadcast := sm.broadcastFunc
	sm.mu.RUnlock()
	
	if broadcast != nil {
		broadcast("season_reset", map[string]interface{}{
			"oldSeasonName": oldSeason.Name,
			"newSeasonName": newSeasonName,
			"message":      "新赛季开始！所有玩家积分已向1200回归。",
		})
	}
	
	log.Printf("Season %s ended, new season %s started", oldSeason.Name, newSeasonName)
}

func (sm *SeasonManager) GetCurrentSeason() *database.Season {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentSeason
}

func (sm *SeasonManager) GetTimeRemaining() time.Duration {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	if sm.currentSeason == nil {
		return 0
	}
	
	remaining := time.Until(sm.currentSeason.EndDate)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	
	var result []byte
	negative := false
	if i < 0 {
		negative = true
		i = -i
	}
	
	for i > 0 {
		result = append([]byte{byte('0' + i%10)}, result...)
		i /= 10
	}
	
	if negative {
		result = append([]byte{'-'}, result...)
	}
	
	return string(result)
}
