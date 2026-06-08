<script>
  import { gameStore, myPlayer, opponents } from '../store/gameStore.js'
  import { onMount, onDestroy } from 'svelte'
  import IsometricGrid from './IsometricGrid.svelte'
  import HandCards from './HandCards.svelte'
  import GameLog from './GameLog.svelte'
  import CardDetail from './CardDetail.svelte'
  import GameInfo from './GameInfo.svelte'

  export let onBack
  export let onWatchReplay = null

  let selectedCard = null
  let targetMode = false
  let selectedTarget = null
  let timer = 20
  let timerInterval = null

  $: gameState = $gameStore.gameState
  $: phase = gameState?.phase
  $: currentTurn = gameState?.currentTurn

  $: if (phase === 'programming' && timer > 0) {
    startTimer()
  }

  function startTimer() {
    if (timerInterval) return
    timer = 20
    timerInterval = setInterval(() => {
      timer -= 1
      if (timer <= 0) {
        clearInterval(timerInterval)
        timerInterval = null
      }
    }, 1000)
  }

  function handleCardSelect(card) {
    if (phase !== 'programming') return
    selectedCard = card
    targetMode = true
  }

  function handleNodeClick(node, ownerId, isOpponent) {
    if (!targetMode || !selectedCard) return
    
    if (selectedCard.targetSelf && isOpponent) return
    if (!selectedCard.targetSelf && !isOpponent) return

    gameStore.playCard(selectedCard.id, node.id, ownerId)
    selectedCard = null
    targetMode = false
  }

  function handleEndTurn() {
    gameStore.endTurn()
    selectedCard = null
    targetMode = false
  }

  function handleCancelTarget() {
    selectedCard = null
    targetMode = false
  }

  async function handleWatchReplay() {
    if (!$gameStore.lastReplayId || !$gameStore.roomId) return
    await gameStore.fetchReplay($gameStore.roomId, $gameStore.lastReplayId)
    if (onWatchReplay) {
      onWatchReplay()
    }
  }

  onMount(() => {
    gameStore.requestGameState()
  })

  onDestroy(() => {
    if (timerInterval) {
      clearInterval(timerInterval)
    }
  })
</script>

<div class="game-board">
  <div class="top-bar">
    <button class="back-btn" on:click={onBack}>← 返回大厅</button>
    <div class="turn-info">
      <span class="turn-label">回合</span>
      <span class="turn-number">{currentTurn || 1}/30</span>
    </div>
    <div class="phase-info">
      {#if phase === 'programming'}
        <span class="phase-name neon-text-cyan">编程阶段</span>
        <span class="phase-timer">{timer}s</span>
      {:else if phase === 'execution'}
        <span class="phase-name neon-text-pink">执行阶段</span>
      {:else if phase === 'gameover'}
        <span class="phase-name neon-text-green">游戏结束</span>
      {:else if phase === 'placement'}
        <span class="phase-name neon-text-yellow">布阵阶段</span>
      {/if}
    </div>
    <div class="player-status">
      <span class="status-dot"></span>
      <span>{$myPlayer?.username || '玩家'}</span>
    </div>
  </div>

  <div class="main-content">
    <div class="left-panel">
      <GameInfo />
    </div>

    <div class="center-area">
      <div class="opponent-area">
        {#each $opponents as opponent (opponent.id)}
          <div class="opponent-section">
            <div class="opponent-header">
              <span class="opponent-name">{opponent.username}</span>
              <span class="opponent-hp">
                核心: {opponent.coreHp}/{opponent.coreMaxHp}
              </span>
            </div>
            <IsometricGrid 
              grid={opponent.grid} 
              isOpponent={true}
              ownerId={opponent.id}
              targetMode={targetMode && !selectedCard?.targetSelf}
              onNodeClick={handleNodeClick}
            />
          </div>
        {/each}
      </div>

      <div class="vs-divider">
        <span class="vs-text">VS</span>
      </div>

      <div class="player-area">
        <div class="player-section">
          <div class="player-header">
            <span class="player-name neon-text-cyan">{$myPlayer?.username || '我'}</span>
            <span class="player-hp">
              核心: {$myPlayer?.coreHp || 30}/{$myPlayer?.coreMaxHp || 30}
            </span>
          </div>
          <IsometricGrid 
            grid={$myPlayer?.grid || []} 
            isOpponent={false}
            ownerId={$myPlayer?.id}
            targetMode={targetMode && selectedCard?.targetSelf}
            onNodeClick={handleNodeClick}
          />
        </div>
      </div>
    </div>

    <div class="right-panel">
      <GameLog />
    </div>
  </div>

  <div class="bottom-panel">
    <HandCards 
      cards={$myPlayer?.hand || []}
      playedCards={$myPlayer?.playedCards || []}
      onCardSelect={handleCardSelect}
      selectedCard={selectedCard}
      disabled={phase !== 'programming'}
    />
    <div class="action-buttons">
      {#if targetMode}
        <button class="btn-neon cancel-btn" on:click={handleCancelTarget}>取消</button>
      {/if}
      {#if phase === 'programming'}
        <button class="btn-neon-pink end-turn-btn" on:click={handleEndTurn}>
          结束回合
        </button>
      {/if}
    </div>
  </div>

  {#if selectedCard}
    <div class="card-detail-floating">
      <CardDetail card={selectedCard} />
      <p class="target-hint">
        {selectedCard.targetSelf ? '点击己方节点' : '点击敌方节点'}
      </p>
    </div>
  {/if}

  {#if phase === 'gameover'}
    <div class="game-over-overlay">
      <div class="game-over-content">
        <h2 class="neon-text-green">游戏结束</h2>
        <p class="winner-info">
          胜者: {gameState?.winnerId === $myPlayer?.id ? '你' : '对手'}
        </p>
        <div class="game-over-buttons">
          {#if $gameStore.lastReplayId}
            <button class="btn-neon-pink" on:click={handleWatchReplay}>
              观看回放
            </button>
          {/if}
          <button class="btn-neon" on:click={onBack}>返回大厅</button>
        </div>
      </div>
    </div>
  {/if}

  {#if $gameStore.error}
    <div class="error-toast">
      {$gameStore.error}
    </div>
  {/if}
</div>

<style>
  .game-board {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    position: relative;
    z-index: 1;
  }

  .top-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px;
    background: var(--bg-secondary);
    border-bottom: 1px solid var(--border-color);
  }

  .back-btn {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-secondary);
    padding: 8px 16px;
    cursor: pointer;
    transition: all 0.3s;
  }

  .back-btn:hover {
    border-color: var(--neon-cyan);
    color: var(--neon-cyan);
  }

  .turn-info {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .turn-label {
    font-size: 11px;
    color: var(--text-secondary);
    letter-spacing: 2px;
  }

  .turn-number {
    font-size: 20px;
    font-weight: bold;
    color: var(--neon-cyan);
  }

  .phase-info {
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .phase-name {
    font-size: 16px;
    font-weight: bold;
    letter-spacing: 2px;
  }

  .phase-timer {
    font-size: 24px;
    font-weight: bold;
    color: var(--neon-yellow);
  }

  .player-status {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .status-dot {
    width: 10px;
    height: 10px;
    background: var(--neon-green);
    border-radius: 50%;
    box-shadow: 0 0 10px var(--neon-green);
  }

  .main-content {
    flex: 1;
    display: flex;
    overflow: hidden;
  }

  .left-panel,
  .right-panel {
    width: 250px;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-color);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
  }

  .center-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 20px;
    gap: 10px;
  }

  .opponent-area,
  .player-area {
    width: 100%;
    display: flex;
    justify-content: center;
  }

  .opponent-section,
  .player-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .opponent-header,
  .player-header {
    display: flex;
    justify-content: space-between;
    width: 400px;
    padding: 6px 12px;
    background: rgba(0, 0, 0, 0.3);
    border-radius: 4px;
  }

  .opponent-name,
  .player-name {
    font-weight: bold;
  }

  .opponent-hp,
  .player-hp {
    color: var(--neon-red);
    font-size: 13px;
  }

  .vs-divider {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .vs-text {
    font-size: 20px;
    font-weight: bold;
    color: var(--neon-pink);
    text-shadow: 0 0 10px var(--neon-pink);
  }

  .bottom-panel {
    background: var(--bg-secondary);
    border-top: 1px solid var(--border-color);
    padding: 15px 20px;
    display: flex;
    align-items: flex-end;
    gap: 20px;
  }

  .action-buttons {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .end-turn-btn {
    padding: 12px 24px;
    font-size: 14px;
  }

  .cancel-btn {
    padding: 8px 16px;
    font-size: 12px;
  }

  .card-detail-floating {
    position: fixed;
    top: 100px;
    right: 280px;
    z-index: 10;
  }

  .target-hint {
    margin-top: 10px;
    text-align: center;
    font-size: 12px;
    color: var(--neon-yellow);
    animation: pulse 1s infinite;
  }

  .game-over-overlay {
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

  .game-over-content {
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 20px;
  }

  .game-over-content h2 {
    font-size: 48px;
    letter-spacing: 4px;
  }

  .winner-info {
    font-size: 24px;
    color: var(--text-primary);
  }

  .game-over-buttons {
    display: flex;
    flex-direction: column;
    gap: 12px;
    margin-top: 10px;
  }

  .game-over-buttons .btn-neon,
  .game-over-buttons .btn-neon-pink {
    width: 200px;
    margin: 0 auto;
  }

  .error-toast {
    position: fixed;
    top: 80px;
    left: 50%;
    transform: translateX(-50%);
    background: rgba(255, 51, 102, 0.9);
    color: white;
    padding: 12px 24px;
    border-radius: 4px;
    z-index: 200;
    animation: slideDown 0.3s ease;
  }

  @keyframes slideDown {
    from {
      opacity: 0;
      transform: translateX(-50%) translateY(-20px);
    }
    to {
      opacity: 1;
      transform: translateX(-50%) translateY(0);
    }
  }
</style>
