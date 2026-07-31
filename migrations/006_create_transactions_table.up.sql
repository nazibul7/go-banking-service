CREATE type transaction_type as enum ('deposit', 'withdraw', 'transfer');

CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'failed');

CREATE TABLE
    transactions (
        id SERIAL PRIMARY KEY,
        from_account_id INT REFERENCES accounts (id) ON DELETE RESTRICT,
        to_account_id INT REFERENCES accounts (id) ON DELETE RESTRICT,
        amount INT NOT NULL CHECK (amount > 0),
        transaction_type transaction_type NOT NULL,
        status transaction_status NOT NULL DEFAULT 'completed',
        idempotency_key TEXT NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW (),
        
        CONSTRAINT chk_trasaction_accounts CHECK (
            (
                transaction_type = 'deposit'
                AND from_account_id is NULL
                AND to_account_id is NOT NULL
            )
            or (
                transaction_type = 'withdraw'
                AND from_account_id IS NOT NULL
                AND to_account_id IS NULL
            )
            or (
                transaction_type = 'transfer'
                AND from_account_id IS NOT NULL
                AND to_account_id IS NOT NULL
                AND from_account_id <> to_account_id
            )
        )
    )