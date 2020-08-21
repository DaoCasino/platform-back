CREATE INDEX sessions_last_wins_idx ON game_sessions (state, last_update DESC, casino_id) WHERE player_win_amount NOT SIMILAR TO '-%|0.0000';
CREATE INDEX sessions_last_lost_idx ON game_sessions (state, last_update DESC, casino_id) WHERE player_win_amount SIMILAR TO '-%';
CREATE INDEX sessions_last_all_idx on game_sessions (state, last_update DESC, casino_id);