CREATE TABLE
    idempotency_keys (
        id SERIAL PRIMARY KEY,
        idempotency_key TEXT NOT NULL,
        user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        status_code INT NOT NULL,
        -- JSONB is used instead of JSON because PostgreSQL stores it in a binary format,
        -- making it faster to query, index, and process. We don't need to preserve the
        -- original formatting or key order of the JSON response.
        response JSONB NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT NOW (),
        expires_at TIMESTAMP NOT NULL,
        -- Idempotency keys only need to be unique per user.
        -- Different users may legitimately generate the same key,
        -- so enforce uniqueness on the (user_id, idempotency_key)
        -- combination instead of idempotency_key alone.
        UNIQUE (user_id, idempotency_key)
    )