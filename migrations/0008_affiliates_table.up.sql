CREATE TABLE affiliates
(
    account_name VARCHAR(13) REFERENCES users(account_name),
    affiliate_id VARCHAR NOT NULL
);

CREATE INDEX affiliates_id_idx ON affiliates (affiliate_id);

