package game

import "math/rand"

type CardType string
type CardCategory string

const (
	CardCategoryAttack  CardCategory = "attack"
	CardCategoryDefense CardCategory = "defense"
	CardCategoryUtility CardCategory = "utility"
)

const (
	CardPortScan     CardType = "port_scan"
	CardBruteForce   CardType = "brute_force"
	CardSQLInjection CardType = "sql_injection"
	CardDDoS         CardType = "ddos"
	CardTrojan       CardType = "trojan"

	CardFirewall    CardType = "firewall"
	CardIDS         CardType = "ids"
	CardEncryption  CardType = "encryption"
	CardHoneypot    CardType = "honeypot"
	CardTrafficClean CardType = "traffic_clean"

	CardBandwidthUpgrade CardType = "bandwidth_upgrade"
	CardNodeRepair       CardType = "node_repair"
	CardDataTheft        CardType = "data_theft"
	CardBackdoor         CardType = "backdoor"
	CardSniff            CardType = "sniff"
)

type Card struct {
	ID          string       `json:"id"`
	Type        CardType     `json:"type"`
	Name        string       `json:"name"`
	Category    CardCategory `json:"category"`
	Description string       `json:"description"`
	Speed       int          `json:"speed"`
	Damage      int          `json:"damage,omitempty"`
	Cooldown    int          `json:"cooldown"`
	Rarity      string       `json:"rarity"`
	TargetSelf  bool         `json:"targetSelf"`
}

type PlayedCard struct {
	Card           *Card  `json:"card"`
	PlayerID       string `json:"playerId"`
	TargetNodeID   string `json:"targetNodeId"`
	TargetPlayerID string `json:"targetPlayerId"`
}

var CardDefinitions = map[CardType]*Card{
	CardPortScan: {
		Type:        CardPortScan,
		Name:        "端口扫描",
		Category:    CardCategoryAttack,
		Description: "揭示目标节点的类型和剩余血量",
		Speed:       9,
		Damage:      0,
		Cooldown:    0,
		Rarity:      "common",
		TargetSelf:  false,
	},
	CardBruteForce: {
		Type:        CardBruteForce,
		Name:        "暴力破解",
		Category:    CardCategoryAttack,
		Description: "对目标节点造成伤害(基础伤害减去防御力)",
		Speed:       5,
		Damage:      8,
		Cooldown:    0,
		Rarity:      "common",
		TargetSelf:  false,
	},
	CardSQLInjection: {
		Type:        CardSQLInjection,
		Name:        "SQL注入",
		Category:    CardCategoryAttack,
		Description: "无视防御，对数据库节点造成双倍伤害",
		Speed:       7,
		Damage:      6,
		Cooldown:    2,
		Rarity:      "rare",
		TargetSelf:  false,
	},
	CardDDoS: {
		Type:        CardDDoS,
		Name:        "DDoS洪水",
		Category:    CardCategoryAttack,
		Description: "消耗目标节点全部带宽，使其本回合无法响应防御指令",
		Speed:       4,
		Damage:      0,
		Cooldown:    3,
		Rarity:      "rare",
		TargetSelf:  false,
	},
	CardTrojan: {
		Type:        CardTrojan,
		Name:        "木马植入",
		Category:    CardCategoryAttack,
		Description: "不造成伤害，但每回合揭示目标周围节点信息",
		Speed:       6,
		Damage:      0,
		Cooldown:    2,
		Rarity:      "rare",
		TargetSelf:  false,
	},
	CardFirewall: {
		Type:        CardFirewall,
		Name:        "防火墙部署",
		Category:    CardCategoryDefense,
		Description: "为目标节点增加3点防御，持续3回合",
		Speed:       8,
		Damage:      0,
		Cooldown:    1,
		Rarity:      "common",
		TargetSelf:  true,
	},
	CardIDS: {
		Type:        CardIDS,
		Name:        "入侵检测",
		Category:    CardCategoryDefense,
		Description: "本回合被攻击时自动反击",
		Speed:       10,
		Damage:      0,
		Cooldown:    2,
		Rarity:      "uncommon",
		TargetSelf:  true,
	},
	CardEncryption: {
		Type:        CardEncryption,
		Name:        "数据加密",
		Category:    CardCategoryDefense,
		Description: "为节点增加2点防御，永久有效",
		Speed:       7,
		Damage:      0,
		Cooldown:    3,
		Rarity:      "uncommon",
		TargetSelf:  true,
	},
	CardHoneypot: {
		Type:        CardHoneypot,
		Name:        "蜜罐陷阱",
		Category:    CardCategoryDefense,
		Description: "节点被攻破时对攻击者造成反伤并暴露其一个节点位置",
		Speed:       8,
		Damage:      5,
		Cooldown:    3,
		Rarity:      "rare",
		TargetSelf:  true,
	},
	CardTrafficClean: {
		Type:        CardTrafficClean,
		Name:        "流量清洗",
		Category:    CardCategoryDefense,
		Description: "解除节点上的DDoS和木马状态",
		Speed:       9,
		Damage:      0,
		Cooldown:    1,
		Rarity:      "common",
		TargetSelf:  true,
	},
	CardBandwidthUpgrade: {
		Type:        CardBandwidthUpgrade,
		Name:        "带宽升级",
		Category:    CardCategoryUtility,
		Description: "永久提升节点2点最大带宽",
		Speed:       5,
		Damage:      0,
		Cooldown:    2,
		Rarity:      "uncommon",
		TargetSelf:  true,
	},
	CardNodeRepair: {
		Type:        CardNodeRepair,
		Name:        "节点修复",
		Category:    CardCategoryUtility,
		Description: "恢复目标节点10点生命值",
		Speed:       6,
		Damage:      0,
		Cooldown:    1,
		Rarity:      "common",
		TargetSelf:  true,
	},
	CardDataTheft: {
		Type:        CardDataTheft,
		Name:        "数据窃取",
		Category:    CardCategoryUtility,
		Description: "从目标数据库节点获取一张稀有卡(只能对已攻破的数据库使用)",
		Speed:       3,
		Damage:      0,
		Cooldown:    3,
		Rarity:      "rare",
		TargetSelf:  false,
	},
	CardBackdoor: {
		Type:        CardBackdoor,
		Name:        "后门安装",
		Category:    CardCategoryUtility,
		Description: "在目标节点安装后门，下次攻击可无视该节点直达后方",
		Speed:       5,
		Damage:      0,
		Cooldown:    4,
		Rarity:      "rare",
		TargetSelf:  false,
	},
	CardSniff: {
		Type:        CardSniff,
		Name:        "链路嗅探",
		Category:    CardCategoryUtility,
		Description: "揭示目标玩家的一行节点信息",
		Speed:       8,
		Damage:      0,
		Cooldown:    2,
		Rarity:      "uncommon",
		TargetSelf:  false,
	},
}

func CreateCard(cardType CardType, id string) *Card {
	def, exists := CardDefinitions[cardType]
	if !exists {
		return nil
	}
	card := *def
	card.ID = id
	return &card
}

func DrawRandomCard(id string, rarity string) *Card {
	cardTypes := make([]CardType, 0)
	for cardType, def := range CardDefinitions {
		if rarity == "" || def.Rarity == rarity {
			cardTypes = append(cardTypes, cardType)
		}
	}
	if len(cardTypes) == 0 {
		return nil
	}
	idx := rand.Intn(len(cardTypes))
	return CreateCard(cardTypes[idx], id)
}
