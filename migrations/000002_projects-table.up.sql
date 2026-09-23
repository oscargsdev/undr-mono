CREATE TABLE IF NOT EXISTS projects(
    id bigserial PRIMARY KEY,
    owner_id UUID NOT NULL,
    name text NOT NULL,
    FOREIGN KEY(owner_id) REFERENCES users(id) ON DELETE CASCADE
);