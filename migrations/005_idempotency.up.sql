CREATE TABLE
    idempotency_keys (
        id SERIAL PRIMARY KEY,
        idempotency_key TEXT NOT NULL,
        user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        response JSONB NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW (),
        UNIQUE (user_id, idempotency_key)
    )