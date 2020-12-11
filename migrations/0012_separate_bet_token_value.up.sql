ALTER TABLE game_sessions
ADD COLUMN symbol VARCHAR(10) DEFAULT '4,BET',
ADD COLUMN deposit_value NUMERIC DEFAULT 0,
ADD COLUMN player_win_value NUMERIC DEFAULT NULL;

CREATE INDEX sessions_deposits_idx ON game_sessions(symbol, deposit_value DESC);

UPDATE game_sessions SET symbol = '4,' || SUBSTRING(deposit, '[A-Z]+$')::VARCHAR(10);
UPDATE game_sessions SET deposit_value = SUBSTRING(deposit, '^\d+.\d+')::NUMERIC * 10000;
UPDATE game_sessions SET player_win_value =
CASE WHEN player_win_amount IS NULL THEN NULL ELSE SUBSTRING(player_win_amount, '^-?\d+.\d+')::NUMERIC * 10000 END;
