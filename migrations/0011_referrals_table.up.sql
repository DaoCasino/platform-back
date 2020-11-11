CREATE TABLE referrals
(
    account_name VARCHAR(13) REFERENCES users(account_name),
    referral_id VARCHAR NOT NULL
);

CREATE INDEX referral_id_idx ON referrals (referral_id);

