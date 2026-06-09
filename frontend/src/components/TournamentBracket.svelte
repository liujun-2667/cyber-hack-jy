<script>
  import { getRankIcon, getRankColor } from '../utils/rank.js'

  export let matches = []
  export let currentRound = 0
  export let compact = false

  $: roundsArray = buildRoundsData(matches)

  function buildRoundsData(matches) {
    const grouped = {}
    matches.forEach(match => {
      const round = match.roundNumber || match.round || 1
      if (!grouped[round]) {
        grouped[round] = []
      }
      grouped[round].push(buildMatchData(match))
    })

    const sortedRounds = Object.keys(grouped).sort((a, b) => a - b)
    const totalRounds = sortedRounds.length

    return sortedRounds.map(roundKey => {
      const roundNumber = parseInt(roundKey)
      const roundMatches = grouped[roundKey]
      return {
        roundNumber,
        roundName: getRoundName(roundNumber, totalRounds),
        matchCount: roundMatches.length,
        matches: roundMatches
      }
    })
  }

  function getRoundName(round, totalRounds) {
    if (round === totalRounds) return '决赛'
    if (round === totalRounds - 1) return '半决赛'
    if (round === totalRounds - 2) return '四分之一决赛'
    return `第${round}轮`
  }

  function buildMatchData(match) {
    const status = getMatchStatus(match)
    const p1 = getPlayerInfo(match, 'p1')
    const p2 = getPlayerInfo(match, 'p2')
    const isHighlighted = match.status === 'in_progress'
    const progressInfo = getProgressInfo(match, status)
    return {
      ...match,
      statusClass: status,
      p1,
      p2,
      isHighlighted,
      showInProgressBadge: status === 'in_progress',
      showByeBadge: status === 'bye',
      progressInfo
    }
  }

  function getProgressInfo(match, status) {
    if (status === 'completed' || status === 'bye') {
      const winnerId = match.winnerId || match.winner_id
      const p1Id = match.player1Id || match.player1_id
      if (winnerId && p1Id) {
        return { type: 'score', text: winnerId === p1Id ? '1 - 0' : '0 - 1' }
      }
      return { type: 'finished', text: '已结束' }
    }
    if (status === 'in_progress') {
      const currentTurn = match.currentTurn
      const maxTurns = match.maxTurns || 30
      if (currentTurn) {
        return { type: 'progress', text: `第${currentTurn}/${maxTurns}回合`, current: currentTurn, max: maxTurns }
      }
      return { type: 'progress', text: '进行中' }
    }
    if (status === 'ready') {
      return { type: 'waiting', text: '等待中' }
    }
    return { type: 'pending', text: '待定' }
  }

  function getMatchStatus(match) {
    if (match.status === 'completed' || match.status === 'finished') return 'completed'
    if (match.status === 'in_progress') return 'in_progress'
    if (match.status === 'pending' && (match.player1_id || match.player1Id) && (match.player2_id || match.player2Id)) return 'ready'
    if (match.status === 'bye' || match.isBye) return 'bye'
    return 'pending'
  }

  function getPlayerInfo(match, side) {
    const playerId = side === 'p1' ? (match.player1Id || match.player1_id) : (match.player2Id || match.player2_id)
    const username = side === 'p1' ? (match.player1Name || match.player1_name || match.player1Username) : (match.player2Name || match.player2_name || match.player2Username)
    const rank = side === 'p1' ? match.player1Rank : match.player2Rank
    const seed = side === 'p1' ? match.player1Seed : match.player2Seed
    
    const winnerId = match.winnerId || match.winner_id
    const isWinner = winnerId && winnerId === playerId
    const isLoser = (match.status === 'completed' || match.status === 'finished') && winnerId && winnerId !== playerId
    const isBye = match.isBye && winnerId === playerId
    
    return {
      id: playerId,
      username: username || '待定',
      rank: rank || 'bronze',
      seed: seed || 0,
      isWinner,
      isLoser,
      isBye,
      hasPlayer: !!playerId
    }
  }
</script>

<div class="bracket-container" class:compact>
  <div class="bracket-rounds">
    {#each roundsArray as round (round.roundNumber)}
      <div class="bracket-round">
        <div class="round-header">
          <span class="round-name">{round.roundName}</span>
          <span class="round-count">{round.matchCount}场</span>
        </div>
        <div class="round-matches">
          {#each round.matches as match (match.id)}
            <div 
              class="match-card {match.statusClass}" 
              class:highlighted={match.isHighlighted}
            >
              <div class="match-player" class:winner={match.p1.isWinner} class:loser={match.p1.isLoser} class:bye={match.p1.isBye}>
                {#if match.p1.hasPlayer}
                  <span class="player-seed">#{match.p1.seed}</span>
                  <span class="player-rank" style="color: {getRankColor(match.p1.rank)}">{getRankIcon(match.p1.rank)}</span>
                  <span class="player-name">{match.p1.username}</span>
                {:else}
                  <span class="player-placeholder">待定</span>
                {/if}
              </div>
              
              <div class="match-divider">
                {#if match.progressInfo.type === 'progress' && match.progressInfo.current}
                  <div class="progress-wrapper">
                    <div class="progress-bar">
                      <div class="progress-fill" style="width: {Math.min(100, (match.progressInfo.current / match.progressInfo.max) * 100)}%"></div>
                    </div>
                    <span class="progress-text">{match.progressInfo.text}</span>
                  </div>
                {:else}
                  <span class="vs-text">{match.progressInfo.text}</span>
                {/if}
              </div>
              
              <div class="match-player" class:winner={match.p2.isWinner} class:loser={match.p2.isLoser} class:bye={match.p2.isBye}>
                {#if match.p2.hasPlayer}
                  <span class="player-seed">#{match.p2.seed}</span>
                  <span class="player-rank" style="color: {getRankColor(match.p2.rank)}">{getRankIcon(match.p2.rank)}</span>
                  <span class="player-name">{match.p2.username}</span>
                {:else}
                  <span class="player-placeholder">待定</span>
                {/if}
              </div>

              {#if match.showInProgressBadge}
                <div class="match-status-badge in-progress">
                  进行中
                </div>
              {/if}
              {#if match.showByeBadge}
                <div class="match-status-badge bye">
                  轮空
                </div>
              {/if}
            </div>
          {/each}
        </div>
      </div>
    {/each}
  </div>
</div>

<style>
  .bracket-container {
    width: 100%;
    overflow-x: auto;
    padding: 10px 0;
  }

  .bracket-container.compact {
    padding: 4px 0;
    font-size: 0.85em;
  }

  .bracket-rounds {
    display: flex;
    gap: 20px;
    min-width: fit-content;
  }

  .bracket-container.compact .bracket-rounds {
    gap: 12px;
  }

  .bracket-round {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 180px;
  }

  .bracket-container.compact .bracket-round {
    min-width: 140px;
    gap: 8px;
  }

  .round-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: rgba(0, 240, 255, 0.05);
    border-radius: 6px;
    border: 1px solid rgba(0, 240, 255, 0.2);
  }

  .bracket-container.compact .round-header {
    padding: 6px 8px;
  }

  .round-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--neon-cyan);
  }

  .bracket-container.compact .round-name {
    font-size: 11px;
  }

  .round-count {
    font-size: 11px;
    color: var(--text-secondary);
  }

  .bracket-container.compact .round-count {
    font-size: 10px;
  }

  .round-matches {
    display: flex;
    flex-direction: column;
    gap: 10px;
    flex: 1;
    justify-content: space-around;
  }

  .bracket-container.compact .round-matches {
    gap: 6px;
  }

  .match-card {
    position: relative;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 8px 10px;
    transition: all 0.3s;
  }

  .bracket-container.compact .match-card {
    padding: 6px 8px;
    border-radius: 6px;
  }

  .match-card.pending {
    opacity: 0.6;
  }

  .match-card.highlighted {
    border-color: var(--neon-pink);
    box-shadow: 0 0 15px rgba(255, 51, 102, 0.3);
  }

  .match-card.completed {
    border-color: rgba(0, 255, 100, 0.3);
  }

  .match-card.bye {
    border-style: dashed;
    opacity: 0.7;
  }

  .match-player {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 8px;
    border-radius: 4px;
    transition: all 0.3s;
  }

  .bracket-container.compact .match-player {
    padding: 4px 6px;
    gap: 4px;
  }

  .match-player.winner {
    background: rgba(0, 255, 100, 0.1);
    color: #00ff64;
  }

  .match-player.loser {
    opacity: 0.4;
    filter: grayscale(0.5);
  }

  .match-player.bye {
    opacity: 0.6;
    font-style: italic;
  }

  .player-seed {
    font-size: 10px;
    color: var(--text-secondary);
    font-weight: bold;
    min-width: 20px;
  }

  .bracket-container.compact .player-seed {
    font-size: 9px;
    min-width: 16px;
  }

  .player-rank {
    font-size: 14px;
  }

  .bracket-container.compact .player-rank {
    font-size: 12px;
  }

  .player-name {
    font-size: 13px;
    font-weight: 500;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .bracket-container.compact .player-name {
    font-size: 11px;
  }

  .player-placeholder {
    font-size: 12px;
    color: var(--text-secondary);
    font-style: italic;
  }

  .bracket-container.compact .player-placeholder {
    font-size: 10px;
  }

  .match-divider {
    text-align: center;
    padding: 2px 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
  }

  .progress-wrapper {
    width: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 3px;
  }

  .progress-bar {
    width: 80%;
    height: 4px;
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
    overflow: hidden;
  }

  .progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--neon-pink), #FFD700);
    transition: width 0.5s ease;
    border-radius: 2px;
  }

  .progress-text {
    font-size: 10px;
    color: #FFD700;
    font-weight: 600;
    letter-spacing: 0.5px;
  }

  .vs-text {
    font-size: 10px;
    color: var(--text-secondary);
    letter-spacing: 2px;
    font-weight: bold;
  }

  .match-status-badge {
    position: absolute;
    top: -8px;
    right: 8px;
    padding: 2px 8px;
    border-radius: 10px;
    font-size: 10px;
    font-weight: bold;
  }

  .match-status-badge.in-progress {
    background: var(--neon-pink);
    color: #fff;
    animation: pulse 1.5s infinite;
  }

  .match-status-badge.bye {
    background: rgba(255, 215, 0, 0.2);
    color: #FFD700;
    border: 1px solid rgba(255, 215, 0, 0.5);
  }

  @keyframes pulse {
    0%, 100% {
      opacity: 1;
    }
    50% {
      opacity: 0.7;
    }
  }
</style>
