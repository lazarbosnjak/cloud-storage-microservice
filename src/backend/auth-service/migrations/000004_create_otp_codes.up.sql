CREATE TABLE otp_codes (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES credentials (id) ON DELETE CASCADE,
    code_hash BYTEA NOT NULL,
    attempts INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL
);

CREATE INDEX idx_otp_codes_user_id ON otp_codes (user_id);