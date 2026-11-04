-- =========================================================
-- Printy DB Schema (PostgreSQL 13+)
-- =========================================================

-- Untuk PostgreSQL < 13, aktifkan extension ini:
-- CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ---------- ENUM ----------
CREATE TYPE user_role AS ENUM ('admin', 'customer');

-- ---------- USERS ----------
CREATE TABLE users (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name         varchar(100)  NOT NULL,
    email             varchar(100)  NOT NULL UNIQUE,
    phone             varchar(100),
    avatar_url        text,
    password_hash     varchar(255),  -- boleh NULL untuk user yang login via OAuth
    role              user_role     NOT NULL DEFAULT 'customer',
    is_email_verified boolean       NOT NULL DEFAULT false,
    created_at        timestamptz   NOT NULL DEFAULT now(),
    updated_at        timestamptz   NOT NULL DEFAULT now()
);

-- ---------- OAUTH ACCOUNTS ----------
CREATE TABLE oauth_accounts (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider         varchar(30)  NOT NULL,
    provider_user_id varchar(255) NOT NULL,
    created_at       timestamptz  NOT NULL DEFAULT now(),
    CONSTRAINT uq_oauth_provider_account UNIQUE (provider, provider_user_id)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);

-- ---------- SESSIONS ----------
CREATE TABLE sessions (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  varchar(255) NOT NULL UNIQUE,
    ip_address  inet,
    user_agent  text,
    expires_at  timestamptz  NOT NULL,
    created_at  timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- ---------- OTPS ----------
CREATE TABLE otps (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       uuid         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type          varchar(50)  NOT NULL,
    otp_code_hash varchar(255) NOT NULL,
    is_used       boolean      NOT NULL DEFAULT false,
    attempts      integer      NOT NULL DEFAULT 0,
    expired_at    timestamptz  NOT NULL,
    created_at    timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_otps_user_id_type ON otps(user_id, type);

-- ---------- EMAIL VERIFICATION TOKENS ----------
CREATE TABLE email_verification_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  varchar(255) NOT NULL UNIQUE,
    used_at     timestamptz,
    expired_at  timestamptz  NOT NULL,
    created_at  timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX idx_email_verification_tokens_user_id ON email_verification_tokens(user_id);

-- ---------- TRIGGER: auto update users.updated_at ----------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();