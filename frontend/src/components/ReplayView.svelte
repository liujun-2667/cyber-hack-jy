<script>
  import { onMount, onDestroy } from 'svelte'
  import { gameStore } from '../store/gameStore.js'
  import IsometricGrid from './IsometricGrid.svelte'
  import ReplayControls from './ReplayControls.svelte'
  import GameLog from './GameLog.svelte'

  export let onBack

  let playInterval = null
  let actionIndex = 0
  let actionTimer = null
  let highlightedNodeId = null
  let animationType = null
  let showingActions = false

  $: replay = $gameStore.currentReplay
  $: currentTurnIndex = $gameStore.replayTurn - 1
  $: isPlaying = $gameStore.replayPlaying
  $: speed = $gameStore.replaySpeed
  $: totalTurns = replay?.turns?.length || 0
  $: currentTurnData = replay?.turns?.[currentTurnIndex] || null
  $: currentActions = currentTurnData?.actions || []
  $: currentLog = currentTurnData?.log || []
  $: nodeStates = currentTurnData?.nodeStates || {}

  $: players = replay?.players || []
  $: player1 = players[0] || null
  $: player2 = players[1] || null

  $: player1Grid = getGridForPlayer(player1?.id)
  $: player2Grid = getGridForPlayer(player2?.id)

  $: {
    if (isPlaying && replay && totalTurns > 0) {
      restartAutoPlay()
    } else {
      stopAutoPlay()
    }
  }

  $: speed, handleSpeedChange()

  $: currentTurnIndex, resetActionAnimation()

  function getGridForPlayer(playerId) {
    if (!playerId || !nodeStates) return []

    const grid = []
    for (let x = 0; x < 5; x++) {
      grid[x] = []
      for (let y = 0; y < 5; y++) {
        grid[x][y] = null
      }
    }

    Object.values(nodeStates).forEach(node => {
      if (node.ownerId === playerId && node.x >= 0 && node.x < 5 && node.y >= 0 && node.y < 5) {
        const isHighlighted = highlightedNodeId === node.id
        grid[node.x][node.y] = {
          id: node.id,
          type: node.type,
          x: node.x,
          y: node.y,
          hp: node.hp,
          maxHp: node.maxHp,
          defense: node.defense,
          bandwidth: node.bandwidth,
          alive: node.isAlive,
          status: node.status,
          ownerId: node.ownerId,
          unknown: false,
          highlighted: isHighlighted,
          animationType: isHighlighted ? animationType : null
        }
      }
    })

    return grid
  }

  function getCoreHp(grid) {
    if (!grid || !grid[2] || !grid[2][2]) return 0
    return grid[2][2].hp || 0
  }

  function getCoreMaxHp(grid) {
    if (!grid || !grid[2] || !grid[2][2]) return 30
    return grid[2][2].maxHp || 30
  }

  function startAutoPlay() {
    if (playInterval) return
    if (currentTurnIndex >= totalTurns - 1) return

    const turnDelay = Math.max(800, 3500 / speed)
    playInterval = setInterval(() => {
      if (currentTurnIndex < totalTurns - 1) {
        gameStore.setReplayTurn($gameStore.replayTurn + 1)
      } else {
        gameStore.setReplayPlaying(false)
      }
    }, turnDelay)
  }

  function restartAutoPlay() {
    stopAutoPlay()
    startAutoPlay()
  }

  function handleSpeedChange() {
    if (isPlaying && playInterval) {
      restartAutoPlay()
    }
    if (actionTimer && showingActions) {
      const remainingActions = currentActions.length - actionIndex
      if (remainingActions > 0) {
        clearTimeout(actionTimer)
        actionTimer = null
        playNextAction()
      }
    }
  }

  function stopAutoPlay() {
    if (playInterval) {
      clearInterval(playInterval)
      playInterval = null
    }
    if (actionTimer) {
      clearTimeout(actionTimer)
      actionTimer = null
    }
  }

  function resetActionAnimation() {
    if (actionTimer) {
      clearTimeout(actionTimer)
      actionTimer = null
    }
    actionIndex = 0
    highlightedNodeId = null
    animationType = null
    showingActions = false

    if (isPlaying && currentActions.length > 0) {
      playActionAnimation()
    }
  }

  function playActionAnimation() {
    if (!currentActions || currentActions.length === 0) return

    showingActions = true
    actionIndex = 0
    playNextAction()
  }

  function playNextAction() {
    if (actionIndex >= currentActions.length) {
      actionTimer = setTimeout(() => {
        highlightedNodeId = null
        animationType = null
        showingActions = false
      }, 500 / speed)
      return
    }

    const action = currentActions[actionIndex]
    const isAttack = action.card?.category === 'attack'
    const isDefense = action.card?.category === 'defense'

    highlightedNodeId = action.targetNodeId
    if (isAttack) {
      animationType = 'attack'
    } else if (isDefense) {
      animationType = 'shield'
    } else {
      animationType = 'highlight'
    }

    const actionDelay = Math.max(300, 800 / speed)
    actionTimer = setTimeout(() => {
      actionIndex++
      playNextAction()
    }, actionDelay)
  }

  function handleBack() {
    stopAutoPlay()
    gameStore.exitReplay()
    onBack?.()
  }

  onDestroy(() => {
    stopAutoPlay()
  })
</script>

<div class="replay-view">
  <div class="top-bar">
    <button class="back-btn" on:click={handleBack}>← 返回</button>
    <div class="title">
      <span class="title-text neon-text-cyan">战斗回放</span>
    </div>
    <div class="replay-info">
      <span>总回合: {totalTurns}</span>
    </div>
  </div>

  <div class="main-content">
    <div class="left-panel">
      <div class="panel">
        <div class="panel-header">
          <span>📊 对战信息</span>
        </div>
        <div class="panel-content">
          {#if player1}
            <div class="player-info">
              <span class="player-name">{player1.username}</span>
              <span class="player-role">玩家1</span>
            </div>
          {/if}
          {#if player2}
            <div class="player-info">
              <span class="player-name">{player2.username}</span>
              <span class="player-role">玩家2</span>
            </div>
          {/if}
          <div class="vs-divider">VS</div>
          {#if currentActions && showingActions}
            <div class="current-action">
              <div class="action-label">当前动作</div>
              <div class="action-detail">
                {#if currentActions[actionIndex]}
                  <span class="action-player">{currentActions[actionIndex].username}</span>
                  <span class="action-card">{currentActions[actionIndex].card?.name}</span>
                {/if}
              </div>
            </div>
          {/if}
        </div>
      </div>
    </div>

    <div class="center-area">
      <div class="opponent-area">
        {#if player2}
          <div class="player-section">
            <div class="player-header opponent-header">
              <span class="player-name">{player2.username}</span>
              <span class="player-hp">
                核心: {getCoreHp(player2Grid)}/{getCoreMaxHp(player2Grid)}
              </span>
            </div>
            <IsometricGrid
              grid={player2Grid}
              isOpponent={true}
              ownerId={player2.id}
              targetMode={false}
              onNodeClick={() => {}}
            />
          </div>
        {/if}
      </div>

      <div class="vs-divider-center">
        <span class="vs-text">REPLAY</span>
      </div>

      <div class="player-area">
        {#if player1}
          <div class="player-section">
            <div class="player-header">
              <span class="player-name neon-text-cyan">{player1.username}</span>
              <span class="player-hp">
                核心: {getCoreHp(player1Grid)}/{getCoreMaxHp(player1Grid)}
              </span>
            </div>
            <IsometricGrid
              grid={player1Grid}
              isOpponent={false}
              ownerId={player1.id}
              targetMode={false}
              onNodeClick={() => {}}
            />
          </div>
        {/if}
      </div>
    </div>

    <div class="right-panel">
      <GameLog logs={currentLog} />
    </div>
  </div>

  <ReplayControls
    currentTurn={$gameStore.replayTurn || 1}
    totalTurns={totalTurns || 1}
    isPlaying={isPlaying}
    speed={speed}
  />
</div>

<style>
  .replay-view {
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

  .title {
    text-align: center;
  }

  .title-text {
    font-size: 18px;
    font-weight: bold;
    letter-spacing: 3px;
  }

  .replay-info {
    font-size: 13px;
    color: var(--text-secondary);
    min-width: 100px;
    text-align: right;
  }

  .main-content {
    flex: 1;
    display: flex;
    overflow: hidden;
    min-height: 0;
  }

  .left-panel,
  .right-panel {
    width: 250px;
    background: var(--bg-secondary);
    border-left: 1px solid var(--border-color);
    border-right: 1px solid var(--border-color);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .center-area {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: space-around;
    padding: 15px 20px;
    gap: 20px;
    overflow: auto;
    min-height: 0;
  }

  .opponent-area,
  .player-area {
    width: 100%;
    display: flex;
    justify-content: center;
    flex-shrink: 0;
  }

  .player-section {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .player-header,
  .opponent-header {
    display: flex;
    justify-content: space-between;
    width: 400px;
    padding: 6px 12px;
    background: rgba(0, 0, 0, 0.3);
    border-radius: 4px;
  }

  .player-name {
    font-weight: bold;
  }

  .player-hp {
    color: var(--neon-red);
    font-size: 13px;
  }

  .vs-divider-center {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .vs-text {
    font-size: 16px;
    font-weight: bold;
    color: var(--neon-pink);
    text-shadow: 0 0 10px var(--neon-pink);
    letter-spacing: 2px;
  }

  .panel {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    flex: 1;
    display: flex;
    flex-direction: column;
  }

  .panel-header {
    background: var(--bg-tertiary);
    padding: 10px 15px;
    border-bottom: 1px solid var(--border-color);
    font-size: 14px;
    letter-spacing: 1px;
    color: var(--neon-cyan);
  }

  .panel-content {
    padding: 15px;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .player-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 8px 12px;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 4px;
  }

  .player-name {
    font-size: 13px;
    color: var(--text-primary);
  }

  .player-role {
    font-size: 11px;
    color: var(--text-secondary);
  }

  .vs-divider {
    text-align: center;
    font-size: 14px;
    color: var(--neon-pink);
    font-weight: bold;
    padding: 10px 0;
    text-shadow: 0 0 5px var(--neon-pink);
  }

  .current-action {
    margin-top: 10px;
    padding: 10px;
    background: rgba(0, 240, 255, 0.1);
    border: 1px solid var(--neon-cyan);
    border-radius: 4px;
  }

  .action-label {
    font-size: 11px;
    color: var(--text-secondary);
    margin-bottom: 6px;
    letter-spacing: 1px;
  }

  .action-detail {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .action-player {
    font-size: 13px;
    color: var(--neon-cyan);
    font-weight: bold;
  }

  .action-card {
    font-size: 12px;
    color: var(--text-primary);
  }
</style>
