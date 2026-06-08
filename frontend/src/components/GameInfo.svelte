<script>
  import { gameStore, myPlayer, opponents } from '../store/gameStore.js'

  $: gameState = $gameStore.gameState
  $: player = $myPlayer
  $: opps = $opponents
  $: cooldowns = gameState?.cooldowns || {}

  $: cooldownList = Object.entries(cooldowns || {}).filter(([, cd]) => cd > 0)
</script>

<div class="game-info panel">
  <div class="panel-header">
    <span>📊 游戏信息</span>
  </div>
  
  <div class="info-content">
    <div class="info-section">
      <div class="info-row">
        <span class="label">当前回合</span>
        <span class="value neon-text-cyan">{gameState?.currentTurn || 1}/30</span>
      </div>
      <div class="info-row">
        <span class="label">游戏阶段</span>
        <span class="value">{getPhaseText(gameState?.phase)}</span>
      </div>
    </div>

    <div class="info-section">
      <h4>核心状态</h4>
      <div class="core-hp-bar">
        <div 
          class="core-hp-fill" 
          style="width: {player ? (player.coreHp / player.coreMaxHp) * 100 : 100}%"
        ></div>
        <span class="core-hp-text">
          {player?.coreHp || 30} / {player?.coreMaxHp || 30}
        </span>
      </div>
    </div>

    {#if cooldownList.length > 0}
      <div class="info-section">
        <h4>技能冷却</h4>
        <div class="cooldown-list">
          {#each cooldownList as [type, cd] (type)}
            <div class="cooldown-item">
              <span class="cd-name">{getCardName(type)}</span>
              <span class="cd-time">{cd}回合</span>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <div class="info-section">
      <h4>对手信息</h4>
      {#each opps as opp (opp.id)}
        <div class="opponent-info">
          <div class="opp-header">
            <span class="opp-name">{opp.username}</span>
            <span class="opp-status" class:alive={opp.isAlive}>
              {opp.isAlive ? '存活' : '淘汰'}
            </span>
          </div>
          <div class="opp-core-hp">
            <div class="mini-hp-bar">
              <div 
                class="mini-hp-fill"
                style="width: {(opp.coreHp / opp.coreMaxHp) * 100}%; background: var(--neon-red)"
              ></div>
            </div>
            <span>{opp.coreHp}/{opp.coreMaxHp}</span>
          </div>
        </div>
      {/each}
    </div>

    <div class="info-section">
      <h4>操作说明</h4>
      <ul class="tips">
        <li>点击卡牌选中，再点击目标释放</li>
        <li>每回合最多打出3张卡牌</li>
        <li>攻击必须从边缘节点逐步深入</li>
        <li>保护好你的核心节点！</li>
      </ul>
    </div>
  </div>
</div>

<style>
  .game-info {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .info-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .info-section h4 {
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 8px;
    letter-spacing: 1px;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    font-size: 13px;
    margin-bottom: 6px;
  }

  .label {
    color: var(--text-secondary);
  }

  .value {
    font-weight: bold;
  }

  .core-hp-bar {
    position: relative;
    height: 24px;
    background: var(--bg-tertiary);
    border: 1px solid var(--border-color);
    border-radius: 4px;
    overflow: hidden;
  }

  .core-hp-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--neon-green), var(--neon-cyan));
    transition: width 0.5s ease;
    box-shadow: 0 0 10px var(--neon-green);
  }

  .core-hp-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 12px;
    font-weight: bold;
    color: white;
    text-shadow: 0 0 3px #000;
  }

  .cooldown-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .cooldown-item {
    display: flex;
    justify-content: space-between;
    font-size: 12px;
    padding: 4px 8px;
    background: rgba(255, 51, 102, 0.1);
    border-radius: 3px;
  }

  .cd-name {
    color: var(--text-primary);
  }

  .cd-time {
    color: var(--neon-red);
    font-weight: bold;
  }

  .opponent-info {
    margin-bottom: 10px;
    padding: 8px;
    background: rgba(0, 0, 0, 0.2);
    border-radius: 4px;
  }

  .opp-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 6px;
  }

  .opp-name {
    font-size: 12px;
    font-weight: bold;
  }

  .opp-status {
    font-size: 10px;
    color: var(--neon-red);
  }

  .opp-status.alive {
    color: var(--neon-green);
  }

  .opp-core-hp {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 10px;
  }

  .mini-hp-bar {
    flex: 1;
    height: 8px;
    background: var(--bg-tertiary);
    border-radius: 2px;
    overflow: hidden;
  }

  .mini-hp-fill {
    height: 100%;
    transition: width 0.3s;
  }

  .tips {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .tips li {
    font-size: 11px;
    color: var(--text-secondary);
    line-height: 1.4;
    padding-left: 12px;
    position: relative;
  }

  .tips li::before {
    content: '▸';
    position: absolute;
    left: 0;
    color: var(--neon-cyan);
  }
</style>

<script>
  function getPhaseText(phase) {
    const texts = {
      setup: '准备',
      placement: '布阵',
      programming: '编程',
      execution: '执行',
      gameover: '结束'
    }
    return texts[phase] || phase
  }

  function getCardName(type) {
    const names = {
      port_scan: '端口扫描',
      brute_force: '暴力破解',
      sql_injection: 'SQL注入',
      ddos: 'DDoS洪水',
      trojan: '木马植入',
      firewall: '防火墙',
      ids: '入侵检测',
      encryption: '数据加密',
      honeypot: '蜜罐陷阱',
      traffic_clean: '流量清洗',
      bandwidth_upgrade: '带宽升级',
      node_repair: '节点修复',
      data_theft: '数据窃取',
      backdoor: '后门安装',
      sniff: '链路嗅探'
    }
    return names[type] || type
  }
</script>
