export const RANKS = {
  bronze: {
    name: '青铜',
    color: '#CD7F32',
    glowColor: 'rgba(205, 127, 50, 0.5)',
    minElo: 0,
    maxElo: 1099,
    icon: '🥉'
  },
  silver: {
    name: '白银',
    color: '#C0C0C0',
    glowColor: 'rgba(192, 192, 192, 0.5)',
    minElo: 1100,
    maxElo: 1399,
    icon: '🥈'
  },
  gold: {
    name: '黄金',
    color: '#FFD700',
    glowColor: 'rgba(255, 215, 0, 0.5)',
    minElo: 1400,
    maxElo: 1699,
    icon: '🥇'
  },
  platinum: {
    name: '铂金',
    color: '#E5E4E2',
    glowColor: 'rgba(229, 228, 226, 0.5)',
    minElo: 1700,
    maxElo: 1999,
    icon: '💎'
  },
  diamond: {
    name: '钻石',
    color: '#B9F2FF',
    glowColor: 'rgba(185, 242, 255, 0.5)',
    minElo: 2000,
    maxElo: 9999,
    icon: '💠'
  }
}

export const RANK_ORDER = ['bronze', 'silver', 'gold', 'platinum', 'diamond']

export function getRank(elo) {
  for (let i = RANK_ORDER.length - 1; i >= 0; i--) {
    const rankKey = RANK_ORDER[i]
    const rank = RANKS[rankKey]
    if (elo >= rank.minElo) {
      return rankKey
    }
  }
  return 'bronze'
}

export function getRankInfo(rankKey) {
  return RANKS[rankKey] || RANKS.bronze
}

export function getRankName(rankKey) {
  return RANKS[rankKey]?.name || '未知'
}

export function getRankColor(rankKey) {
  return RANKS[rankKey]?.color || '#888888'
}

export function getRankIcon(rankKey) {
  return RANKS[rankKey]?.icon || '❓'
}

export function getRankProgress(elo, rankKey) {
  const rank = RANKS[rankKey]
  if (!rank) return 0
  
  if (rankKey === 'diamond') {
    return 100
  }
  
  const range = rank.maxElo - rank.minElo
  const progress = elo - rank.minElo
  return Math.max(0, Math.min(100, (progress / range) * 100))
}

export function formatElo(elo) {
  return Math.round(elo).toLocaleString()
}

export function getWinRate(wins, losses) {
  const total = wins + losses
  if (total === 0) return 0
  return Math.round((wins / total) * 1000) / 10
}

const CARD_NAMES = {
  port_scan: '端口扫描',
  brute_force: '暴力破解',
  sql_injection: 'SQL注入',
  ddos: 'DDoS洪水',
  trojan: '木马植入',
  firewall: '防火墙部署',
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

export function getCardName(cardType) {
  return CARD_NAMES[cardType] || cardType
}
