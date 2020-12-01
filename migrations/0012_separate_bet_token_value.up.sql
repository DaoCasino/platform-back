ALTER TABLE game_sessions
ADD COLUMN symbol VARCHAR(7) DEFAULT 'BET',
ADD COLUMN deposit_value NUMERIC(13, 4) DEFAULT 0.0000,
ADD COLUMN player_win_value NUMERIC(13, 4) DEFAULT NULL;

UPDATE game_sessions SET symbol = SUBSTRING(deposit, '[A-Z]+$')::VARCHAR(7);
UPDATE game_sessions SET deposit_value = SUBSTRING(deposit, '^\d+.\d+')::NUMERIC(13, 4);
UPDATE game_sessions SET player_win_value =
CASE WHEN player_win_amount IS NULL THEN NULL ELSE SUBSTRING(player_win_amount, '^\d+.\d+')::NUMERIC(13, 4) END;
