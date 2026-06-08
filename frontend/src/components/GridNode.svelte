<script>
  export let node
  export let x = 0
  export let y = 0
  export let isOpponent = false
  export let targetMode = false
  export let onClick = () => {}

  const NODE_WIDTH = 50
  const NODE_HEIGHT = 50

  $: isUnknown = node && node.unknown
  $: isAlive = node && (node.alive !== false && node.hp !== 0)
  $: nodeType = node?.type || 'unknown'

  $: colors = {
    core: { primary: '#ff00ff', secondary: '#990099', glow: 'rgba(255, 0, 255, 0.6)' },
    server: { primary: '#00f0ff', secondary: '#0088aa', glow: 'rgba(0, 240, 255, 0.5)' },
    database: { primary: '#00ff88', secondary: '#008844', glow: 'rgba(0, 255, 136, 0.5)' },
    firewall: { primary: '#ff9900', secondary: '#995500', glow: 'rgba(255, 153, 0, 0.5)' },
    router: { primary: '#9933ff', secondary: '#551199', glow: 'rgba(153, 51, 255, 0.5)' },
    unknown: { primary: '#333355', secondary: '#1a1a2e', glow: 'rgba(100, 100, 150, 0.3)' }
  }

  $: color = colors[nodeType] || colors.unknown
  $: hpPercent = node?.maxHp ? (node.hp / node.maxHp) * 100 : 100
  $: hpColor = hpPercent > 60 ? '#00ff88' : hpPercent > 30 ? '#ffff00' : '#ff3366'

  $: hasStatusEffect = node?.status?.hasDDoS || node?.status?.hasTrojan || node?.status?.hasIDS || node?.status?.firewallTurns > 0

  function handleClick(e) {
    if (targetMode && isAlive && !isUnknown) {
      onClick()
    }
  }

  function handleKeydown(e) {
    if (targetMode && isAlive && !isUnknown) {
      if (e.key === 'Enter' || e.key === ' ') {
        e.preventDefault()
        onClick()
      }
    }
  }

  function getTypeIcon(type) {
    const icons = {
      core: '◈',
      server: '▣',
      database: '▤',
      firewall: '▮',
      router: '◇'
    }
    return icons[type] || '?'
  }
</script>

<g 
  class="grid-node" 
  class:unknown={isUnknown}
  class:dead={!isAlive}
  class:targetable={targetMode && isAlive && !isUnknown}
  transform="translate({x}, {y})"
  role={targetMode && isAlive && !isUnknown ? 'button' : undefined}
  tabindex={targetMode && isAlive && !isUnknown ? 0 : -1}
  aria-label={targetMode && isAlive && !isUnknown ? `选择节点 ${node?.name || '未知'}` : undefined}
  on:click={handleClick}
  on:keydown={handleKeydown}
  style="pointer-events: all; cursor: {targetMode && isAlive && !isUnknown ? 'pointer' : 'default'}"
>
  <defs>
    <filter id="glow-{node?.id || 'unknown'}" x="-50%" y="-50%" width="200%" height="200%">
      <feGaussianBlur stdDeviation="3" result="coloredBlur"/>
      <feMerge>
        <feMergeNode in="coloredBlur"/>
        <feMergeNode in="SourceGraphic"/>
      </feMerge>
    </filter>
  </defs>

  {#if isUnknown}
    <polygon 
      points="25,0 50,12.5 50,37.5 25,50 0,37.5 0,12.5"
      fill="#1a1a2e"
      stroke="#333355"
      stroke-width="1"
      opacity="0.6"
    />
    <text x="25" y="28" text-anchor="middle" fill="#444466" font-size="16">?</text>
  {:else if !isAlive}
    <polygon 
      points="25,0 50,12.5 50,37.5 25,50 0,37.5 0,12.5"
      fill="#0a0a0f"
      stroke="#330000"
      stroke-width="1"
      opacity="0.4"
    />
    <text x="25" y="28" text-anchor="middle" fill="#330000" font-size="12">✕</text>
  {:else}
    <polygon 
      points="25,0 50,12.5 50,37.5 25,50 0,37.5 0,12.5"
      fill={color.secondary}
      stroke={color.primary}
      stroke-width="2"
      filter="url(#glow-{node?.id || 'node'})"
    />
    
    <polygon 
      points="25,8 42,17.5 42,32.5 25,42 8,32.5 8,17.5"
      fill={color.primary}
      opacity="0.2"
    />

    <text x="25" y="22" text-anchor="middle" fill={color.primary} font-size="10" font-weight="bold">
      {getTypeIcon(nodeType)}
    </text>

    {#if node?.hp !== undefined && nodeType !== 'unknown'}
      <rect x="8" y="38" width="34" height="4" fill="#1a1a2e" rx="2" />
      <rect 
        x="9" 
        y="39" 
        width={32 * (hpPercent / 100)} 
        height="2" 
        fill={hpColor} 
        rx="1"
      />
    {/if}

    {#if hasStatusEffect}
      <g class="status-icons">
        {#if node?.status?.firewallTurns > 0}
          <circle cx="42" cy="8" r="5" fill="#ff9900" opacity="0.8" />
        {/if}
        {#if node?.status?.hasTrojan}
          <circle cx="8" cy="8" r="5" fill="#9933ff" opacity="0.8" />
        {/if}
        {#if node?.status?.hasDDoS}
          <circle cx="42" cy="42" r="5" fill="#ff3366" opacity="0.8" />
        {/if}
      </g>
    {/if}
  {/if}
</g>

<style>
  .grid-node {
    transition: all 0.3s ease;
  }

  .grid-node.targetable:hover polygon {
    filter: brightness(1.3) drop-shadow(0 0 8px currentColor);
  }

  .grid-node.dead {
    opacity: 0.3;
  }
</style>
