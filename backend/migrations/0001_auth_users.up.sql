-- ============================================================================
--  Migration 0001 — Auth & Users (ยกมาจาก docs/schema.sql เฉพาะส่วน AUTH)
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

DO $$ BEGIN
    CREATE TYPE otp_purpose AS ENUM ('login', 'verify_email', 'change_email');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- trigger function อัปเดต updated_at อัตโนมัติ
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- users
CREATE TABLE IF NOT EXISTS users (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id       UUID          NOT NULL DEFAULT gen_random_uuid(),
    email           CITEXT        NOT NULL,
    email_verified  BOOLEAN       NOT NULL DEFAULT FALSE,
    display_name    TEXT,
    avatar_url      TEXT,
    phone           TEXT,
    role            TEXT          NOT NULL DEFAULT 'customer'
                    CHECK (role IN ('customer', 'admin')),
    is_active       BOOLEAN       NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),

    CONSTRAINT uq_users_email     UNIQUE (email),
    CONSTRAINT uq_users_public_id UNIQUE (public_id)
);

DROP TRIGGER IF EXISTS trg_users_updated ON users;
CREATE TRIGGER trg_users_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- user_oauth_accounts (Google)
CREATE TABLE IF NOT EXISTS user_oauth_accounts (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id          BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         TEXT        NOT NULL DEFAULT 'google',
    provider_user_id TEXT        NOT NULL,
    provider_email   CITEXT,
    raw_profile      JSONB,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_oauth_provider_uid UNIQUE (provider, provider_user_id)
);

CREATE INDEX IF NOT EXISTS idx_oauth_user_id ON user_oauth_accounts(user_id);

-- otp_codes (เก็บเฉพาะแฮชของ OTP)
CREATE TABLE IF NOT EXISTS otp_codes (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id      BIGINT       REFERENCES users(id) ON DELETE CASCADE,
    email        CITEXT       NOT NULL,
    purpose      otp_purpose  NOT NULL DEFAULT 'login',
    code_hash    TEXT         NOT NULL,
    expires_at   TIMESTAMPTZ  NOT NULL,
    consumed_at  TIMESTAMPTZ,
    attempts     SMALLINT     NOT NULL DEFAULT 0,
    max_attempts SMALLINT     NOT NULL DEFAULT 5,
    request_ip   INET,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),

    CONSTRAINT chk_otp_attempts CHECK (attempts >= 0)
);

CREATE INDEX IF NOT EXISTS idx_otp_email_purpose ON otp_codes(email, purpose);
CREATE INDEX IF NOT EXISTS idx_otp_expires       ON otp_codes(expires_at);
CREATE INDEX IF NOT EXISTS idx_otp_active        ON otp_codes(email) WHERE consumed_at IS NULL;

-- auth_sessions (refresh token — เก็บเฉพาะแฮช)
CREATE TABLE IF NOT EXISTS auth_sessions (
    id                 BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id            BIGINT      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash TEXT        NOT NULL,
    user_agent         TEXT,
    ip_address         INET,
    expires_at         TIMESTAMPTZ NOT NULL,
    revoked_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_session_token UNIQUE (refresh_token_hash)
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON auth_sessions(user_id);
