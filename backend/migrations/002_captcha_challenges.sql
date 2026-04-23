-- Migration: Create captcha_challenges table
-- Stores math CAPTCHA challenges with expiry

CREATE TABLE IF NOT EXISTS captcha_challenges (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL UNIQUE,
    question VARCHAR(50) NOT NULL,
    answer INTEGER NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_captcha_session ON captcha_challenges(session_id);
CREATE INDEX IF NOT EXISTS idx_captcha_expires ON captcha_challenges(expires_at);