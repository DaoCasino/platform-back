CREATE TABLE cashback
(
    account_name VARCHAR(13) PRIMARY KEY REFERENCES users(account_name),
    eth_address VARCHAR(42),
    paid_cashback NUMERIC(27, 18) DEFAULT 0
);

INSERT INTO cashback(account_name) SELECT account_name FROM users;