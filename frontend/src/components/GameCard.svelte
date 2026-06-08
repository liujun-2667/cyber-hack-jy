<script>
  export let card

  $: categoryColor = {
    attack: '#ff3366',
    defense: '#00f0ff',
    utility: '#ff9900'
  }[card?.category] || '#888888'

  $: rarityColor = {
    common: '#888888',
    uncommon: '#00ff88',
    rare: '#00f0ff',
    legendary: '#ff00ff'
  }[card?.rarity] || '#888888'

  $: rarityText = {
    common: '普通',
    uncommon: '稀有',
    rare: '史诗',
    legendary: '传说'
  }[card?.rarity] || '普通'

  function getCardIcon(type) {
    const icons = {
      port_scan: '🔍',
      brute_force: '⚔️',
      sql_injection: '💉',
      ddos: '🌊',
      trojan: '🐴',
      firewall: '🛡️',
      ids: '📡',
      encryption: '🔐',
      honeypot: '🍯',
      traffic_clean: '🧹',
      bandwidth_upgrade: '📶',
      node_repair: '🔧',
      data_theft: '💾',
      backdoor: '🚪',
      sniff: '👃'
    }
    return icons[type] || '❓'
  }

  function getCategoryText(category) {
    const texts = {
      attack: '攻击',
      defense: '防御',
      utility: '功能'
    }
    return texts[category] || '未知'
  }
</script>

<div class="game-card" style="--card-color: {categoryColor}; --rarity-color: {rarityColor}">
  <div class="card-header">
    <span class="card-name">{card?.name || '未知卡牌'}</span>
    <span class="card-speed">速度 {card?.speed || 0}</span>
  </div>
  
  <div class="card-art">
    <span class="card-icon">{getCardIcon(card?.type)}</span>
  </div>
  
  <div class="card-info">
    <div class="card-type">
      <span class="type-badge" style="background: var(--rarity-color)">
        {getCategoryText(card?.category)}
      </span>
      <span class="rarity-text" style="color: var(--rarity-color)">
        {rarityText}
      </span>
    </div>
    
    {#if card?.damage > 0}
      <div class="card-stat">
        <span>伤害</span>
        <span class="value">{card.damage}</span>
      </div>
    {/if}
    
    {#if card?.cooldown > 0}
      <div class="card-stat">
        <span>冷却</span>
        <span class="value">{card.cooldown}回合</span>
      </div>
    {/if}
    
    <p class="card-desc">{card?.description || ''}</p>
  </div>
</div>

<style>
  .game-card {
    width: 100px;
    height: 130px;
    background: linear-gradient(135deg, #1a1a2e 0%, #0a0a0f 100%);
    border: 1px solid var(--card-color);
    border-radius: 6px;
    padding: 8px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    box-shadow: 0 0 10px rgba(0, 0, 0, 0.5), inset 0 0 10px rgba(0, 0, 0, 0.3);
    position: relative;
    overflow: hidden;
  }

  .game-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: -100%;
    width: 100%;
    height: 100%;
    background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.05), transparent);
    transition: left 0.5s;
  }

  .game-card:hover::before {
    left: 100%;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
  }

  .card-name {
    font-size: 11px;
    font-weight: bold;
    color: var(--card-color);
    line-height: 1.2;
  }

  .card-speed {
    font-size: 9px;
    color: var(--text-secondary);
    background: rgba(0, 0, 0, 0.3);
    padding: 2px 4px;
    border-radius: 3px;
    white-space: nowrap;
  }

  .card-art {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 4px;
    border: 1px solid rgba(255, 255, 255, 0.05);
  }

  .card-icon {
    font-size: 28px;
    filter: drop-shadow(0 0 5px var(--card-color));
  }

  .card-info {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .card-type {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .type-badge {
    font-size: 9px;
    padding: 2px 6px;
    border-radius: 3px;
    color: #000;
    font-weight: bold;
  }

  .rarity-text {
    font-size: 9px;
    font-weight: bold;
  }

  .card-stat {
    display: flex;
    justify-content: space-between;
    font-size: 10px;
    color: var(--text-secondary);
  }

  .card-stat .value {
    color: var(--text-primary);
    font-weight: bold;
  }

  .card-desc {
    font-size: 9px;
    color: var(--text-secondary);
    line-height: 1.3;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }
</style>
