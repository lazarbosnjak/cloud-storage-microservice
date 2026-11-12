CREATE TABLE credentials (
    uuid UUID PRIMARY KEY DEFAULT uuidv7(),
    email TEXT UNIQUE NOT NULL CHECK (email = LOWER(email) AND email <> ''),
    password_hash TEXT NOT NULL,
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    failed_attempts INTEGER NOT NULL DEFAULT 0 CHECK (failed_attempts >= 0),
    locked_until TIMESTAMPTZ NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);