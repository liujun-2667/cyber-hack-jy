<script>
  import { gameStore } from '../store/gameStore.js'
  import { getRankInfo, getRankProgress, formatElo, getWinRate } from '../utils/rank.js'
  import StatsPanel from './StatsPanel.svelte'

  export let compact = false

  let showStats = false

  $: playerInfo = $gameStore.playerInfo
  $: eloRating = playerInfo?.eloRating || 1200
  $: rankKey = playerInfo?.currentRank || 'bronze'
  $: rankInfo = getRankInfo(rankKey)
  $: progress = getRankProgress(eloRating, rankKey)
  $: winRate = getWinRate(playerInfo?.wins || 0, playerInfo?.losses || 0)

  function toggleStats() {
    showStats = !showStats
  }
</script>

<div class="player-rank-container">
  <div class="rank-display" class:compact={compact} on:click={toggleStats}>
    <div class="rank-icon" style="color: {rankInfo.color}; text-shadow: 0 0 10px {rankInfo.glowColor};">
      {rankInfo.icon}
    </div>
    <div class="rank-info">
      <div class="rank-name" style="color: {rankInfo.color};">
        {rankInfo.name}
      </div>
      <div class="elo-rating">
        <span class="elo-value">{formatElo(eloRating)}</span>
        <span class="elo-label">ELO</span>
      </div>
      {#if !compact}
        <div class="rank-progress-bar">
          <div 
            class="rank-progress-fill" 
            style="width: {progress}%; background: {rankInfo.color};"
          ></div>
        </div>
      {/if}
    </div>
    {#if playerInfo?.currentStreak > 0}
      <div class="streak-badge">
        <span class="flame">🔥</span>
        <span>{playerInfo.currentStreak}连胜</span>
      </div>
    {/if}
    <div class="expand-icon">
      {showStats ? '▲' : '▼'}
    </div>
  </div>

  {#if showStats}
    <div class="stats-dropdown">
      <StatsPanel playerId={playerInfo?.playerId} />
    </div>
  {/if}
</div>

<style>
  .player-rank-container {
    position: relative;
    display: inline-block;
  }

  .rank-display {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 10px 16px;
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.3s ease;
  }

  .rank-display:hover {
    border-color: var(--neon-cyan);
    box-shadow: 0 0 15px rgba(0, 240, 255, 0.2);
  }

  .rank-display.compact {
    padding: 6px 12px;
    gap: 8px;
  }

  .rank-icon {
    font-size: 32px;
    line-height: 1;
  }

  .compact .rank-icon {
    font-size: 24px;
  }

  .rank-info {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 100px;
  }

  .compact .rank-info {
    min-width: 80px;
  }

  .rank-name {
    font-size: 14px;
    font-weight: bold;
    letter-spacing: 1px;
  }

  .compact .rank-name {
    font-size: 12px;
  }

  .elo-rating {
    display: flex;
    align-items: baseline;
    gap: 4px;
  }

  .elo-value {
    font-size: 18px;
    font-weight: bold;
    color: var(--text-primary);
  }

  .compact .elo-value {
    font-size: 14px;
  }

  .elo-label {
    font-size: 10px;
    color: var(--text-secondary);
    letter-spacing: 1px;
  }

  .rank-progress-bar {
    width: 100%;
    height: 4px;
    background: var(--bg-tertiary);
    border-radius: 2px;
    overflow: hidden;
    margin-top: 2px;
  }

  .rank-progress-fill {
    height: 100%;
    border-radius: 2px;
    transition: width 0.5s ease;
    box-shadow: 0 0 6px currentColor;
  }

  .streak-badge {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    background: rgba(255, 107, 53, 0.2);
    border: 1px solid #FF6B35;
    border-radius: 12px;
    font-size: 11px;
    color: #FF6B35;
    font-weight: bold;
  }

  .flame {
    font-size: 12px;
  }

  .expand-icon {
    font-size: 10px;
    color: var(--text-secondary);
    margin-left: 4px;
  }

  .stats-dropdown {
    position: absolute;
    top: calc(100% + 8px);
    right: 0;
    z-index: 100;
    min-width: 320px;
    animation: slideDown 0.3s ease;
  }

  @keyframes slideDown {
    from {
      opacity: 0;
      transform: translateY(-10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }
</style>
