CREATE TABLE active_token_nonces
(
    id           SERIAL PRIMARY KEY UNIQUE,
    account_name VARCHAR(13),
    token_nonce  NUMERIC NOT NULL,
    created      timestamp default current_timestamp
);