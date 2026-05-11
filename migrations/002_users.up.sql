CREATE TABLE
    users (
        id              SERIAL PRIMARY KEY,
        email           TEXT      NOT NULL UNIQUE,
        password_hash   TEXT      NOT NULL,
        role            TEXT      NOT NULL DEFAULT 'user',
        created_at      TIMESTAMP NOT NULL DEFAULT NOW ()
    );