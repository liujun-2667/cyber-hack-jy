<script>
  import { onMount } from 'svelte'
  import GridNode from './GridNode.svelte'

  export let grid = []
  export let isOpponent = false
  export let ownerId = ''
  export let targetMode = false
  export let onNodeClick = () => {}

  const TILE_WIDTH = 64
  const TILE_HEIGHT = 32
  const GRID_OFFSET_X = 200
  const GRID_OFFSET_Y = 80

  function getNodeAt(x, y) {
    if (grid[x] && grid[x][y]) {
      return grid[x][y]
    }
    return null
  }

  function handleNodeClick(node) {
    if (!node || !node.alive) return
    if (node.unknown && isOpponent) return
    onNodeClick(node, ownerId, isOpponent)
  }

  $: gridArray = []
  $: {
    gridArray = []
    for (let y = 0; y < 5; y++) {
      for (let x = 0; x < 5; x++) {
        const node = getNodeAt(x, y)
        const isoX = GRID_OFFSET_X + (x - y) * (TILE_WIDTH / 2)
        const isoY = GRID_OFFSET_Y + (x + y) * (TILE_HEIGHT / 2)
        gridArray.push({ x, y, node, isoX, isoY, zIndex: x + y })
      }
    }
    gridArray.sort((a, b) => a.zIndex - b.zIndex)
  }
</script>

<div class="isometric-grid" class:target-mode={targetMode}>
  <svg class="grid-lines" width="400" height="250" viewBox="0 0 400 250">
    {#each gridArray as item (item.x + '-' + item.y)}
      {#if item.node}
        <GridNode
          node={item.node}
          x={item.isoX}
          y={item.isoY}
          isOpponent={isOpponent}
          targetMode={targetMode}
          onClick={() => handleNodeClick(item.node)}
        />
      {/if}
    {/each}
  </svg>
</div>

<style>
  .isometric-grid {
    position: relative;
    width: 400px;
    height: 250px;
    cursor: default;
  }

  .isometric-grid.target-mode {
    cursor: crosshair;
  }

  .grid-lines {
    position: absolute;
    top: 0;
    left: 0;
    pointer-events: none;
  }
</style>
