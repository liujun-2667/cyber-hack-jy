<script>
  import { gameStore } from '../store/gameStore.js'
  import { onMount, onDestroy } from 'svelte'
  import PlayerRank from './PlayerRank.svelte'
  import Leaderboard from './Leaderboard.svelte'
  import { formatElo, getRankInfo } from '../utils/rank.js'

  export let onLogin

  let username = ''
  let roomId = ''
  let isMatching = false
  let matchStatus = ''
  let showLeaderboard = false
  let matchWaitTime = 0
  let matchRangeLabel = '±200'
  let matchTimerInterval = null
  let seasonInfo = null

  $: matchmakingStatus = $gameStore.matchmakingStatus
  $: playerInfo = $gameStore.playerInfo
  $: isLoggedIn = !!playerInfo

  function quickMatch() {
    if (!username.trim()) {
      alert('请输入用户名')
      return
    }
    onLogin(username.trim())
    
    setTimeout(() => {
      gameStore.quickMatch()
      startMatchTimer()
    }, 500)
  }

  function createRoom() {
    if (!username.trim()) {
      alert('请输入用户名')
      return
    }
    onLogin(username.trim())
    gameStore.createRoom()
  }

  function joinRoom() {
    if (!username.trim()) {
      alert('请输入用户名')
      return
    }
    if (!roomId.trim()) {
      alert('请输入房间号')
      return
    }
    onLogin(username.trim())
    gameStore.joinRoom(roomId.trim())
  }

  function startMatchTimer() {
    matchWaitTime = 0
    updateMatchRange()
    matchTimerInterval = setInterval(() => {
      matchWaitTime += 1
      updateMatchRange()
    }, 1000)
  }

  function updateMatchRange() {
    if (matchWaitTime < 10) {
      matchRangeLabel = '±200'
    } else if (matchWaitTime < 20) {
      matchRangeLabel = '±400'
    } else {
      matchRangeLabel = '不限'
    }
  }

  function handleCancelMatch() {
    gameStore.cancelMatch()
    stopMatchTimer()
  }

  function stopMatchTimer() {
    if (matchTimerInterval) {
      clearInterval(matchTimerInterval)
      matchTimerInterval = null
    }
  }

  function openLeaderboard() {
    showLeaderboard = true
  }

  function closeLeaderboard() {
    showLeaderboard = false
  }

  async function loadSeasonInfo() {
    try {
      seasonInfo = await gameStore.fetchSeasonInfo()
    } catch (e) {
      console.error('Failed to load season info:', e)
    }
  }

  function formatTimeRemaining(seconds) {
    if (!seconds) return ''
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    if (days > 0) {
      return `${days}天${hours}小时`
    }
    return `${hours}小时`
  }

  $: if ($gameStore.gameState && $gameStore.gameState.phase) {
    isMatching = false
    stopMatchTimer()
  }

  $: if ($gameStore.isMatching !== undefined) {
    isMatching = $gameStore.isMatching
    if (!isMatching) {
      stopMatchTimer()
    }
  }

  onMount(() => {
    loadSeasonInfo()
  })

  onDestroy(() => {
    stopMatchTimer()
  })
</script>

<div class="lobby-container">
  {#if isLoggedIn}
    <div class="lobby-top-bar">
      <div class="season-info" title="赛季剩余时间">
        <span class="season-icon">🏆</span>
        <span class="season-name">{seasonInfo?.name || '第1赛季'}</span>
        <span class="season-time">剩余 {formatTimeRemaining(seasonInfo?.timeRemaining)}</span>
      </div>
      <div class="top-bar-right">
        <button class="leaderboard-btn" on:click={openLeaderboard}>
          <span>🏆</span>
          排行榜
        </button>
        <PlayerRank />
      </div>
    </div>
  {/if}

  <div class="lobby-header">
    <h1 class="game-title">
      <span class="neon-text-cyan">CYBER</span><span class="neon-text-pink">HACK</span>
    </h1>
    <p class="subtitle">赛博朋克黑客攻防对战</p>
  </div>

  <div class="lobby-content">
    <div class="login-section">
      <div class="input-group">
        <label>黑客代号</label>
        <input 
          type="text" 
          bind:value={username} 
          placeholder="输入你的黑客代号"
          maxlength="20"
          disabled={isLoggedIn}
        />
      </div>
    </div>

    <div class="match-options">
      <button class="btn-neon match-btn" on:click={quickMatch} disabled={isMatching}>
        <span class="btn-icon">⚡</span>
        快速匹配
        {#if isLoggedIn}
          <span class="match-rank-hint">ELO积分匹配</span>
        {/if}
      </button>

      <div class="divider">
        <span>或</span>
      </div>

      <div class="room-options">
        <button class="btn-neon-pink room-btn" on:click={createRoom}>
          创建房间
        </button>
        
        <div class="join-room">
          <input 
            type="text" 
            bind:value={roomId} 
            placeholder="输入房间号"
          />
          <button class="btn-neon" on:click={joinRoom}>加入</button>
        </div>
      </div>
    </div>

    <div class="game-info">
      <div class="info-item">
        <span class="info-label">游戏模式</span>
        <span class="info-value">2-4人回合制</span>
      </div>
      <div class="info-item">
        <span class="info-label">每回合时间</span>
        <span class="info-value">20秒编程</span>
      </div>
      <div class="info-item">
        <span class="info-label">卡牌类型</span>
        <span class="info-value">15种指令卡</span>
      </div>
    </div>
  </div>

  <div class="lobby-footer">
    <p>© 2077 CyberHack Arena - 进入网络空间</p>
  </div>

  {#if isMatching}
    <div class="matching-overlay">
      <div class="matching-content">
        <div class="loading-spinner"></div>
        <h3 class="neon-text-cyan">正在匹配对手</h3>
        
        <div class="match-stats">
          <div class="match-stat-item">
            <span class="stat-label">等待时间</span>
            <span class="stat-value">{matchWaitTime}秒</span>
          </div>
          <div class="match-stat-item">
            <span class="stat-label">匹配范围</span>
            <span class="stat-value range">{matchRangeLabel}</span>
          </div>
        </div>

        <div class="match-range-hint">
          {#if matchWaitTime < 10}
            <p>正在为您寻找实力相当的对手...</p>
          {:else if matchWaitTime < 20}
            <p>匹配范围已扩大，请稍候...</p>
          {:else}
            <p>匹配范围已扩大至全服，正在为您寻找对手...</p>
          {/if}
        </div>

        <button class="btn-neon cancel-btn" on:click={handleCancelMatch}>
          取消匹配
        </button>
      </div>
    </div>
  {/if}

  {#if showLeaderboard}
    <Leaderboard onClose={closeLeaderboard} />
  {/if}
</div>

<style>
  .lobby-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    position: relative;
    z-index: 1;
  }

  .lobby-top-bar {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 24px;
    background: rgba(0, 0, 0, 0.6);
    border-bottom: 1px solid var(--border-color);
    z-index: 50;
    backdrop-filter: blur(10px);
  }

  .season-info {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    background: rgba(255, 215, 0, 0.1);
    border: 1px solid rgba(255, 215, 0, 0.3);
    border-radius: 6px;
    font-size: 13px;
  }

  .season-icon {
    font-size: 16px;
  }

  .season-name {
    color: #FFD700;
    font-weight: bold;
  }

  .season-time {
    color: var(--text-secondary);
    font-size: 12px;
  }

  .top-bar-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .leaderboard-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 8px 16px;
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    border-radius: 6px;
    cursor: pointer;
    font-size: 13px;
    transition: all 0.3s;
  }

  .leaderboard-btn:hover {
    border-color: var(--neon-pink);
    color: var(--neon-pink);
    box-shadow: 0 0 10px rgba(255, 51, 102, 0.3);
  }

  .lobby-header {
    text-align: center;
    margin-bottom: 50px;
  }

  .game-title {
    font-size: 72px;
    font-weight: 900;
    letter-spacing: 8px;
    margin-bottom: 10px;
  }

  .subtitle {
    font-size: 18px;
    color: var(--text-secondary);
    letter-spacing: 4px;
  }

  .lobby-content {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 40px;
    width: 400px;
    box-shadow: 0 0 30px rgba(0, 240, 255, 0.1);
  }

  .login-section {
    margin-bottom: 30px;
  }

  .input-group {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .input-group label {
    font-size: 12px;
    color: var(--neon-cyan);
    letter-spacing: 2px;
  }

  .input-group input {
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    padding: 12px 16px;
    font-size: 14px;
    outline: none;
    transition: border-color 0.3s;
  }

  .input-group input:focus {
    border-color: var(--neon-cyan);
    box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
  }

  .match-options {
    display: flex;
    flex-direction: column;
    gap: 15px;
  }

  .match-btn {
    width: 100%;
    padding: 16px;
    font-size: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
  }

  .btn-icon {
    font-size: 20px;
  }

  .divider {
    text-align: center;
    position: relative;
    color: var(--text-secondary);
    font-size: 12px;
  }

  .divider::before,
  .divider::after {
    content: '';
    position: absolute;
    top: 50%;
    width: 40%;
    height: 1px;
    background: var(--border-color);
  }

  .divider::before {
    left: 0;
  }

  .divider::after {
    right: 0;
  }

  .room-options {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .room-btn {
    width: 100%;
    padding: 12px;
  }

  .join-room {
    display: flex;
    gap: 10px;
  }

  .join-room input {
    flex: 1;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    padding: 10px 12px;
    font-size: 13px;
    outline: none;
  }

  .join-room input:focus {
    border-color: var(--neon-pink);
  }

  .join-room button {
    padding: 10px 20px;
  }

  .game-info {
    margin-top: 30px;
    padding-top: 20px;
    border-top: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .info-item {
    display: flex;
    justify-content: space-between;
    font-size: 13px;
  }

  .info-label {
    color: var(--text-secondary);
  }

  .info-value {
    color: var(--neon-cyan);
  }

  .lobby-footer {
    position: absolute;
    bottom: 20px;
    color: var(--text-secondary);
    font-size: 12px;
  }

  .matching-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.85);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
  }

  .matching-content {
    text-align: center;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 20px;
    background: var(--bg-secondary);
    border: 1px solid var(--neon-cyan);
    border-radius: 12px;
    padding: 40px 60px;
    box-shadow: 0 0 30px rgba(0, 240, 255, 0.2);
  }

  .matching-content h3 {
    margin: 0;
    font-size: 20px;
    letter-spacing: 2px;
  }

  .match-stats {
    display: flex;
    gap: 30px;
    margin: 10px 0;
  }

  .match-stat-item {
    display: flex;
    flex-direction: column;
    gap: 4px;
    padding: 12px 20px;
    background: var(--bg-tertiary);
    border-radius: 8px;
    min-width: 100px;
  }

  .match-stat-item .stat-label {
    font-size: 12px;
    color: var(--text-secondary);
    letter-spacing: 1px;
  }

  .match-stat-item .stat-value {
    font-size: 20px;
    font-weight: bold;
    color: var(--neon-cyan);
  }

  .match-stat-item .stat-value.range {
    color: var(--neon-green);
  }

  .match-range-hint {
    color: var(--text-secondary);
    font-size: 13px;
    min-height: 20px;
  }

  .match-range-hint p {
    margin: 0;
    animation: pulse 2s ease-in-out infinite;
  }

  @keyframes pulse {
    0%, 100% { opacity: 0.6; }
    50% { opacity: 1; }
  }

  .cancel-btn {
    min-width: 160px;
    margin-top: 10px;
  }

  .loading-spinner {
    width: 60px;
    height: 60px;
    border: 3px solid var(--border-color);
    border-top-color: var(--neon-cyan);
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .match-btn {
    position: relative;
  }

  .match-rank-hint {
    display: block;
    font-size: 11px;
    color: var(--text-secondary);
    margin-top: 4px;
    letter-spacing: 1px;
  }

  .match-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
</style>
