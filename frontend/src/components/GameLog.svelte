<script>
  import { gameStore } from '../store/gameStore.js'
  import { onMount } from 'svelte'

  let logs = []

  $: gameState = $gameStore.gameState
  $: {
    if (gameState?.gameLog) {
      logs = gameState.gameLog.slice(-20)
    }
  }

  let logContainer

  $: {
    if (logContainer && logs.length > 0) {
      setTimeout(() => {
        logContainer.scrollTop = logContainer.scrollHeight
      }, 50)
    }
  }
</script>

<div class="game-log panel">
  <div class="panel-header">
    <span>📜 战斗日志</span>
  </div>
  <div class="log-content" bind:this={logContainer}>
    {#each logs as log, index (index)}
      <div class="log-item">
        <span class="log-text">{log}</span>
      </div>
    {/each}
    {#if logs.length === 0}
      <div class="empty-log">
        <span>暂无战斗记录</span>
      </div>
    {/if}
  </div>
</div>

<style>
  .game-log {
    flex: 1;
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .log-content {
    flex: 1;
    overflow-y: auto;
    padding: 10px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .log-content::-webkit-scrollbar {
    width: 4px;
  }

  .log-content::-webkit-scrollbar-track {
    background: var(--bg-tertiary);
  }

  .log-content::-webkit-scrollbar-thumb {
    background: var(--neon-cyan);
    border-radius: 2px;
  }

  .log-item {
    font-size: 11px;
    line-height: 1.4;
    padding: 4px 8px;
    background: rgba(0, 0, 0, 0.2);
    border-left: 2px solid var(--neon-cyan);
    border-radius: 0 3px 3px 0;
  }

  .log-text {
    color: var(--text-primary);
    word-break: break-all;
  }

  .empty-log {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    font-size: 12px;
  }
</style>
