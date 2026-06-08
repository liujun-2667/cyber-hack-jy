<script>
  import { gameStore } from '../store/gameStore.js'
  import { onMount, onDestroy } from 'svelte'
  import { getRankName, getRankIcon, getRankColor, formatElo } from '../utils/rank.js'

  export let onClose

  let showCreateForm = false
  let tournamentName = ''
  let selectedMaxPlayers = 8
  let selectedDuration = 5
  let selectedMinRank = 'none'
  let countdownTimers = {}
  let timerInterval = null

  $: tournamentList = $gameStore.tournamentList || []
  $: playerInfo = $gameStore.playerInfo
  $: isLoggedIn = !!playerInfo

  const maxPlayersOptions = [
    { value: 8, label: '8人' },
    { value: 16, label: '16人' },
    { value: 32, label: '32人' }
  ]

  const durationOptions = [
    { value: 5, label: '5分钟' },
    { value: 10, label: '10分钟' },
    { value: 15, label: '15分钟' }
  ]

  const minRankOptions = [
    { value: 'none', label: '无限制' },
    { value: 'silver', label: '白银以上' },
    { value: 'gold', label: '黄金以上' }
  ]

  function formatTimeRemaining(deadline) {
    if (!deadline) return '计算中...'
    const now = new Date().getTime()
    const deadlineTime = new Date(deadline).getTime()
    const diff = Math.max(0, Math.floor((deadlineTime - now) / 1000))
    
    if (diff <= 0) return '即将开始'
    
    const minutes = Math.floor(diff / 60)
    const seconds = diff % 60
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }

  function getStatusText(status) {
    const statusMap = {
      'registering': '报名中',
      'in_progress': '进行中',
      'completed': '已结束',
      'cancelled': '已取消'
    }
    return statusMap[status] || status
  }

  function getStatusClass(status) {
    const classMap = {
      'registering': 'status-registering',
      'in_progress': 'status-progress',
      'completed': 'status-completed',
      'cancelled': 'status-cancelled'
    }
    return classMap[status] || ''
  }

  function getMinRankDisplay(minRank) {
    if (minRank === 'none' || !minRank) return '无限制'
    return getRankName(minRank) + '以上'
  }

  function canRegister(tournament) {
    if (tournament.status !== 'registering') return false
    if (tournament.playerCount >= tournament.maxPlayers) return false
    
    if (tournament.minRank === 'none' || !tournament.minRank) return true
    if (!playerInfo || !playerInfo.eloRating) return false
    
    const rankThresholds = {
      silver: 1100,
      gold: 1400,
      platinum: 1700,
      diamond: 2000
    }
    
    return playerInfo.eloRating >= (rankThresholds[tournament.minRank] || 0)
  }

  function handleCreateTournament() {
    if (!tournamentName.trim()) {
      alert('请输入锦标赛名称')
      return
    }
    if (!isLoggedIn) {
      alert('请先登录')
      return
    }
    gameStore.createTournament(
      tournamentName.trim(),
      selectedMaxPlayers,
      selectedMinRank,
      selectedDuration
    )
    showCreateForm = false
    tournamentName = ''
  }

  function handleJoin(tournamentId) {
    if (!isLoggedIn) {
      alert('请先登录')
      return
    }
    gameStore.joinTournament(tournamentId)
  }

  function handleWatch(tournamentId) {
    gameStore.watchTournament(tournamentId)
  }

  function handleLeave(tournamentId) {
    gameStore.leaveTournament(tournamentId)
  }

  async function loadTournaments() {
    await gameStore.fetchTournaments()
  }

  function startTimers() {
    timerInterval = setInterval(() => {
      tournamentList.forEach(t => {
        countdownTimers[t.id] = formatTimeRemaining(t.registrationDeadline)
      })
    }, 1000)
  }

  function stopTimers() {
    if (timerInterval) {
      clearInterval(timerInterval)
      timerInterval = null
    }
  }

  onMount(() => {
    loadTournaments()
    gameStore.requestTournamentList()
    startTimers()
  })

  onDestroy(() => {
    stopTimers()
  })
</script>

<div class="tournament-overlay">
  <div class="tournament-panel">
    <div class="panel-header">
      <h2 class="panel-title">
        <span class="title-icon">🏆</span>
        锦标赛大厅
      </h2>
      <button class="close-btn" on:click={onClose}>✕</button>
    </div>

    <div class="panel-actions">
      <button class="btn-neon create-btn" on:click={() => showCreateForm = !showCreateForm}>
        <span>➕</span>
        {showCreateForm ? '取消创建' : '创建锦标赛'}
      </button>
      <button class="btn-outline refresh-btn" on:click={loadTournaments}>
        🔄 刷新
      </button>
    </div>

    {#if showCreateForm}
      <div class="create-form">
        <h3 class="form-title">创建锦标赛</h3>
        
        <div class="form-group">
          <label>锦标赛名称</label>
          <input 
            type="text" 
            bind:value={tournamentName} 
            placeholder="输入锦标赛名称"
            maxlength="30"
          />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>参赛人数</label>
            <select bind:value={selectedMaxPlayers}>
              {#each maxPlayersOptions as opt}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>
          </div>

          <div class="form-group">
            <label>报名时长</label>
            <select bind:value={selectedDuration}>
              {#each durationOptions as opt}
                <option value={opt.value}>{opt.label}</option>
              {/each}
            </select>
          </div>
        </div>

        <div class="form-group">
          <label>最低段位要求</label>
          <select bind:value={selectedMinRank}>
            {#each minRankOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>

        <button class="btn-neon submit-btn" on:click={handleCreateTournament}>
          确认创建
        </button>
      </div>
    {/if}

    <div class="tournament-list">
      {#if tournamentList.length === 0}
        <div class="empty-state">
          <span class="empty-icon">🏆</span>
          <p>暂无锦标赛</p>
          <p class="empty-hint">点击上方按钮创建第一个锦标赛</p>
        </div>
      {:else}
        {#each tournamentList as tournament (tournament.id)}
          <div class="tournament-card" class:is-active={tournament.status === 'in_progress'}>
            <div class="card-header">
              <h4 class="tournament-name">{tournament.name}</h4>
              <span class="status-badge {getStatusClass(tournament.status)}">
                {getStatusText(tournament.status)}
              </span>
            </div>

            <div class="card-info">
              <div class="info-item">
                <span class="info-label">参赛人数</span>
                <span class="info-value">
                  {tournament.playerCount || 0} / {tournament.maxPlayers}
                </span>
              </div>
              
              {#if tournament.status === 'registering'}
                <div class="info-item">
                  <span class="info-label">剩余时间</span>
                  <span class="info-value countdown">
                    {countdownTimers[tournament.id] || formatTimeRemaining(tournament.registrationDeadline)}
                  </span>
                </div>
              {/if}

              <div class="info-item">
                <span class="info-label">段位要求</span>
                <span class="info-value rank-requirement">
                  {getMinRankDisplay(tournament.minRank)}
                </span>
              </div>
            </div>

            <div class="card-actions">
              {#if tournament.status === 'registering'}
                {#if canRegister(tournament)}
                  <button 
                    class="btn-neon btn-sm" 
                    on:click={() => handleJoin(tournament.id)}
                  >
                    报名参赛
                  </button>
                {:else}
                  <button class="btn-disabled btn-sm" disabled title="段位不足或人数已满">
                    无法报名
                  </button>
                {/if}
              {/if}

              {#if tournament.status === 'in_progress' || tournament.status === 'completed'}
                <button 
                  class="btn-outline btn-sm" 
                  on:click={() => handleWatch(tournament.id)}
                >
                  {tournament.status === 'in_progress' ? '观战' : '查看结果'}
                </button>
              {/if}
            </div>
          </div>
        {/each}
      {/if}
    </div>
  </div>
</div>

<style>
  .tournament-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.85);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    backdrop-filter: blur(5px);
  }

  .tournament-panel {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    width: 520px;
    max-height: 80vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 0 40px rgba(0, 240, 255, 0.15);
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-color);
  }

  .panel-title {
    font-size: 20px;
    margin: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    color: var(--text-primary);
  }

  .title-icon {
    font-size: 24px;
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

  .panel-actions {
    display: flex;
    gap: 12px;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-color);
  }

  .create-btn {
    flex: 1;
    padding: 10px;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
  }

  .refresh-btn {
    padding: 10px 16px;
    font-size: 14px;
  }

  .btn-outline {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    border-radius: 6px;
    cursor: pointer;
    transition: all 0.3s;
  }

  .btn-outline:hover {
    border-color: var(--neon-cyan);
    color: var(--neon-cyan);
  }

  .create-form {
    padding: 16px 20px;
    border-bottom: 1px solid var(--border-color);
    background: rgba(0, 240, 255, 0.03);
  }

  .form-title {
    font-size: 16px;
    color: var(--neon-cyan);
    margin: 0 0 16px 0;
  }

  .form-group {
    margin-bottom: 14px;
  }

  .form-group label {
    display: block;
    font-size: 12px;
    color: var(--neon-cyan);
    margin-bottom: 6px;
    letter-spacing: 1px;
  }

  .form-group input,
  .form-group select {
    width: 100%;
    padding: 10px 12px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    border-radius: 6px;
    font-size: 14px;
    outline: none;
    box-sizing: border-box;
    transition: border-color 0.3s;
  }

  .form-group input:focus,
  .form-group select:focus {
    border-color: var(--neon-cyan);
    box-shadow: 0 0 8px rgba(0, 240, 255, 0.2);
  }

  .form-row {
    display: flex;
    gap: 12px;
  }

  .form-row .form-group {
    flex: 1;
  }

  .submit-btn {
    width: 100%;
    padding: 12px;
    font-size: 14px;
    margin-top: 8px;
  }

  .tournament-list {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .empty-state {
    text-align: center;
    padding: 40px 20px;
    color: var(--text-secondary);
  }

  .empty-icon {
    font-size: 48px;
    opacity: 0.3;
    display: block;
    margin-bottom: 12px;
  }

  .empty-hint {
    font-size: 13px;
    opacity: 0.6;
    margin-top: 4px;
  }

  .tournament-card {
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 14px 16px;
    transition: all 0.3s;
  }

  .tournament-card:hover {
    border-color: rgba(0, 240, 255, 0.3);
  }

  .tournament-card.is-active {
    border-color: rgba(255, 215, 0, 0.5);
    box-shadow: 0 0 15px rgba(255, 215, 0, 0.1);
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
  }

  .tournament-name {
    font-size: 16px;
    font-weight: 600;
    margin: 0;
    color: var(--text-primary);
  }

  .status-badge {
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
  }

  .status-registering {
    background: rgba(0, 255, 100, 0.15);
    color: #00ff64;
    border: 1px solid rgba(0, 255, 100, 0.3);
  }

  .status-progress {
    background: rgba(255, 215, 0, 0.15);
    color: #FFD700;
    border: 1px solid rgba(255, 215, 0, 0.3);
  }

  .status-completed {
    background: rgba(0, 240, 255, 0.15);
    color: var(--neon-cyan);
    border: 1px solid rgba(0, 240, 255, 0.3);
  }

  .status-cancelled {
    background: rgba(255, 51, 102, 0.15);
    color: var(--neon-pink);
    border: 1px solid rgba(255, 51, 102, 0.3);
  }

  .card-info {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 8px;
    margin-bottom: 14px;
  }

  .info-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .info-label {
    font-size: 11px;
    color: var(--text-secondary);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .info-value {
    font-size: 14px;
    color: var(--text-primary);
    font-weight: 500;
  }

  .countdown {
    color: #FFD700;
    font-family: monospace;
    font-size: 15px;
  }

  .rank-requirement {
    font-size: 13px;
  }

  .card-actions {
    display: flex;
    gap: 10px;
  }

  .btn-sm {
    padding: 8px 16px;
    font-size: 13px;
    flex: 1;
  }

  .btn-neon {
    background: transparent;
    border: 1px solid var(--neon-cyan);
    color: var(--neon-cyan);
    padding: 10px 20px;
    font-size: 14px;
    cursor: pointer;
    border-radius: 6px;
    transition: all 0.3s;
    text-shadow: 0 0 10px rgba(0, 240, 255, 0.5);
    box-shadow: 0 0 10px rgba(0, 240, 255, 0.2);
  }

  .btn-neon:hover {
    background: rgba(0, 240, 255, 0.1);
    box-shadow: 0 0 20px rgba(0, 240, 255, 0.4);
  }

  .btn-disabled {
    background: rgba(255, 255, 255, 0.05);
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    padding: 8px 16px;
    font-size: 13px;
    border-radius: 6px;
    cursor: not-allowed;
    opacity: 0.5;
  }
</style>
