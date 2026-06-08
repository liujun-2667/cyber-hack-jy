<script>
  import { onMount } from 'svelte'
  import Lobby from './components/Lobby.svelte'
  import GameBoard from './components/GameBoard.svelte'
  import { gameStore } from './store/gameStore.js'

  let currentView = 'lobby'
  let username = ''

  $: if ($gameStore.gameState && $gameStore.gameState.phase) {
    currentView = 'game'
  }

  function handleLogin(name) {
    username = name
    gameStore.connect(name)
  }

  function handleBackToLobby() {
    currentView = 'lobby'
    gameStore.disconnect()
  }
</script>

<div class="app-container">
  <div class="cyber-grid-bg"></div>
  
  {#if currentView === 'lobby'}
    <Lobby onLogin={handleLogin} />
  {:else if currentView === 'game'}
    <GameBoard onBack={handleBackToLobby} />
  {/if}
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
</style>
