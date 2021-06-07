CREATE TYPE e_cashback_state AS ENUM ('accrued','claim');
ALTER TABLE cashback ADD state e_cashback_state NOT NULL DEFAULT 'accrued';