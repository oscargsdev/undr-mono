CREATE TABLE IF NOT EXISTS projects(
    id bigserial PRIMARY KEY,
    owner_id UUID NOT NULL,
    name text NOT NULL,
    handle text NOT NULL UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    FOREIGN KEY(owner_id) REFERENCES users(id) ON DELETE CASCADE
);