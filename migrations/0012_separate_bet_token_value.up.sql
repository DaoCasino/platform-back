ALTER TABLE game_sessions
  ADD COLUMN symbol VARCHAR(10) DEFAULT '4,BET',
  ADD COLUMN deposit_value NUMERIC DEFAULT 0,
  ADD COLUMN player_win_value NUMERIC DEFAULT NULL,
  ADD COLUMN token VARCHAR(7) DEFAULT 'BET';

CREATE INDEX sessions_deposits_idx ON game_sessions(token, deposit_value DESC);

UPDATE game_sessions SET
  symbol = ('4,' || SUBSTRING(deposit, '[A-Z]+$')::VARCHAR)::VARCHAR(10),
  deposit_value = SUBSTRING(deposit, '^\d+.\d+')::NUMERIC * 10000,
  player_win_value =
    CASE WHEN player_win_amount IS NULL THEN NULL ELSE SUBSTRING(player_win_amount, '^-?\d+.\d+')::NUMERIC * 10000 END,
  token = SUBSTRING(deposit, '[A-Z]+$')::VARCHAR(7);


