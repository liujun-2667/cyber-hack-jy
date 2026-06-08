<script>
  import { gameStore } from '../store/gameStore.js'

  export let currentTurn = 1
  export let totalTurns = 1
  export let isPlaying = false
  export let speed = 1

  function togglePlay() {
    gameStore.setReplayPlaying(!isPlaying)
  }

  function handleProgressChange(e) {
    const turn = parseInt(e.target.value)
    gameStore.setReplayTurn(turn)
  }

  function changeSpeed() {
    const speeds = [1, 2, 4]
    const currentIndex = speeds.indexOf(speed)
    const nextIndex = (currentIndex + 1) % speeds.length
    gameStore.setReplaySpeed(speeds[nextIndex])
  }

  function prevTurn() {
    if (currentTurn > 1) {
      gameStore.setReplayTurn(currentTurn - 1)
    }
  }

  function nextTurn() {
    if (currentTurn < totalTurns) {
      gameStore.setReplayTurn(currentTurn + 1)
    }
  }
</script>

<div class="replay-controls">
  <div class="controls-left">
    <button class="control-btn" on:click={prevTurn} disabled={currentTurn <= 1}>
      ⏮
    </button>
    <button class="control-btn play-btn" on:click={togglePlay}>
      {isPlaying ? '⏸' : '▶'}
    </button>
    <button class="control-btn" on:click={nextTurn} disabled={currentTurn >= totalTurns}>
      ⏭
    </button>
  </div>

  <div class="controls-center">
    <span class="turn-indicator">
      回合 {currentTurn} / {totalTurns}
    </span>
    <div class="progress-container">
      <input
        type="range"
        min="1"
        max={totalTurns}
        value={currentTurn}
        on:input={handleProgressChange}
        class="progress-slider"
      />
      <div class="progress-track">
        <div
          class="progress-fill"
          style="width: {((currentTurn - 1) / (totalTurns - 1)) * 100}%"
        ></div>
      </div>
    </div>
  </div>

  <div class="controls-right">
    <button class="speed-btn" on:click={changeSpeed}>
      {speed}x
    </button>
  </div>
</div>

<style>
  .replay-controls {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 20px;
    background: var(--bg-secondary);
    border-top: 1px solid var(--border-color);
    gap: 20px;
  }

  .controls-left,
  .controls-right {
    display: flex;
    align-items: center;
    gap: 10px;
    min-width: 100px;
  }

  .controls-right {
    justify-content: flex-end;
  }

  .controls-center {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }

  .control-btn {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    width: 36px;
    height: 36px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 14px;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s;
  }

  .control-btn:hover:not(:disabled) {
    border-color: var(--neon-cyan);
    color: var(--neon-cyan);
    box-shadow: 0 0 8px rgba(0, 240, 255, 0.3);
  }

  .control-btn:disabled {
    opacity: 0.3;
    cursor: not-allowed;
  }

  .play-btn {
    width: 44px;
    height: 44px;
    font-size: 16px;
    border-color: var(--neon-cyan);
    color: var(--neon-cyan);
  }

  .play-btn:hover {
    background: rgba(0, 240, 255, 0.1);
  }

  .turn-indicator {
    font-size: 13px;
    color: var(--text-secondary);
    letter-spacing: 1px;
  }

  .progress-container {
    width: 100%;
    max-width: 400px;
    position: relative;
    height: 6px;
  }

  .progress-track {
    position: absolute;
    top: 50%;
    left: 0;
    right: 0;
    height: 4px;
    background: var(--bg-tertiary);
    border-radius: 2px;
    transform: translateY(-50%);
    pointer-events: none;
  }

  .progress-fill {
    height: 100%;
    background: var(--neon-cyan);
    border-radius: 2px;
    box-shadow: 0 0 6px var(--neon-cyan);
  }

  .progress-slider {
    position: relative;
    width: 100%;
    height: 6px;
    -webkit-appearance: none;
    appearance: none;
    background: transparent;
    cursor: pointer;
    z-index: 1;
  }

  .progress-slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: var(--neon-cyan);
    cursor: pointer;
    box-shadow: 0 0 10px var(--neon-cyan);
    border: 2px solid var(--bg-primary);
  }

  .progress-slider::-moz-range-thumb {
    width: 16px;
    height: 16px;
    border-radius: 50%;
    background: var(--neon-cyan);
    cursor: pointer;
    box-shadow: 0 0 10px var(--neon-cyan);
    border: 2px solid var(--bg-primary);
  }

  .speed-btn {
    background: transparent;
    border: 1px solid var(--border-color);
    color: var(--text-primary);
    padding: 8px 16px;
    border-radius: 4px;
    cursor: pointer;
    font-size: 13px;
    font-weight: bold;
    transition: all 0.3s;
  }

  .speed-btn:hover {
    border-color: var(--neon-pink);
    color: var(--neon-pink);
    box-shadow: 0 0 8px rgba(255, 0, 255, 0.3);
  }
</style>
