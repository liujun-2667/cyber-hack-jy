<script>
  import GameCard from './GameCard.svelte'

  export let cards = []
  export let playedCards = []
  export let selectedCard = null
  export let disabled = false
  export let onCardSelect = () => {}

  function handleCardClick(card) {
    if (disabled) return
    onCardSelect(card)
  }

  function handleKeydown(event, card) {
    if (disabled) return
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault()
      onCardSelect(card)
    }
  }

  $: playedCardIds = playedCards.map(pc => pc.card?.id)
</script>

<div class="hand-cards" class:disabled={disabled}>
  <div class="hand-header">
    <span class="hand-label">手牌</span>
    <span class="hand-count">{cards.length}/7</span>
  </div>
  <div class="cards-container">
    {#each cards as card (card.id)}
      <div 
        class="card-wrapper"
        class:selected={selectedCard?.id === card.id}
        class:played={playedCardIds.includes(card.id)}
        role="button"
        tabindex="0"
        aria-label={`选择卡牌 ${card.name}`}
        on:click={() => handleCardClick(card)}
        on:keydown={(e) => handleKeydown(e, card)}
      >
        <GameCard card={card} />
      </div>
    {/each}
    
    {#if cards.length === 0}
      <div class="empty-hand">
        <span>没有手牌</span>
      </div>
    {/if}
  </div>
</div>

<style>
  .hand-cards {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .hand-cards.disabled {
    opacity: 0.5;
    pointer-events: none;
  }

  .hand-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0 4px;
  }

  .hand-label {
    font-size: 12px;
    color: var(--neon-cyan);
    letter-spacing: 2px;
  }

  .hand-count {
    font-size: 12px;
    color: var(--text-secondary);
  }

  .cards-container {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding: 4px;
    min-height: 140px;
  }

  .card-wrapper {
    flex-shrink: 0;
    transition: all 0.2s ease;
    cursor: pointer;
    transform: translateY(0);
  }

  .card-wrapper:hover {
    transform: translateY(-10px);
  }

  .card-wrapper:focus {
    outline: 2px solid var(--neon-cyan);
    outline-offset: 2px;
    transform: translateY(-10px);
  }

  .card-wrapper:focus-visible {
    outline: 2px solid var(--neon-cyan);
    outline-offset: 2px;
    transform: translateY(-10px);
  }

  .card-wrapper.selected {
    transform: translateY(-15px);
  }

  .card-wrapper.selected :global(.game-card) {
    box-shadow: 0 0 20px var(--neon-yellow), 0 0 40px rgba(255, 255, 0, 0.3);
    border-color: var(--neon-yellow);
  }

  .card-wrapper.played {
    opacity: 0.3;
    pointer-events: none;
  }

  .empty-hand {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: var(--text-secondary);
    font-size: 13px;
  }
</style>
