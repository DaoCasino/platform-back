ALTER TABLE game_sessions
  ADD COLUMN deposit_value NUMERIC,
  ADD COLUMN player_win_value NUMERIC,
  ADD COLUMN token VARCHAR(7),
  ADD COLUMN token_prec NUMERIC;

CREATE INDEX sessions_deposits_idx ON game_sessions(token, deposit_value DESC);

UPDATE game_sessions SET
  deposit_value = SUBSTRING(deposit, '^\d+.\d+')::NUMERIC * 10000,
  player_win_value =
    CASE WHEN player_win_amount IS NULL THEN NULL ELSE SUBSTRING(player_win_amount, '^-?\d+.\d+')::NUMERIC * 10000 END,
  token = SUBSTRING(deposit, '[A-Z]+$')::VARCHAR(7),
  token_prec = 4;


