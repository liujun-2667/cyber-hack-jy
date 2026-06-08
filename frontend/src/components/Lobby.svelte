<script>
  import { gameStore } from '../store/gameStore.js'
  import { onMount } from 'svelte'

  export let onLogin

  let username = ''
  let roomId = ''
  let isMatching = false
  let matchStatus = ''

  function quickMatch() {
    if (!username.trim()) {
      alert('请输入用户名')
      return
    }
    onLogin(username.trim())
    isMatching = true
    matchStatus = '正在匹配中...'
    
    setTimeout(() => {
      gameStore.quickMatch()
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

  $: if ($gameStore.gameState && $gameStore.gameState.phase) {
    isMatching = false
  }
</script>

<div class="lobby-container">
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
        />
      </div>
    </div>

    <div class="match-options">
      <button class="btn-neon match-btn" on:click={quickMatch}>
        <span class="btn-icon">⚡</span>
        快速匹配
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
        <p>{matchStatus}</p>
        <button class="btn-neon" on:click={() => { isMatching = false; gameStore.cancelMatch() }}>
          取消匹配
        </button>
      </div>
    </div>
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
    background: rgba(0, 0, 0, 0.8);
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
</style>
