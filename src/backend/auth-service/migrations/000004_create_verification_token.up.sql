CREATE TABLE verification_token (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES credentials (id) ON DELETE CASCADE,
    token UUID NOT NULL UNIQUE DEFAULT uuidv7(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    revoked_at TIMESTAMPTZ NULL,
    reason INT NOT NULL
);

COMMENT ON COLUMN verification_token.reason 
    IS '0 - Registration verification; 1 - Magic Link';

CREATE INDEX idx_verification_token_user_id ON verification_token (user_id);
