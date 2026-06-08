CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    elo_rating INTEGER DEFAULT 1200,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    current_streak INTEGER DEFAULT 0,
    best_streak INTEGER DEFAULT 0,
    rank_protection_games INTEGER DEFAULT 0,
    current_rank VARCHAR(20) DEFAULT 'bronze',
    best_rank VARCHAR(20) DEFAULT 'bronze',
    total_nodes_destroyed INTEGER DEFAULT 0,
    total_turns_survived INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS game_records (
    id UUID PRIMARY KEY,
    room_id VARCHAR(50) NOT NULL,
    player_ids UUID[] NOT NULL,
    winner_id UUID,
    game_mode VARCHAR(20) NOT NULL,
    turns INTEGER DEFAULT 0,
    duration INTEGER DEFAULT 0,
    season_id INTEGER DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS player_game_stats (
    id SERIAL PRIMARY KEY,
    game_id UUID REFERENCES game_records(id),
    player_id UUID REFERENCES players(id),
    nodes_destroyed INTEGER DEFAULT 0,
    cards_played INTEGER DEFAULT 0,
    damage_dealt INTEGER DEFAULT 0,
    damage_taken INTEGER DEFAULT 0,
    core_hp_remaining INTEGER DEFAULT 0,
    result VARCHAR(10) NOT NULL,
    elo_change INTEGER DEFAULT 0,
    elo_after INTEGER DEFAULT 1200,
    rank_after VARCHAR(20) DEFAULT 'bronze',
    rank_change VARCHAR(10) DEFAULT 'none'
);

CREATE TABLE IF NOT EXISTS card_usage_stats (
    id SERIAL PRIMARY KEY,
    player_id UUID REFERENCES players(id),
    card_type VARCHAR(50) NOT NULL,
    usage_count INTEGER DEFAULT 0,
    UNIQUE(player_id, card_type)
);

CREATE TABLE IF NOT EXISTS game_card_stats (
    id SERIAL PRIMARY KEY,
    game_id UUID REFERENCES game_records(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id),
    card_type VARCHAR(50) NOT NULL,
    usage_count INTEGER DEFAULT 0
);

CREATE INDEX idx_game_card_stats_game ON game_card_stats(game_id);
CREATE INDEX idx_game_card_stats_player ON game_card_stats(player_id);

CREATE TABLE IF NOT EXISTS seasons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS season_player_stats (
    id SERIAL PRIMARY KEY,
    season_id INTEGER REFERENCES seasons(id),
    player_id UUID REFERENCES players(id),
    start_elo INTEGER DEFAULT 1200,
    end_elo INTEGER DEFAULT 1200,
    best_elo INTEGER DEFAULT 1200,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    best_rank VARCHAR(20) DEFAULT 'bronze',
    UNIQUE(season_id, player_id)
);

CREATE INDEX idx_players_elo ON players(elo_rating);
CREATE INDEX idx_game_records_player ON game_records USING GIN(player_ids);
CREATE INDEX idx_card_usage_player ON card_usage_stats(player_id);
CREATE INDEX idx_season_player ON season_player_stats(season_id, player_id);

INSERT INTO seasons (name, start_date, end_date, is_active)
SELECT '第1赛季', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP + INTERVAL '30 days', true
WHERE NOT EXISTS (SELECT 1 FROM seasons);
