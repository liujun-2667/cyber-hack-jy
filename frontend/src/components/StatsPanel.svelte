<script>
  import { onMount } from 'svelte'
  import { gameStore } from '../store/gameStore.js'
  import { getWinRate, getCardName, formatElo, getRankInfo } from '../utils/rank.js'

  export let playerId = null

  let stats = null
  let loading = true

  $: playerInfo = $gameStore.playerInfo
  $: currentPlayerId = playerId || playerInfo?.playerId

  async function loadStats() {
    if (!currentPlayerId) return
    
    loading = true
    try {
      const data = await gameStore.fetchPlayerStats(currentPlayerId)
      stats = data
    } catch (e) {
      console.error('Failed to load stats:', e)
    } finally {
      loading = false
    }
  }

  onMount(() => {
    loadStats()
  })

  $: if (currentPlayerId) {
    loadStats()
  }
</script>

<div class="stats-panel">
  <div class="panel-header">
    <h3>战绩统计</h3>
  </div>

  {#if loading}
    <div class="loading">加载中...</div>
  {:else if stats}
    <div class="stats-content">
      <div class="stats-row">
        <div class="stat-item">
          <span class="stat-label">总场次</span>
          <span class="stat-value">{stats.totalGames}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">胜率</span>
          <span class="stat-value win-rate">{stats.winRate}%</span>
        </div>
      </div>

      <div class="stats-row">
        <div class="stat-item">
          <span class="stat-label">胜场</span>
          <span class="stat-value wins">{stats.wins}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">败场</span>
          <span class="stat-value losses">{stats.losses}</span>
        </div>
      </div>

      <div class="stats-row">
        <div class="stat-item">
          <span class="stat-label">当前连胜</span>
          <span class="stat-value streak">{stats.currentStreak}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">最高连胜</span>
          <span class="stat-value best-streak">{stats.bestStreak}</span>
        </div>
      </div>

      <div class="stats-divider"></div>

      <div class="stats-row">
        <div class="stat-item full">
          <span class="stat-label">场均击破节点</span>
          <span class="stat-value">{stats.avgNodesDestroyed}</span>
        </div>
      </div>

      <div class="stats-row">
        <div class="stat-item full">
          <span class="stat-label">场均存活回合</span>
          <span class="stat-value">{stats.avgTurnsSurvived}</span>
        </div>
      </div>

      {#if stats.topCards && stats.topCards.length > 0}
        <div class="stats-divider"></div>
        
        <div class="top-cards-section">
          <h4>最常用卡牌</h4>
          <div class="top-cards-list">
            {#each stats.topCards as card, i}
              <div class="top-card-item">
                <span class="card-rank">{i + 1}</span>
                <span class="card-name">{getCardName(card.cardType)}</span>
                <span class="card-count">{card.usageCount}次</span>
              </div>
            {/each}
          </div>
        </div>
      {/if}
    </div>
  {:else}
    <div class="no-data">暂无数据</div>
  {/if}
</div>

<style>
  .stats-panel {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    overflow: hidden;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  }

  .panel-header {
    padding: 12px 16px;
    background: rgba(0, 240, 255, 0.1);
    border-bottom: 1px solid var(--border-color);
  }

  .panel-header h3 {
    margin: 0;
    font-size: 14px;
    color: var(--neon-cyan);
    letter-spacing: 2px;
  }

  .stats-content {
    padding: 16px;
  }

  .stats-row {
    display: flex;
    gap: 16px;
    margin-bottom: 12px;
  }

  .stat-item {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: 6px;
  }

  .stat-item.full {
    flex: none;
    width: 100%;
    flex-direction: row;
    justify-content: space-between;
    align-items: center;
  }

  .stat-label {
    font-size: 11px;
    color: var(--text-secondary);
    letter-spacing: 1px;
  }

  .stat-value {
    font-size: 20px;
    font-weight: bold;
    color: var(--text-primary);
  }

  .full .stat-value {
    font-size: 16px;
  }

  .win-rate {
    color: var(--neon-green);
  }

  .wins {
    color: var(--neon-green);
  }

  .losses {
    color: var(--neon-red);
  }

  .streak {
    color: #FF6B35;
  }

  .best-streak {
    color: #FFD700;
  }

  .stats-divider {
    height: 1px;
    background: var(--border-color);
    margin: 12px 0;
  }

  .top-cards-section h4 {
    margin: 0 0 10px 0;
    font-size: 12px;
    color: var(--neon-pink);
    letter-spacing: 1px;
  }

  .top-cards-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .top-card-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 10px;
    background: var(--bg-tertiary);
    border-radius: 4px;
    font-size: 13px;
  }

  .card-rank {
    width: 20px;
    height: 20px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--neon-pink);
    color: white;
    font-size: 11px;
    font-weight: bold;
    border-radius: 50%;
  }

  .card-name {
    flex: 1;
    color: var(--text-primary);
  }

  .card-count {
    color: var(--text-secondary);
    font-size: 12px;
  }

  .loading,
  .no-data {
    padding: 30px;
    text-align: center;
    color: var(--text-secondary);
    font-size: 13px;
  }
</style>
