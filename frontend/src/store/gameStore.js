import { writable, derived } from 'svelte/store'

function createGameStore() {
  const { subscribe, set, update } = writable({
    connected: false,
    gameState: null,
    roomId: null,
    error: null,
    gameLog: [],
    players: [],
    inGame: false,
    lastReplayId: null,
    replayMode: false,
    currentReplay: null,
    replayTurn: 0,
    replayPlaying: false,
    replaySpeed: 1,
    playerInfo: null,
    matchmakingStatus: null,
    isMatching: false,
    rankResults: null,
    seasonInfo: null,
    onlineCount: 0,
    tournamentList: [],
    currentTournament: null,
    currentBracket: [],
    tournamentChat: [],
    watchingTournament: null,
    isInTournament: false,
    toasts: []
  })

  let ws = null
  let reconnectAttempts = 0
  let reconnectTimer = null
  let username = ''
  let toastIdCounter = 0

  function showToast(message, type = 'info', duration = 3000) {
    const id = ++toastIdCounter
    update(state => ({
      ...state,
      toasts: [...state.toasts, { id, message, type }]
    }))
    if (duration > 0) {
      setTimeout(() => {
        update(state => ({
          ...state,
          toasts: state.toasts.filter(t => t.id !== id)
        }))
      }, duration)
    }
    return id
  }

  function dismissToast(id) {
    update(state => ({
      ...state,
      toasts: state.toasts.filter(t => t.id !== id)
    }))
  }

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
      players: [],
      inGame: false,
      lastReplayId: null,
      replayMode: false,
      currentReplay: null,
      replayTurn: 0,
      replayPlaying: false,
      replaySpeed: 1,
      playerInfo: null,
      matchmakingStatus: null,
      isMatching: false,
      rankResults: null,
      seasonInfo: null
    })
  }

  function handleMessage(message) {
    console.log('Received:', message.type, message.payload)
    
    switch (message.type) {
      case 'game_state':
        update(state => ({ ...state, gameState: message.payload, inGame: true }))
        break
      case 'room_created':
      case 'room_joined':
      case 'matched':
        update(state => ({ 
          ...state, 
          roomId: message.payload.roomId,
          players: message.payload.players || [],
          isMatching: false,
          matchmakingStatus: null
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
          lastReplayId: message.payload.replayId || null,
          rankResults: message.payload.rankResults || null,
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
        update(state => ({
          ...state,
          isMatching: true,
          matchmakingStatus: {
            position: message.payload.position,
            eloRating: message.payload.eloRating,
            estimatedRange: message.payload.estimatedRange,
            waitTime: 0
          }
        }))
        break
      case 'matchmaking_status':
        update(state => ({
          ...state,
          matchmakingStatus: {
            ...state.matchmakingStatus,
            waitTime: message.payload.waitTime,
            estimatedRange: message.payload.estimatedRange,
            currentRange: message.payload.currentRange
          }
        }))
        break
      case 'matchmaking_cancelled':
        update(state => ({
          ...state,
          isMatching: false,
          matchmakingStatus: null
        }))
        break
      case 'player_info':
        update(state => ({
          ...state,
          playerInfo: message.payload,
          seasonInfo: message.payload.season || null
        }))
        break
      case 'season_reset':
        update(state => ({
          ...state,
          seasonResetNotice: message.payload
        }))
        break
      case 'error':
        update(state => ({ ...state, error: message.payload.message }))
        setTimeout(() => {
          update(state => ({ ...state, error: null }))
        }, 3000)
        break
      case 'chat':
        break
      case 'online_count':
        update(state => ({ ...state, onlineCount: message.payload.count }))
        break
      case 'tournament_list':
      case 'tournament_list_update':
        update(state => ({ 
          ...state, 
          tournamentList: message.payload.tournaments || [] 
        }))
        break
      case 'tournament_created':
      case 'tournament_joined':
        update(state => ({
          ...state,
          isInTournament: true,
          currentTournament: message.payload.tournament || state.currentTournament
        }))
        break
      case 'tournament_left':
        update(state => ({
          ...state,
          isInTournament: false,
          currentTournament: null
        }))
        break
      case 'tournament_update':
        update(state => {
          const tournamentId = message.payload.tournament?.id
          if (state.currentTournament?.id !== tournamentId && 
              state.watchingTournament !== tournamentId) {
            return state
          }
          const existingPlayers = state.currentTournament?.players || []
          const updatedTournament = {
            ...message.payload.tournament,
            players: message.payload.players || existingPlayers || []
          }
          return {
            ...state,
            currentTournament: updatedTournament
          }
        })
        break
      case 'tournament_watching':
        update(state => {
          const tournament = message.payload.tournament || {}
          if (message.payload.players) {
            tournament.players = message.payload.players
          }
          return {
            ...state,
            watchingTournament: message.payload.tournamentId,
            currentTournament: tournament,
            currentBracket: message.payload.bracket || [],
            tournamentChat: message.payload.chatMessages || []
          }
        })
        break
      case 'tournament_unwatched':
        update(state => ({
          ...state,
          watchingTournament: null,
          currentBracket: [],
          tournamentChat: []
        }))
        break
      case 'bracket_update':
        update(state => ({
          ...state,
          currentBracket: message.payload.matches || []
        }))
        break
      case 'chat_update':
        update(state => ({
          ...state,
          tournamentChat: message.payload.messages || []
        }))
        break
      case 'tournament_cancelled':
        update(state => ({
          ...state,
          isInTournament: false,
          currentTournament: null,
          watchingTournament: null,
          tournamentList: state.tournamentList.filter(t => t.id !== message.payload.tournamentId)
        }))
        showToast(`锦标赛「${message.payload.tournamentName || ''}」已取消：${message.payload.reason || '报名人数不足'}`, 'warning', 4000)
        break
      case 'tournament_kicked':
        update(state => ({
          ...state,
          isInTournament: false,
          currentTournament: null
        }))
        showToast(`你已被移出锦标赛「${message.payload.tournamentName || ''}」`, 'warning', 4000)
        break
      case 'match_progress':
        update(state => {
          const matches = state.currentBracket.map(m => {
            if (m.id === message.payload.matchId) {
              return {
                ...m,
                currentTurn: message.payload.currentTurn,
                maxTurns: message.payload.maxTurns
              }
            }
            return m
          })
          return { ...state, currentBracket: matches }
        })
        break
      case 'tournament_match_start':
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
    send('cancel_match')
    update(state => ({
      ...state,
      isMatching: false,
      matchmakingStatus: null
    }))
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

  function getApiBaseUrl() {
    const protocol = window.location.protocol
    const host = window.location.hostname
    const port = window.location.port === '5173' ? '8080' : (window.location.port || '80')
    return `${protocol}//${host}:${port}`
  }

  async function fetchReplay(roomId, replayId) {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/replay?roomId=${roomId}&replayId=${replayId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch replay')
      }
      const replayData = await response.json()
      update(state => ({
        ...state,
        currentReplay: replayData,
        replayMode: true,
        replayTurn: 1,
        replayPlaying: false
      }))
      return replayData
    } catch (error) {
      console.error('Error fetching replay:', error)
      update(state => ({ ...state, error: '获取回放失败' }))
      setTimeout(() => {
        update(state => ({ ...state, error: null }))
      }, 3000)
      return null
    }
  }

  async function fetchLeaderboard(limit = 20, playerId = null) {
    try {
      let url = `${getApiBaseUrl()}/api/leaderboard?limit=${limit}`
      if (playerId) {
        url += `&playerId=${playerId}`
      }
      const response = await fetch(url)
      if (!response.ok) {
        throw new Error('Failed to fetch leaderboard')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching leaderboard:', error)
      return null
    }
  }

  async function fetchPlayerStats(playerId) {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/player/stats?playerId=${playerId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch player stats')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching player stats:', error)
      return null
    }
  }

  async function fetchRecentGames(playerId, limit = 5) {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/player/recent-games?playerId=${playerId}&limit=${limit}`)
      if (!response.ok) {
        throw new Error('Failed to fetch recent games')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching recent games:', error)
      return null
    }
  }

  async function fetchSeasonInfo() {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/season`)
      if (!response.ok) {
        throw new Error('Failed to fetch season info')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching season info:', error)
      return null
    }
  }

  function setReplayTurn(turn) {
    update(state => ({
      ...state,
      replayTurn: turn
    }))
  }

  function setReplayPlaying(playing) {
    update(state => ({
      ...state,
      replayPlaying: playing
    }))
  }

  function setReplaySpeed(speed) {
    update(state => ({
      ...state,
      replaySpeed: speed
    }))
  }

  function exitReplay() {
    update(state => ({
      ...state,
      replayMode: false,
      currentReplay: null,
      replayTurn: 0,
      replayPlaying: false
    }))
  }

  function createTournament(name, maxPlayers, minRank, durationMinutes) {
    send('tournament_create', { name, maxPlayers, minRank, durationMinutes })
  }

  function requestTournamentList() {
    send('tournament_list')
  }

  function joinTournament(tournamentId) {
    send('tournament_join', { tournamentId })
  }

  function leaveTournament(tournamentId) {
    send('tournament_leave', { tournamentId })
  }

  function watchTournament(tournamentId) {
    send('tournament_watch', { tournamentId })
  }

  function unwatchTournament(tournamentId) {
    send('tournament_unwatch', { tournamentId })
  }

  function requestTournamentDetail(tournamentId) {
    send('tournament_detail', { tournamentId })
  }

  function requestTournamentBracket(tournamentId) {
    send('tournament_bracket', { tournamentId })
  }

  function sendTournamentChat(tournamentId, message) {
    send('tournament_chat', { tournamentId, message })
  }

  function requestTournamentChatHistory(tournamentId, limit = 50) {
    send('tournament_chat_history', { tournamentId, limit })
  }

  async function fetchTournaments(status = '') {
    try {
      let url = `${getApiBaseUrl()}/api/tournaments`
      if (status) {
        url += `?status=${status}`
      }
      const response = await fetch(url)
      if (!response.ok) {
        throw new Error('Failed to fetch tournaments')
      }
      const data = await response.json()
      update(state => ({ ...state, tournamentList: data.tournaments || [] }))
      return data.tournaments || []
    } catch (error) {
      console.error('Error fetching tournaments:', error)
      return []
    }
  }

  async function fetchTournament(tournamentId) {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/tournament?tournamentId=${tournamentId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch tournament')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching tournament:', error)
      return null
    }
  }

  async function fetchTournamentBracket(tournamentId) {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/tournament/bracket?tournamentId=${tournamentId}`)
      if (!response.ok) {
        throw new Error('Failed to fetch bracket')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching bracket:', error)
      return null
    }
  }

  async function fetchPlayerTournaments(playerId, limit = 20) {
    try {
      const response = await fetch(`${getApiBaseUrl()}/api/player/tournaments?playerId=${playerId}&limit=${limit}`)
      if (!response.ok) {
        throw new Error('Failed to fetch player tournaments')
      }
      return await response.json()
    } catch (error) {
      console.error('Error fetching player tournaments:', error)
      return { tournaments: [] }
    }
  }

  function kickPlayer(tournamentId, playerId) {
    update(state => {
      if (state.currentTournament?.id !== tournamentId || !state.currentTournament?.players) {
        return state
      }
      const newPlayers = state.currentTournament.players.filter(
        p => (p.playerId || p.id || p.player_id) !== playerId
      )
      return {
        ...state,
        currentTournament: {
          ...state.currentTournament,
          players: newPlayers,
          playerCount: newPlayers.length
        }
      }
    })
    send('tournament_kick', { tournamentId, playerId })
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
    sendChat,
    fetchReplay,
    fetchLeaderboard,
    fetchPlayerStats,
    fetchRecentGames,
    fetchSeasonInfo,
    setReplayTurn,
    setReplayPlaying,
    setReplaySpeed,
    exitReplay,
    getApiBaseUrl,
    createTournament,
    requestTournamentList,
    joinTournament,
    leaveTournament,
    watchTournament,
    unwatchTournament,
    requestTournamentDetail,
    requestTournamentBracket,
    sendTournamentChat,
    requestTournamentChatHistory,
    fetchTournaments,
    fetchTournament,
    fetchTournamentBracket,
    fetchPlayerTournaments,
    kickPlayer,
    showToast,
    dismissToast
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
