CREATE TABLE IF NOT EXISTS links (
    id         SERIAL   PRIMARY KEY,
    short_code VARCHAR(32) NOT NULL UNIQUE,
    long_url   TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
