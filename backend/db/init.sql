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

CREATE TABLE IF NOT EXISTS tournaments (
    id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    max_players INTEGER NOT NULL,
    min_rank VARCHAR(20) DEFAULT 'none',
    creator_id UUID REFERENCES players(id),
    status VARCHAR(20) DEFAULT 'registering',
    registration_deadline TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    winner_id UUID REFERENCES players(id),
    current_round INTEGER DEFAULT 0,
    total_rounds INTEGER DEFAULT 0,
    bracket JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tournament_players (
    id SERIAL PRIMARY KEY,
    tournament_id UUID REFERENCES tournaments(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id),
    username VARCHAR(50) NOT NULL,
    elo_rating INTEGER DEFAULT 1200,
    current_rank VARCHAR(20) DEFAULT 'bronze',
    seed INTEGER,
    final_position INTEGER,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tournament_id, player_id)
);

CREATE TABLE IF NOT EXISTS tournament_matches (
    id UUID PRIMARY KEY,
    tournament_id UUID REFERENCES tournaments(id) ON DELETE CASCADE,
    round_number INTEGER NOT NULL,
    match_index INTEGER NOT NULL,
    player1_id UUID REFERENCES players(id),
    player2_id UUID REFERENCES players(id),
    player1_name VARCHAR(50),
    player2_name VARCHAR(50),
    winner_id UUID REFERENCES players(id),
    room_id VARCHAR(50),
    status VARCHAR(20) DEFAULT 'pending',
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tournament_chat (
    id SERIAL PRIMARY KEY,
    tournament_id UUID REFERENCES tournaments(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id),
    username VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tournament_records (
    id SERIAL PRIMARY KEY,
    tournament_id UUID REFERENCES tournaments(id) ON DELETE CASCADE,
    player_id UUID REFERENCES players(id),
    tournament_name VARCHAR(100) NOT NULL,
    final_position INTEGER,
    total_matches INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    elo_bonus INTEGER DEFAULT 0,
    has_top4_badge BOOLEAN DEFAULT false,
    bracket_snapshot JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tournaments_status ON tournaments(status);
CREATE INDEX idx_tournament_players_tournament ON tournament_players(tournament_id);
CREATE INDEX idx_tournament_matches_tournament ON tournament_matches(tournament_id);
CREATE INDEX idx_tournament_chat_tournament ON tournament_chat(tournament_id);
CREATE INDEX idx_tournament_records_player ON tournament_records(player_id);
CREATE INDEX idx_tournaments_created_at ON tournaments(created_at DESC);
