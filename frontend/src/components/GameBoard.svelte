<script>
  import { gameStore, myPlayer, opponents } from '../store/gameStore.js'
  import { onMount, onDestroy } from 'svelte'
  import IsometricGrid from './IsometricGrid.svelte'
  import HandCards from './HandCards.svelte'
  import GameLog from './GameLog.svelte'
  import CardDetail from './CardDetail.svelte'
  import GameInfo from './GameInfo.svelte'
  import { getRankInfo, formatElo, getRankProgress } from '../utils/rank.js'

  export let onBack
  export let onWatchReplay = null

  let selectedCard = null
  let targetMode = false
  let selectedTarget = null
  let timer = 20
  let timerInterval = null
  let showRankAnimation = false
  let animationPhase = 'result'

  $: gameState = $gameStore.gameState
  $: phase = gameState?.phase
  $: currentTurn = gameState?.currentTurn
  $: rankResults = $gameStore.rankResults
  $: myPlayerId = $myPlayer?.id
  $: myRankResult = rankResults?.[myPlayerId]
  $: isWin = myRankResult?.isWinner
  $: rankChange = myRankResult?.rankChange
  $: newRankInfo = myRankResult ? getRankInfo(myRankResult.newRank) : null
  $: oldRankInfo = myRankResult ? getRankInfo(myRankResult.oldRank) : null

  $: if (phase === 'programming' && timer > 0) {
    startTimer()
  }

  $: if (phase === 'gameover' && rankResults) {
    setTimeout(() => {
      showRankAnimation = true
      setTimeout(() => {
        animationPhase = 'rank'
      }, 800)
    }, 500)
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
      {#if showRankAnimation && rankChange === 'promote'}
        <div class="particles gold">
          {#each Array.from({ length: 30 }) as _, i}
            <span class="particle" style="--delay: {i * 0.05}s;"></span>
          {/each}
        </div>
      {:else if showRankAnimation && rankChange === 'demote'}
        <div class="particles gray">
          {#each Array.from({ length: 20 }) as _, i}
            <span class="shard" style="--delay: {i * 0.06}s;"></span>
          {/each}
        </div>
      {/if}

      <div class="game-over-content" class:show={showRankAnimation}>
        {#if rankChange === 'promote'}
          <div class="rank-change promote">
            <div class="rank-icon-large" style="color: {newRankInfo?.color};">
              {newRankInfo?.icon}
            </div>
            <div class="rank-change-text">
              <span class="rank-change-title">🎉 晋级成功!</span>
              <span class="rank-name" style="color: {newRankInfo?.color};">
                {newRankInfo?.name}
              </span>
            </div>
          </div>
        {:else if rankChange === 'demote'}
          <div class="rank-change demote">
            <div class="rank-icon-large shrunk" style="color: {newRankInfo?.color};">
              {newRankInfo?.icon}
            </div>
            <div class="rank-change-text">
              <span class="rank-change-title">💔 降级</span>
              <span class="rank-name" style="color: {newRankInfo?.color};">
                {newRankInfo?.name}
              </span>
            </div>
          </div>
        {/if}

        <h2 class={isWin ? 'neon-text-green' : 'neon-text-red'}>
          {isWin ? '胜利!' : '失败'}
        </h2>

        {#if myRankResult}
          <div class="elo-change-section">
            <div class="elo-row">
              <span class="elo-label">积分变动</span>
              <span class="elo-change" class:positive={myRankResult.eloChange > 0}>
                {myRankResult.eloChange > 0 ? '+' : ''}{myRankResult.eloChange}
              </span>
            </div>
            <div class="elo-row">
              <span class="elo-label">当前积分</span>
              <span class="elo-current">{formatElo(myRankResult.newElo)}</span>
            </div>
            <div class="elo-progress">
              <div class="elo-bar">
                <div 
                  class="elo-bar-fill"
                  style="width: {getRankProgress(myRankResult.newElo, myRankResult.newRank)}%; background: {newRankInfo?.color};"
                ></div>
              </div>
              <span class="elo-rank-name" style="color: {newRankInfo?.color};">
                {newRankInfo?.name}
              </span>
            </div>
          </div>
        {/if}

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
    overflow: hidden;
  }

  .particles {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    pointer-events: none;
    overflow: hidden;
  }

  .particle {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 8px;
    height: 8px;
    background: #FFD700;
    border-radius: 50%;
    box-shadow: 0 0 10px #FFD700, 0 0 20px #FFA500;
    animation: particleBurst 2s ease-out forwards;
    animation-delay: var(--delay);
  }

  @keyframes particleBurst {
    0% {
      transform: translate(-50%, -50%) scale(1);
      opacity: 1;
    }
    100% {
      transform: translate(
        calc(-50% + calc(cos(var(--angle, 0deg)) * 300px)),
        calc(-50% + calc(sin(var(--angle, 0deg)) * 300px) - 200px)
      );
      opacity: 0;
    }
  }

  .particles.gold .particle {
    background: #FFD700;
    box-shadow: 0 0 10px #FFD700, 0 0 20px #FFA500;
  }

  .particles.gray .shard {
    background: #888;
    box-shadow: 0 0 6px #666;
  }

  .shard {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 6px;
    height: 6px;
    background: #888;
    clip-path: polygon(50% 0%, 100% 100%, 0% 100%);
    animation: shardFall 2.5s ease-in forwards;
    animation-delay: var(--delay);
    opacity: 0;
  }

  @keyframes shardFall {
    0% {
      transform: translate(-50%, -50%) rotate(0deg) scale(1);
      opacity: 1;
    }
    100% {
      transform: translate(
        calc(-50% + calc(cos(var(--angle, 0deg)) * 200px)),
        calc(-50% + 300px)
      ) rotate(360deg) scale(0.5);
      opacity: 0;
    }
  }

  .game-over-content {
    text-align: center;
    display: flex;
    flex-direction: column;
    gap: 16px;
    position: relative;
    z-index: 1;
    opacity: 0;
    transform: scale(0.8);
    transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .game-over-content.show {
    opacity: 1;
    transform: scale(1);
  }

  .rank-change {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
    margin-bottom: 10px;
  }

  .rank-change.promote .rank-icon-large {
    animation: rankUp 1s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .rank-change.demote .rank-icon-large {
    animation: rankDown 1s cubic-bezier(0.34, 1.56, 0.64, 1);
  }

  .rank-icon-large {
    font-size: 80px;
    line-height: 1;
    filter: drop-shadow(0 0 20px currentColor);
  }

  .rank-icon-large.shrunk {
    font-size: 60px;
  }

  @keyframes rankUp {
    0% {
      transform: scale(0.3);
      opacity: 0;
    }
    50% {
      transform: scale(1.3);
    }
    100% {
      transform: scale(1);
      opacity: 1;
    }
  }

  @keyframes rankDown {
    0% {
      transform: scale(1.5);
      opacity: 0;
    }
    50% {
      transform: scale(0.8);
    }
    100% {
      transform: scale(1);
      opacity: 1;
    }
  }

  .rank-change-text {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .rank-change-title {
    font-size: 24px;
    color: var(--text-primary);
    font-weight: bold;
  }

  .rank-name {
    font-size: 18px;
    font-weight: bold;
    letter-spacing: 2px;
  }

  .game-over-content h2 {
    font-size: 48px;
    letter-spacing: 4px;
    margin: 0;
  }

  .winner-info {
    font-size: 24px;
    color: var(--text-primary);
  }

  .elo-change-section {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 16px 24px;
    min-width: 280px;
  }

  .elo-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .elo-label {
    font-size: 13px;
    color: var(--text-secondary);
  }

  .elo-change {
    font-size: 20px;
    font-weight: bold;
    color: var(--neon-red);
  }

  .elo-change.positive {
    color: var(--neon-green);
  }

  .elo-current {
    font-size: 18px;
    font-weight: bold;
    color: var(--neon-cyan);
  }

  .elo-progress {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .elo-bar {
    width: 100%;
    height: 6px;
    background: var(--bg-tertiary);
    border-radius: 3px;
    overflow: hidden;
  }

  .elo-bar-fill {
    height: 100%;
    border-radius: 3px;
    transition: width 1s ease;
    box-shadow: 0 0 8px currentColor;
  }

  .elo-rank-name {
    text-align: right;
    font-size: 12px;
    font-weight: bold;
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
