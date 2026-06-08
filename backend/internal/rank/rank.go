package rank

import (
	"math"
)

type Rank string

const (
	RankBronze   Rank = "bronze"
	RankSilver   Rank = "silver"
	RankGold     Rank = "gold"
	RankPlatinum Rank = "platinum"
	RankDiamond  Rank = "diamond"
)

type RankThresholds struct {
	MinElo int
	MaxElo int
}

var rankThresholds = map[Rank]RankThresholds{
	RankBronze:   {MinElo: 0, MaxElo: 1099},
	RankSilver:   {MinElo: 1100, MaxElo: 1399},
	RankGold:     {MinElo: 1400, MaxElo: 1699},
	RankPlatinum: {MinElo: 1700, MaxElo: 1999},
	RankDiamond:  {MinElo: 2000, MaxElo: 9999},
}

var rankOrder = []Rank{RankBronze, RankSilver, RankGold, RankPlatinum, RankDiamond}

var rankNames = map[Rank]string{
	RankBronze:   "青铜",
	RankSilver:   "白银",
	RankGold:     "黄金",
	RankPlatinum: "铂金",
	RankDiamond:  "钻石",
}

var rankColors = map[Rank]string{
	RankBronze:   "#CD7F32",
	RankSilver:   "#C0C0C0",
	RankGold:     "#FFD700",
	RankPlatinum: "#E5E4E2",
	RankDiamond:  "#B9F2FF",
}

func GetRank(elo int) Rank {
	for i := len(rankOrder) - 1; i >= 0; i-- {
		rank := rankOrder[i]
		threshold := rankThresholds[rank]
		if elo >= threshold.MinElo {
			return rank
		}
	}
	return RankBronze
}

func GetRankName(rank Rank) string {
	if name, ok := rankNames[rank]; ok {
		return name
	}
	return "未知"
}

func GetRankColor(rank Rank) string {
	if color, ok := rankColors[rank]; ok {
		return color
	}
	return "#888888"
}

func GetRankThreshold(rank Rank) RankThresholds {
	if threshold, ok := rankThresholds[rank]; ok {
		return threshold
	}
	return rankThresholds[RankBronze]
}

func GetRankIndex(rank Rank) int {
	for i, r := range rankOrder {
		if r == rank {
			return i
		}
	}
	return 0
}

type EloChangeResult struct {
	WinnerChange int
	LoserChange  int
	WinnerNewElo int
	LoserNewElo  int
	WinnerRank   Rank
	LoserRank    Rank
	RankChange   string
	WinnerPrevRank Rank
	LoserPrevRank  Rank
}

const (
	MaxEloGain = 50
	MinEloLoss = 10
	BaseK      = 32
)

func CalculateEloChange(winnerElo, loserElo int) EloChangeResult {
	winnerPrevRank := GetRank(winnerElo)
	loserPrevRank := GetRank(loserElo)

	expectedWinner := 1.0 / (1.0 + math.Pow(10, float64(loserElo-winnerElo)/400.0))
	expectedLoser := 1.0 - expectedWinner

	eloDiff := math.Abs(float64(winnerElo - loserElo))

	k := float64(BaseK)
	if eloDiff > 200 {
		k = 20
	} else if eloDiff > 100 {
		k = 26
	}

	winnerChange := int(math.Round(k * (1.0 - expectedWinner)))
	loserChange := int(math.Round(k * (0.0 - expectedLoser)))

	if winnerElo < loserElo {
		if winnerChange > MaxEloGain {
			winnerChange = MaxEloGain
		}
		if loserChange < -MinEloLoss {
			loserChange = -MinEloLoss
		}
	} else {
		if winnerChange < 10 {
			winnerChange = 10
		}
		if loserChange < -MaxEloGain {
			loserChange = -MaxEloGain
		}
	}

	winnerNewElo := winnerElo + winnerChange
	loserNewElo := loserElo + loserChange

	if winnerNewElo < 0 {
		winnerNewElo = 0
	}
	if loserNewElo < 0 {
		loserNewElo = 0
	}

	winnerNewRank := GetRank(winnerNewElo)
	loserNewRank := GetRank(loserNewElo)

	rankChange := "none"
	if GetRankIndex(winnerNewRank) > GetRankIndex(winnerPrevRank) {
		rankChange = "promote"
	} else if GetRankIndex(loserNewRank) < GetRankIndex(loserPrevRank) {
		rankChange = "demote"
	}

	return EloChangeResult{
		WinnerChange:   winnerChange,
		LoserChange:    loserChange,
		WinnerNewElo:   winnerNewElo,
		LoserNewElo:    loserNewElo,
		WinnerRank:     winnerNewRank,
		LoserRank:      loserNewRank,
		RankChange:     rankChange,
		WinnerPrevRank: winnerPrevRank,
		LoserPrevRank:  loserPrevRank,
	}
}

type RankProtection struct {
	GamesRemaining int
	CurrentRank    Rank
}

func NewRankProtection(rank Rank) *RankProtection {
	return &RankProtection{
		GamesRemaining: 3,
		CurrentRank:    rank,
	}
}

func (rp *RankProtection) Decrement() {
	if rp.GamesRemaining > 0 {
		rp.GamesRemaining--
	}
}

func (rp *RankProtection) IsProtected() bool {
	return rp.GamesRemaining > 0
}

func ApplyRankProtection(prevElo, newElo int, protectionGames int, prevRank Rank) (int, Rank, int) {
	newRank := GetRank(newElo)
	newProtectionGames := protectionGames

	if GetRankIndex(newRank) < GetRankIndex(prevRank) && protectionGames > 0 {
		minElo := GetRankThreshold(prevRank).MinElo
		if newElo < minElo {
			newElo = minElo
		}
		newRank = prevRank
		newProtectionGames = protectionGames - 1
		if newProtectionGames < 0 {
			newProtectionGames = 0
		}
	}

	if GetRankIndex(newRank) > GetRankIndex(prevRank) {
		newProtectionGames = 3
	} else if GetRankIndex(newRank) == GetRankIndex(prevRank) && protectionGames > 0 {
		newProtectionGames = protectionGames - 1
		if newProtectionGames < 0 {
			newProtectionGames = 0
		}
	}

	return newElo, newRank, newProtectionGames
}

func CalculateSeasonReset(currentElo int) int {
	baseElo := 1200
	diff := currentElo - baseElo
	return baseElo + diff/2
}
