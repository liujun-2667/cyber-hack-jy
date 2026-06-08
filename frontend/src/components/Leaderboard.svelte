<script>
  import { onMount, onDestroy } from 'svelte'
  import { gameStore } from '../store/gameStore.js'
  import { getRankInfo, formatElo } from '../utils/rank.js'

  export let onClose

  let leaderboard = []
  let playerRank = null
  let loading = true
  let myPlayerId = null
  let lastUpdateTime = 0
  let secondsAgo = 0
  let refreshInterval = null
  let timerInterval = null
  let updatedIds = new Set()
  let newIds = new Set()

  const REFRESH_INTERVAL = 15000

  $: playerInfo = $gameStore.playerInfo

  function updateSecondsAgo() {
    if (lastUpdateTime > 0) {
      secondsAgo = Math.floor((Date.now() - lastUpdateTime) / 1000)
    }
  }

  function isEntriesEqual(a, b) {
    const keysA = Object.keys(a).filter(k => !k.startsWith('_'))
    const keysB = Object.keys(b).filter(k => !k.startsWith('_'))
    if (keysA.length !== keysB.length) return false
    for (const key of keysA) {
      if (a[key] !== b[key]) return false
    }
    return true
  }

  function findChangedIds(oldEntries, newEntries) {
    const oldMap = new Map(oldEntries.map(e => [e.playerId, e]))
    const changed = new Set()
    const added = new Set()
    
    for (const newEntry of newEntries) {
      const oldEntry = oldMap.get(newEntry.playerId)
      if (oldEntry) {
        if (!isEntriesEqual(oldEntry, newEntry)) {
          changed.add(newEntry.playerId)
        }
      } else {
        added.add(newEntry.playerId)
      }
    }
    
    return { changed, added }
  }

  function triggerAnimation(changed, added) {
    updatedIds = new Set(changed)
    newIds = new Set(added)
    
    setTimeout(() => {
      updatedIds = new Set()
      newIds = new Set()
    }, 1000)
  }

  async function loadLeaderboard() {
    try {
      const pid = myPlayerId || playerInfo?.playerId
      const data = await gameStore.fetchLeaderboard(20, pid)
      if (data) {
        const newLeaderboard = data.leaderboard || []
        const oldLeaderboard = leaderboard
        
        if (oldLeaderboard.length > 0) {
          const { changed, added } = findChangedIds(oldLeaderboard, newLeaderboard)
          if (changed.size > 0 || added.size > 0) {
            triggerAnimation(changed, added)
          }
        }
        
        leaderboard = newLeaderboard
        playerRank = data.playerRank || null
        lastUpdateTime = Date.now()
        secondsAgo = 0
      }
    } catch (e) {
      console.error('Failed to load leaderboard:', e)
    } finally {
      loading = false
    }
  }

  function getRankBadgeClass(rank) {
    if (rank === 1) return 'gold'
    if (rank === 2) return 'silver'
    if (rank === 3) return 'bronze'
    return 'normal'
  }

  function startAutoRefresh() {
    if (refreshInterval) clearInterval(refreshInterval)
    if (timerInterval) clearInterval(timerInterval)
    
    refreshInterval = setInterval(() => {
      loadLeaderboard()
    }, REFRESH_INTERVAL)
    
    timerInterval = setInterval(() => {
      updateSecondsAgo()
    }, 1000)
  }

  function stopAutoRefresh() {
    if (refreshInterval) {
      clearInterval(refreshInterval)
      refreshInterval = null
    }
    if (timerInterval) {
      clearInterval(timerInterval)
      timerInterval = null
    }
  }

  onMount(() => {
    if (playerInfo?.playerId) {
      myPlayerId = playerInfo.playerId
    }
    loadLeaderboard().then(() => {
      startAutoRefresh()
    })
  })

  onDestroy(() => {
    stopAutoRefresh()
  })

  $: if (playerInfo?.playerId && !myPlayerId) {
    myPlayerId = playerInfo.playerId
    loadLeaderboard()
  }
</script>

<div class="leaderboard-overlay" on:click|self={onClose}>
  <div class="leaderboard-modal">
    <div class="modal-header">
      <h2 class="neon-text-cyan">🏆 全服排行榜</h2>
      <div class="header-right">
        <span class="last-update">上次更新: {secondsAgo}秒前</span>
        <button class="close-btn" on:click={onClose}>✕</button>
      </div>
    </div>

    <div class="modal-content">
      {#if loading}
        <div class="loading">加载中...</div>
      {:else}
        <div class="leaderboard-list">
          <div class="list-header">
            <span class="col-rank">排名</span>
            <span class="col-player">玩家</span>
            <span class="col-elo">积分</span>
            <span class="col-winrate">胜率</span>
          </div>

          {#each leaderboard as entry (entry.playerId)}
            <div 
              class="leaderboard-item"
              class:is-me={entry.playerId === myPlayerId}
              class:updated={updatedIds.has(entry.playerId)}
              class:new-entry={newIds.has(entry.playerId)}
            >
              <span class="col-rank">
                <span class="rank-badge {getRankBadgeClass(entry.rank)}">
                  {entry.rank}
                </span>
              </span>
              <span class="col-player">
                <span class="rank-icon" style="color: {getRankInfo(entry.rankName)?.color};">
                  {getRankInfo(entry.rankName)?.icon || '❓'}
                </span>
                <span class="player-name">{entry.username}</span>
              </span>
              <span class="col-elo">
                <span class="elo-value">{formatElo(entry.eloRating)}</span>
              </span>
              <span class="col-winrate">
                <span class="winrate">{entry.winRate}%</span>
              </span>
            </div>
          {/each}

          {#if playerRank && !playerRank.isInTop}
            <div class="my-rank-divider">
              <span>你的排名</span>
            </div>
            <div class="leaderboard-item is-me">
              <span class="col-rank">
                <span class="rank-badge normal">
                  {playerRank.rank}
                </span>
              </span>
              <span class="col-player">
                <span class="rank-icon" style="color: {getRankInfo(playerRank.player?.rankName)?.color};">
                  {getRankInfo(playerRank.player?.rankName)?.icon || '❓'}
                </span>
                <span class="player-name">{playerRank.player?.username}</span>
              </span>
              <span class="col-elo">
                <span class="elo-value">{formatElo(playerRank.player?.eloRating || 0)}</span>
              </span>
              <span class="col-winrate">
                <span class="winrate">{playerRank.player?.winRate || 0}%</span>
              </span>
            </div>
          {/if}
        </div>
      {/if}
    </div>

    <div class="modal-footer">
      <button class="btn-neon" on:click={onClose}>关闭</button>
    </div>
  </div>
</div>

<style>
  .leaderboard-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.85);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 200;
  }

  .leaderboard-modal {
    width: 90%;
    max-width: 500px;
    max-height: 80vh;
    background: var(--bg-secondary);
    border: 1px solid var(--neon-cyan);
    border-radius: 12px;
    box-shadow: 0 0 30px rgba(0, 240, 255, 0.3);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .modal-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 20px 24px;
    border-bottom: 1px solid var(--border-color);
    background: rgba(0, 240, 255, 0.05);
  }

  .modal-header h2 {
    margin: 0;
    font-size: 20px;
    letter-spacing: 2px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .last-update {
    font-size: 11px;
    color: var(--text-secondary);
    letter-spacing: 1px;
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    font-size: 20px;
    cursor: pointer;
    padding: 4px 8px;
    transition: color 0.3s;
  }

  .close-btn:hover {
    color: var(--neon-pink);
  }

  .modal-content {
    flex: 1;
    overflow-y: auto;
    padding: 16px 24px;
  }

  .loading {
    text-align: center;
    padding: 40px;
    color: var(--text-secondary);
  }

  .leaderboard-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .list-header {
    display: flex;
    align-items: center;
    padding: 8px 12px;
    font-size: 11px;
    color: var(--text-secondary);
    letter-spacing: 1px;
    text-transform: uppercase;
    border-bottom: 1px solid var(--border-color);
    margin-bottom: 8px;
  }

  .col-rank { width: 60px; text-align: center; }
  .col-player { flex: 1; }
  .col-elo { width: 80px; text-align: right; }
  .col-winrate { width: 60px; text-align: right; }

  .leaderboard-item {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border-radius: 6px;
    transition: all 0.3s ease;
  }

  .leaderboard-item:hover {
    background: rgba(0, 240, 255, 0.1);
  }

  .leaderboard-item.is-me {
    background: rgba(0, 240, 255, 0.15);
    border: 1px solid var(--neon-cyan);
  }

  .leaderboard-item.updated {
    animation: updateFlash 1s ease-out;
  }

  .leaderboard-item.new-entry {
    animation: slideIn 0.5s ease-out;
  }

  @keyframes updateFlash {
    0% {
      background: rgba(255, 215, 0, 0.4);
    }
    100% {
      background: var(--bg-tertiary);
    }
  }

  @keyframes slideIn {
    0% {
      opacity: 0;
      transform: translateX(20px);
    }
    100% {
      opacity: 1;
      transform: translateX(0);
    }
  }

  .rank-badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    border-radius: 50%;
    font-weight: bold;
    font-size: 12px;
  }

  .rank-badge.gold {
    background: linear-gradient(135deg, #FFD700, #FFA500);
    color: #000;
    box-shadow: 0 0 10px rgba(255, 215, 0, 0.5);
  }

  .rank-badge.silver {
    background: linear-gradient(135deg, #C0C0C0, #808080);
    color: #000;
    box-shadow: 0 0 8px rgba(192, 192, 192, 0.5);
  }

  .rank-badge.bronze {
    background: linear-gradient(135deg, #CD7F32, #8B4513);
    color: #fff;
    box-shadow: 0 0 8px rgba(205, 127, 50, 0.5);
  }

  .rank-badge.normal {
    background: var(--bg-secondary);
    color: var(--text-secondary);
    border: 1px solid var(--border-color);
  }

  .col-player {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .rank-icon {
    font-size: 18px;
  }

  .player-name {
    font-size: 14px;
    color: var(--text-primary);
  }

  .elo-value {
    font-size: 14px;
    font-weight: bold;
    color: var(--neon-cyan);
  }

  .winrate {
    font-size: 13px;
    color: var(--neon-green);
  }

  .my-rank-divider {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 12px 0 8px 0;
    color: var(--text-secondary);
    font-size: 12px;
  }

  .my-rank-divider::before,
  .my-rank-divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: var(--border-color);
  }

  .modal-footer {
    padding: 16px 24px;
    border-top: 1px solid var(--border-color);
    display: flex;
    justify-content: center;
  }

  .modal-footer .btn-neon {
    min-width: 120px;
  }
</style>
