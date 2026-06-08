import { writable, derived } from 'svelte/store'

function createGameStore() {
  const { subscribe, set, update } = writable({
    connected: false,
    gameState: null,
    roomId: null,
    error: null,
    gameLog: [],
    players: []
  })

  let ws = null
  let reconnectAttempts = 0
  let reconnectTimer = null
  let username = ''

  function connect(name) {
    username = name
    
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.hostname
    const port = window.location.port === '5173' ? '8080' : (window.location.port || '80')
    
    const wsUrl = `${protocol}//${host}:${port}/ws?username=${encodeURIComponent(name)}`
    
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      console.log('WebSocket connected')
      update(state => ({ ...state, connected: true, error: null }))
      reconnectAttempts = 0
    }

    ws.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        handleMessage(message)
      } catch (e) {
        console.error('Error parsing message:', e)
      }
    }

    ws.onclose = () => {
      console.log('WebSocket disconnected')
      update(state => ({ ...state, connected: false }))
      
      if (reconnectAttempts < 5) {
        reconnectAttempts++
        reconnectTimer = setTimeout(() => {
          console.log('Reconnecting...')
          connect(username)
        }, 3000)
      }
    }

    ws.onerror = (error) => {
      console.error('WebSocket error:', error)
      update(state => ({ ...state, error: '连接错误' }))
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    if (ws) {
      ws.close()
      ws = null
    }
    set({
      connected: false,
      gameState: null,
      roomId: null,
      error: null,
      gameLog: [],
      players: []
    })
  }

  function handleMessage(message) {
    console.log('Received:', message.type, message.payload)
    
    switch (message.type) {
      case 'game_state':
        update(state => ({ ...state, gameState: message.payload }))
        break
      case 'room_created':
      case 'room_joined':
      case 'matched':
        update(state => ({ 
          ...state, 
          roomId: message.payload.roomId,
          players: message.payload.players || []
        }))
        break
      case 'player_joined':
        update(state => {
          const players = [...(state.players || []), { id: message.payload.playerId, username: message.payload.username }]
          return { ...state, players }
        })
        break
      case 'player_left':
      case 'player_disconnected':
        update(state => {
          const players = (state.players || []).filter(p => p.id !== message.payload.playerId)
          return { ...state, players }
        })
        break
      case 'phase_change':
        update(state => ({
          ...state,
          gameState: state.gameState ? { 
            ...state.gameState, 
            phase: message.payload.phase, 
            currentTurn: message.payload.turn 
          } : null
        }))
        break
      case 'card_played':
        break
      case 'execution_result':
        update(state => ({
          ...state,
          gameLog: message.payload.gameLog || []
        }))
        break
      case 'game_over':
        update(state => ({
          ...state,
          gameState: state.gameState ? { 
            ...state.gameState, 
            phase: 'gameover', 
            winnerId: message.payload.winnerId 
          } : null
        }))
        break
      case 'game_started':
        break
      case 'matchmaking_queued':
        break
      case 'error':
        update(state => ({ ...state, error: message.payload.message }))
        setTimeout(() => {
          update(state => ({ ...state, error: null }))
        }, 3000)
        break
      case 'chat':
        break
      default:
        console.log('Unhandled message type:', message.type)
    }
  }

  function send(type, payload = {}) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type, payload }))
      return true
    }
    console.warn('WebSocket not connected, cannot send:', type)
    return false
  }

  function quickMatch() {
    send('quick_match')
  }

  function cancelMatch() {
    disconnect()
  }

  function createRoom() {
    send('create_room', { maxPlayers: 2, gameMode: 'custom' })
  }

  function joinRoom(roomId) {
    send('join_room', { roomId })
  }

  function placeNode(nodeType, x, y) {
    send('place_node', { nodeType, x, y })
  }

  function startGame() {
    send('start_game')
  }

  function playCard(cardId, targetNodeId, targetPlayerId) {
    send('play_card', { cardId, targetNodeId, targetPlayerId })
  }

  function endTurn() {
    send('end_turn')
  }

  function requestGameState() {
    send('game_state')
  }

  function sendChat(message) {
    send('chat', { message })
  }

  return {
    subscribe,
    connect,
    disconnect,
    quickMatch,
    cancelMatch,
    createRoom,
    joinRoom,
    placeNode,
    startGame,
    playCard,
    endTurn,
    requestGameState,
    sendChat
  }
}

export const gameStore = createGameStore()

export const myPlayer = derived(gameStore, ($store) => {
  if (!$store.gameState) return null
  return {
    id: $store.gameState.myPlayerId,
    username: $store.gameState.myUsername,
    hand: $store.gameState.hand || [],
    playedCards: $store.gameState.playedCards || [],
    grid: $store.gameState.myGrid || [],
    coreHp: $store.gameState.coreHp || 0,
    coreMaxHp: $store.gameState.coreMaxHp || 30,
    cooldowns: $store.gameState.cooldowns || {}
  }
})

export const opponents = derived(gameStore, ($store) => {
  if (!$store.gameState || !$store.gameState.opponents) return []
  return Object.values($store.gameState.opponents)
})
