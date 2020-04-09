CREATE TABLE users
(
    account_name VARCHAR(13) PRIMARY KEY UNIQUE,
    email        VARCHAR(64) NOT NULL
);

CREATE TABLE game_sessions
(
    id                NUMERIC PRIMARY KEY UNIQUE,
    player            VARCHAR(13) REFERENCES users (account_name),
    game_id           NUMERIC  NOT NULL,
    casino_id         NUMERIC  NOT NULL,
    blockchain_req_id NUMERIC  NOT NULL,
    state             SMALLINT NOT NULL
);

CREATE TABLE casinos
(
    id      NUMERIC PRIMARY KEY UNIQUE
);

CREATE TABLE game_session_updates
(
    ses_id      NUMERIC REFERENCES game_sessions (id),
    update_type INTEGER   NOT NULL,
    timestamp   TIMESTAMP NOT NULL,
    data        json
);