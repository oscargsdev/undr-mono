CREATE TABLE IF NOT EXISTS genres(
    id bigserial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    created_by UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY(created_by) REFERENCES users(id) ON DELETE CASCADE
)