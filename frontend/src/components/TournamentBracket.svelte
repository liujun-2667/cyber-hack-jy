<script>
  import { getRankIcon, getRankColor } from '../utils/rank.js'

  export let matches = []
  export let currentRound = 0

  $: roundsArray = buildRoundsData(matches)

  function buildRoundsData(matches) {
    const grouped = {}
    matches.forEach(match => {
      const round = match.round || 1
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
    return {
      ...match,
      statusClass: status,
      p1,
      p2,
      isHighlighted,
      showInProgressBadge: status === 'in_progress',
      showByeBadge: status === 'bye'
    }
  }

  function getMatchStatus(match) {
    if (match.status === 'completed') return 'completed'
    if (match.status === 'in_progress') return 'in_progress'
    if (match.status === 'pending' && match.player1_id && match.player2_id) return 'ready'
    if (match.status === 'bye' || match.isBye) return 'bye'
    return 'pending'
  }

  function getPlayerInfo(match, side) {
    const playerId = side === 'p1' ? match.player1_id : match.player2_id
    const username = side === 'p1' ? match.player1_username : match.player2_username
    const rank = side === 'p1' ? match.player1_rank : match.player2_rank
    const seed = side === 'p1' ? match.player1_seed : match.player2_seed
    
    const isWinner = match.winner_id && match.winner_id === playerId
    const isLoser = match.status === 'completed' && match.winner_id && match.winner_id !== playerId
    const isBye = match.isBye && match.winner_id === playerId
    
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

<div class="bracket-container">
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
                <span class="vs-text">VS</span>
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

  .bracket-rounds {
    display: flex;
    gap: 20px;
    min-width: fit-content;
  }

  .bracket-round {
    display: flex;
    flex-direction: column;
    gap: 12px;
    min-width: 180px;
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

  .round-name {
    font-size: 13px;
    font-weight: 600;
    color: var(--neon-cyan);
  }

  .round-count {
    font-size: 11px;
    color: var(--text-secondary);
  }

  .round-matches {
    display: flex;
    flex-direction: column;
    gap: 10px;
    flex: 1;
    justify-content: space-around;
  }

  .match-card {
    position: relative;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 8px 10px;
    transition: all 0.3s;
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

  .player-rank {
    font-size: 14px;
  }

  .player-name {
    font-size: 13px;
    font-weight: 500;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .player-placeholder {
    font-size: 12px;
    color: var(--text-secondary);
    font-style: italic;
  }

  .match-divider {
    text-align: center;
    padding: 2px 0;
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
