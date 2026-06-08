<script>
  import GameCard from './GameCard.svelte'

  export let card

  function getCategoryText(category) {
    const texts = {
      attack: '攻击类',
      defense: '防御类',
      utility: '功能类'
    }
    return texts[category] || '未知'
  }

  function getRarityText(rarity) {
    const texts = {
      common: '普通',
      uncommon: '稀有',
      rare: '史诗',
      legendary: '传说'
    }
    return texts[rarity] || '普通'
  }
</script>

<div class="card-detail">
  <div class="card-large">
    <GameCard {card} />
  </div>
  
  <div class="card-stats-extended">
    <div class="stat-row">
      <span class="stat-label">名称</span>
      <span class="stat-value">{card?.name}</span>
    </div>
    <div class="stat-row">
      <span class="stat-label">类型</span>
      <span class="stat-value">{getCategoryText(card?.category)}</span>
    </div>
    <div class="stat-row">
      <span class="stat-label">稀有度</span>
      <span class="stat-value">{getRarityText(card?.rarity)}</span>
    </div>
    <div class="stat-row">
      <span class="stat-label">速度</span>
      <span class="stat-value">{card?.speed}</span>
    </div>
    {#if card?.damage > 0}
      <div class="stat-row">
        <span class="stat-label">伤害</span>
        <span class="stat-value damage">{card.damage}</span>
      </div>
    {/if}
    <div class="stat-row">
      <span class="stat-label">冷却</span>
      <span class="stat-value">{card?.cooldown || 0} 回合</span>
    </div>
  </div>
  
  <div class="card-description">
    <h4>效果说明</h4>
    <p>{card?.description}</p>
  </div>
</div>

<style>
  .card-detail {
    background: var(--bg-secondary);
    border: 1px solid var(--border-color);
    border-radius: 8px;
    padding: 16px;
    width: 240px;
    display: flex;
    flex-direction: column;
    gap: 16px;
    box-shadow: 0 0 20px rgba(0, 240, 255, 0.1);
  }

  .card-large {
    display: flex;
    justify-content: center;
    transform: scale(1.3);
    margin: 15px 0;
  }

  .card-stats-extended {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .stat-row {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    padding: 4px 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .stat-label {
    color: var(--text-secondary);
  }

  .stat-value {
    color: var(--text-primary);
    font-weight: bold;
  }

  .stat-value.damage {
    color: var(--neon-red);
  }

  .card-description {
    background: rgba(0, 0, 0, 0.2);
    padding: 10px;
    border-radius: 4px;
  }

  .card-description h4 {
    font-size: 11px;
    color: var(--neon-cyan);
    margin-bottom: 6px;
    letter-spacing: 1px;
  }

  .card-description p {
    font-size: 12px;
    color: var(--text-primary);
    line-height: 1.5;
  }
</style>
