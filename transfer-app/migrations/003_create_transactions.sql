CREATE TYPE transaction_type AS ENUM ('p2p', 'bank_transfer', 'funding', 'reversal');
CREATE TYPE transaction_status AS ENUM ('pending', 'success', 'failed', 'reversed');

CREATE TABLE IF NOT EXISTS transactions (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reference          VARCHAR(100) UNIQUE NOT NULL,
    sender_wallet_id   UUID REFERENCES wallets(id),
    receiver_wallet_id UUID REFERENCES wallets(id),
    amount             BIGINT NOT NULL,   -- in kobo
    fee                BIGINT NOT NULL DEFAULT 0,
    type               transaction_type NOT NULL,
    status             transaction_status NOT NULL DEFAULT 'pending',
    description        TEXT,
    metadata           JSONB DEFAULT '{}',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_reference ON transactions(reference);
CREATE INDEX IF NOT EXISTS idx_transactions_sender ON transactions(sender_wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_receiver ON transactions(receiver_wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at DESC);
