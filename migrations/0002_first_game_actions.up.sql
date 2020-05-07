CREATE TABLE first_game_actions
(
    ses_id  NUMERIC REFERENCES game_sessions (id),
    type    SMALLINT NOT NULL,
    params  NUMERIC[] NOT NULL
);
