CREATE TABLE IF NOT EXISTS players (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    elo_rating INTEGER DEFAULT 1000,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
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
    result VARCHAR(10) NOT NULL
);

CREATE INDEX idx_players_elo ON players(elo_rating);
CREATE INDEX idx_game_records_player ON game_records USING GIN(player_ids);
