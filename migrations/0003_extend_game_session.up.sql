ALTER TABLE game_sessions
    ADD COLUMN deposit           VARCHAR(64) NOT NULL DEFAULT '0.0000 BET',
    ADD COLUMN last_update       NUMERIC     NOT NULL DEFAULT 0,
    ADD COLUMN player_win_amount VARCHAR(64) DEFAULT NULL;