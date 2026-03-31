CREATE TABLE IF NOT EXISTS visits (
    id         BIGSERIAL   PRIMARY KEY,
    link_id    INT      NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    user_agent TEXT        NOT NULL DEFAULT '',
    visited_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_visits_link_id ON visits(link_id);
