CREATE TABLE game_session_txns
(
    trx_id varchar(64) PRIMARY KEY,
    ses_id NUMERIC REFERENCES game_sessions (id),
    action_type smallint not null,
    action_params numeric[] not null
);
