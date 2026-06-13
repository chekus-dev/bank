CREATE TABLE IF NOT EXISTS wallets (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance        BIGINT NOT NULL DEFAULT 0,   -- stored in kobo
    ledger_balance BIGINT NOT NULL DEFAULT 0,
    currency       VARCHAR(3) NOT NULL DEFAULT 'NGN',
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

CREATE TABLE IF NOT EXISTS beneficiaries (
    id             UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_name   VARCHAR(255) NOT NULL,
    account_number VARCHAR(20) NOT NULL,
    bank_code      VARCHAR(10) NOT NULL,
    bank_name      VARCHAR(255) NOT NULL,
    recipient_code VARCHAR(100) NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_beneficiaries_user_id ON beneficiaries(user_id);
