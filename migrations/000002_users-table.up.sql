CREATE TABLE users(
    id UUID PRIMARY KEY,
    email TEXT,
    display_name TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now()
);