CREATE TABLE credentials (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    email TEXT UNIQUE NOT NULL CHECK (email = LOWER(email) AND email <> ''),
    password_hash TEXT NOT NULL,
    user_role SMALLINT NOT NULL DEFAULT 2 CHECK (user_role IN (1,2)),
    password_changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    failed_attempts INTEGER NOT NULL DEFAULT 0 CHECK (failed_attempts >= 0),
    locked_until TIMESTAMPTZ NULL,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    locked_reason TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_enabled BOOLEAN NOT NULL DEFAULT FALSE
);

COMMENT ON COLUMN credentials.user_role IS '1 = admin, 2 = user';
