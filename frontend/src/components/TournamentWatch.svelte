<script>
  import { gameStore } from '../store/gameStore.js'
  import { onMount, onDestroy } from 'svelte'
  import TournamentBracket from './TournamentBracket.svelte'
  import { getRankIcon, getRankColor, formatElo } from '../utils/rank.js'

  export let onClose

  let chatMessage = ''
  let chatContainer = null
  let playerInfo = null

  $: currentTournament = $gameStore.currentTournament
  $: currentBracket = $gameStore.currentBracket || []
  $: tournamentChat = $gameStore.tournamentChat || []
  $: playerInfo = $gameStore.playerInfo

  $: tournamentPlayers = currentTournament?.players || []
  $: isCreator = currentTournament && playerInfo && currentTournament.creatorId === playerInfo.playerId
  $: isRegistering = currentTournament?.status === 'registering'

  function handleSendChat() {
    if (!chatMessage.trim()) return
    if (!currentTournament?.id) return
    
    gameStore.sendTournamentChat(currentTournament.id, chatMessage.trim())
    chatMessage = ''
  }

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendChat()
    }
  }

  function handleKickPlayer(playerId) {
    if (!confirm('确定要踢出该玩家吗？')) return
    gameStore.kickPlayer(currentTournament.id, playerId)
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

  function formatTime(timestamp) {
    if (!timestamp) return ''
    const date = new Date(timestamp)
    return date.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
  }

  function handleClose() {
    if (currentTournament?.id) {
      gameStore.unwatchTournament(currentTournament.id)
    }
    onClose()
  }

  $: if (chatContainer) {
    chatContainer.scrollTop = chatContainer.scrollHeight
  }

  onDestroy(() => {
    if (currentTournament?.id) {
      gameStore.unwatchTournament(currentTournament.id)
    }
  })
</script>

<div class="watch-overlay">
  <div class="watch-panel">
    <div class="panel-header">
      <div class="tournament-info">
        <h2 class="tournament-name">{currentTournament?.name || '锦标赛'}</h2>
        <span class="status-badge {getStatusClass(currentTournament?.status)}">
          {getStatusText(currentTournament?.status)}
        </span>
        {#if isCreator && isRegistering}
          <span class="creator-badge">创建者</span>
        {/if}
      </div>
      <button class="close-btn" on:click={handleClose}>✕</button>
    </div>

    <div class="panel-subheader">
      <div class="info-item">
        <span class="info-label">参赛人数</span>
        <span class="info-value">{currentTournament?.playerCount || 0} / {currentTournament?.maxPlayers || 0}</span>
      </div>
      <div class="info-item">
        <span class="info-label">当前轮次</span>
        <span class="info-value">第 {currentTournament?.currentRound || 0} / {currentTournament?.totalRounds || 0} 轮</span>
      </div>
      <div class="info-item">
        <span class="info-label">观战人数</span>
        <span class="info-value">{currentTournament?.viewerCount || 0}</span>
      </div>
    </div>

    <div class="panel-content">
      <div class="left-section">
        <div class="players-section">
          <div class="section-title">
            <span>👥</span>
            参赛玩家 ({tournamentPlayers.length})
          </div>
          <div class="players-list">
            {#if tournamentPlayers.length === 0}
              <div class="players-empty">暂无玩家报名</div>
            {:else}
              {#each tournamentPlayers as player (player.playerId || player.id)}
                <div class="player-item">
                  <span class="player-seed">#{player.seed || '-'}</span>
                  <span class="player-rank-icon" style="color: {getRankColor(player.currentRank || 'bronze')}">
                    {getRankIcon(player.currentRank || 'bronze')}
                  </span>
                  <span class="player-username">{player.username}</span>
                  <span class="player-elo">{formatElo(player.eloRating || 1200)}</span>
                  {#if isCreator && isRegistering && player.playerId !== playerInfo?.playerId}
                    <button 
                      class="kick-btn" 
                      on:click={() => handleKickPlayer(player.playerId || player.id)}
                      title="踢出玩家"
                    >
                      ✕
                    </button>
                  {/if}
                </div>
              {/each}
            {/if}
          </div>
        </div>

        <div class="bracket-section">
          <div class="section-title">
            <span>🏆</span>
            对阵表
          </div>
          <div class="bracket-wrapper">
            <TournamentBracket 
              matches={currentBracket} 
              currentRound={currentTournament?.currentRound || 0}
            />
          </div>
        </div>
      </div>

      <div class="chat-section">
        <div class="section-title">
          <span>💬</span>
          赛事聊天
        </div>
        
        <div class="chat-messages" bind:this={chatContainer}>
          {#if tournamentChat.length === 0}
            <div class="chat-empty">
              暂无聊天记录
            </div>
          {:else}
            {#each tournamentChat as msg (msg.id)}
              <div class="chat-message" class:is-system={msg.isSystem}>
                {#if msg.isSystem}
                  <div class="system-message">
                    {msg.message}
                  </div>
                {:else}
                  <div class="message-header">
                    <span class="player-rank">{getRankIcon(msg.rank || 'bronze')}</span>
                    <span class="player-name">{msg.username || '匿名'}</span>
                    <span class="message-time">{formatTime(msg.timestamp)}</span>
                  </div>
                  <div class="message-content">
                    {msg.message}
                  </div>
                {/if}
              </div>
            {/each}
          {/if}
        </div>

        <div class="chat-input-wrapper">
          <input
            type="text"
            bind:value={chatMessage}
            placeholder="发送消息..."
            on:keydown={handleKeyDown}
            disabled={!playerInfo}
            maxlength="200"
          />
          <button 
            class="send-btn" 
            on:click={handleSendChat}
            disabled={!chatMessage.trim() || !playerInfo}
          >
            发送
          </button>
        </div>
      </div>
    </div>
  </div>
</div>

<style>
  .watch-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.9);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    backdrop-filter: blur(5px);
  }

  .watch-panel {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    width: 95%;
    max-width: 1100px;
    height: 85vh;
    display: flex;
    flex-direction: column;
    box-shadow: 0 0 50px rgba(0, 240, 255, 0.15);
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 16px 24px;
    border-bottom: 1px solid var(--border-color);
  }

  .tournament-info {
    display: flex;
    align-items: center;
    gap: 14px;
  }

  .tournament-name {
    font-size: 22px;
    margin: 0;
    color: var(--text-primary);
  }

  .status-badge {
    padding: 4px 12px;
    border-radius: 14px;
    font-size: 12px;
    font-weight: 600;
  }

  .creator-badge {
    padding: 4px 10px;
    border-radius: 10px;
    font-size: 11px;
    font-weight: 600;
    background: rgba(255, 215, 0, 0.15);
    color: #FFD700;
    border: 1px solid rgba(255, 215, 0, 0.3);
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
    animation: pulse-gold 2s infinite;
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

  @keyframes pulse-gold {
    0%, 100% { box-shadow: 0 0 5px rgba(255, 215, 0, 0.3); }
    50% { box-shadow: 0 0 15px rgba(255, 215, 0, 0.5); }
  }

  .close-btn {
    background: transparent;
    border: none;
    color: var(--text-secondary);
    font-size: 24px;
    cursor: pointer;
    padding: 4px 10px;
    transition: color 0.3s;
  }

  .close-btn:hover {
    color: var(--neon-pink);
  }

  .panel-subheader {
    display: flex;
    gap: 30px;
    padding: 12px 24px;
    border-bottom: 1px solid var(--border-color);
    background: rgba(0, 0, 0, 0.2);
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
    font-size: 15px;
    color: var(--text-primary);
    font-weight: 600;
  }

  .panel-content {
    flex: 1;
    display: flex;
    gap: 0;
    overflow: hidden;
  }

  .left-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border-color);
    overflow: hidden;
  }

  .players-section {
    border-bottom: 1px solid var(--border-color);
    max-height: 200px;
    display: flex;
    flex-direction: column;
  }

  .players-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .players-empty {
    text-align: center;
    color: var(--text-secondary);
    font-size: 12px;
    padding: 16px;
    font-style: italic;
  }

  .player-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    background: var(--bg-tertiary);
    border-radius: 6px;
    font-size: 13px;
  }

  .player-seed {
    font-size: 10px;
    color: var(--text-secondary);
    font-weight: bold;
    min-width: 20px;
  }

  .player-rank-icon {
    font-size: 14px;
  }

  .player-username {
    flex: 1;
    color: var(--text-primary);
    font-weight: 500;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .player-elo {
    font-size: 11px;
    color: var(--neon-cyan);
    font-family: monospace;
  }

  .kick-btn {
    background: transparent;
    border: 1px solid rgba(255, 51, 102, 0.3);
    color: var(--neon-pink);
    width: 22px;
    height: 22px;
    border-radius: 50%;
    font-size: 11px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    flex-shrink: 0;
  }

  .kick-btn:hover {
    background: rgba(255, 51, 102, 0.2);
    border-color: var(--neon-pink);
  }

  .bracket-section {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 12px 16px;
    font-size: 14px;
    font-weight: 600;
    color: var(--neon-cyan);
    border-bottom: 1px solid var(--border-color);
    background: rgba(0, 240, 255, 0.03);
  }

  .bracket-wrapper {
    flex: 1;
    overflow: auto;
    padding: 16px;
  }

  .chat-section {
    width: 300px;
    display: flex;
    flex-direction: column;
    background: rgba(0, 0, 0, 0.2);
  }

  .chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .chat-empty {
    text-align: center;
    color: var(--text-secondary);
    font-size: 13px;
    padding: 30px 0;
    font-style: italic;
  }

  .chat-message.is-system {
    text-align: center;
  }

  .system-message {
    display: inline-block;
    padding: 4px 12px;
    background: rgba(255, 215, 0, 0.1);
    color: #FFD700;
    font-size: 11px;
    border-radius: 12px;
    font-style: italic;
  }

  .message-header {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 4px;
  }

  .player-rank {
    font-size: 14px;
  }

  .player-name {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-primary);
  }

  .message-time {
    font-size: 10px;
    color: var(--text-secondary);
    margin-left: auto;
  }

  .message-content {
    font-size: 13px;
    color: var(--text-secondary);
    line-height: 1.4;
    word-break: break-word;
  }

  .chat-input-wrapper {
    display: flex;
    gap: 8px;
    padding: 12px;
    border-top: 1px solid var(--border-color);
  }

  .chat-input-wrapper input {
    flex: 1;
    padding: 8px 12px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    border-radius: 6px;
    font-size: 13px;
    outline: none;
    transition: border-color 0.3s;
  }

  .chat-input-wrapper input:focus {
    border-color: var(--neon-cyan);
  }

  .chat-input-wrapper input:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .send-btn {
    padding: 8px 16px;
    background: transparent;
    border: 1px solid var(--neon-cyan);
    color: var(--neon-cyan);
    border-radius: 6px;
    font-size: 13px;
    cursor: pointer;
    transition: all 0.3s;
  }

  .send-btn:hover:not(:disabled) {
    background: rgba(0, 240, 255, 0.1);
  }

  .send-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
</style>
