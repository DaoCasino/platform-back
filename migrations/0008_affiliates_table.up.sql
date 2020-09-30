CREATE TABLE affiliates
(
    account_name VARCHAR(13) PRIMARY KEY UNIQUE,
    affiliate_id VARCHAR NOT NULL
);

CREATE INDEX affiliates_id_idx ON affiliates (affiliate_id);

