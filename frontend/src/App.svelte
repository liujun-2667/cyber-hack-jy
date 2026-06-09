<script>
  import { onMount } from 'svelte'
  import Lobby from './components/Lobby.svelte'
  import GameBoard from './components/GameBoard.svelte'
  import ReplayView from './components/ReplayView.svelte'
  import TournamentWatch from './components/TournamentWatch.svelte'
  import { gameStore } from './store/gameStore.js'

  let currentView = 'lobby'
  let username = ''

  $: toasts = $gameStore.toasts || []
  $: watchingTournament = $gameStore.watchingTournament

  $: if ($gameStore.inGame && !$gameStore.replayMode) {
    currentView = 'game'
  } else if ($gameStore.replayMode) {
    currentView = 'replay'
  }

  function handleLogin(name) {
    username = name
    gameStore.connect(name)
  }

  function handleBackToLobby() {
    currentView = 'lobby'
    gameStore.disconnect()
  }

  function handleBackFromReplay() {
    currentView = 'game'
  }

  function handleCloseTournamentWatch() {
    if (watchingTournament) {
      gameStore.unwatchTournament(watchingTournament)
    }
  }

  function getToastClass(type) {
    const classMap = {
      info: 'toast-info',
      success: 'toast-success',
      warning: 'toast-warning',
      error: 'toast-error'
    }
    return classMap[type] || 'toast-info'
  }
</script>

<div class="app-container">
  <div class="cyber-grid-bg"></div>
  
  {#if currentView === 'lobby'}
    <Lobby onLogin={handleLogin} />
  {:else if currentView === 'game'}
    <GameBoard onBack={handleBackToLobby} onWatchReplay={() => currentView = 'replay'} />
  {:else if currentView === 'replay'}
    <ReplayView onBack={handleBackFromReplay} />
  {/if}

  {#if watchingTournament}
    <TournamentWatch onClose={handleCloseTournamentWatch} />
  {/if}

  <div class="toast-container">
    {#each toasts as toast (toast.id)}
      <div class="toast-item {getToastClass(toast.type)}" on:click={() => gameStore.dismissToast(toast.id)}>
        {toast.message}
      </div>
    {/each}
  </div>
</div>

<style>
  .app-container {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
  }

  .cyber-grid-bg {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-image: 
      linear-gradient(rgba(0, 240, 255, 0.03) 1px, transparent 1px),
      linear-gradient(90deg, rgba(0, 240, 255, 0.03) 1px, transparent 1px);
    background-size: 50px 50px;
    animation: gridMove 20s linear infinite;
    pointer-events: none;
  }

  @keyframes gridMove {
    0% { transform: perspective(500px) rotateX(60deg) translateY(0); }
    100% { transform: perspective(500px) rotateX(60deg) translateY(50px); }
  }

  .toast-container {
    position: fixed;
    top: 20px;
    right: 20px;
    z-index: 9999;
    display: flex;
    flex-direction: column;
    gap: 10px;
    pointer-events: none;
  }

  .toast-item {
    padding: 12px 20px;
    border-radius: 8px;
    font-size: 14px;
    font-weight: 500;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
    cursor: pointer;
    pointer-events: auto;
    animation: toastIn 0.3s ease-out;
    max-width: 360px;
    border: 1px solid var(--border-color);
  }

  @keyframes toastIn {
    from {
      opacity: 0;
      transform: translateX(100px);
    }
    to {
      opacity: 1;
      transform: translateX(0);
    }
  }

  .toast-info {
    background: rgba(0, 240, 255, 0.15);
    color: var(--neon-cyan);
    border-color: rgba(0, 240, 255, 0.4);
  }

  .toast-success {
    background: rgba(0, 255, 100, 0.15);
    color: #00ff64;
    border-color: rgba(0, 255, 100, 0.4);
  }

  .toast-warning {
    background: rgba(255, 215, 0, 0.15);
    color: #FFD700;
    border-color: rgba(255, 215, 0, 0.4);
  }

  .toast-error {
    background: rgba(255, 51, 102, 0.15);
    color: var(--neon-pink);
    border-color: rgba(255, 51, 102, 0.4);
  }
</style>
