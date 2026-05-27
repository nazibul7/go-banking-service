CREATE TABLE
    refresh_tokens (
        id SERIAL PRIMARY KEY,
        user_id INT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        token_hash TEXT NOT NULL UNIQUE,
        expires_at TIMESTAMP NOT NULL,
        revoked BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMP NOT NULL DEFAULT NOW ()
    );