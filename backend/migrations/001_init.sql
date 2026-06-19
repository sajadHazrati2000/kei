-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "btree_gist";

-- organizations
CREATE TABLE organizations (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL,
    slug             TEXT NOT NULL UNIQUE,
    timezone         TEXT NOT NULL DEFAULT 'UTC',
    overlap_start    TIMETZ NOT NULL DEFAULT '09:00+00',
    overlap_end      TIMETZ NOT NULL DEFAULT '17:00+00',
    default_language TEXT  NOT NULL DEFAULT 'en',
    working_days     INT[] NOT NULL DEFAULT '{1,2,3,4,5}',
    working_start    TIMETZ NOT NULL DEFAULT '09:00+00',
    working_end      TIMETZ NOT NULL DEFAULT '17:00+00',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- users
CREATE TYPE user_role AS ENUM ('admin', 'member', 'viewer');
CREATE TYPE calendar_pref AS ENUM ('gregorian', 'jalali');

CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    email         TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          user_role NOT NULL DEFAULT 'member',
    timezone      TEXT NOT NULL DEFAULT 'UTC',
    language      TEXT NOT NULL DEFAULT 'en',
    calendar_pref calendar_pref NOT NULL DEFAULT 'gregorian',
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, email)
);

-- availability_slots
CREATE TYPE slot_status AS ENUM ('free', 'busy', 'focus');

CREATE TABLE availability_slots (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    slot_range    TSTZRANGE NOT NULL,
    status        slot_status NOT NULL DEFAULT 'free',
    is_override   BOOLEAN NOT NULL DEFAULT FALSE,
    recurrence_id UUID,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- Prevent overlapping busy/focus slots for the same user
    EXCLUDE USING GIST (
        user_id WITH =,
        slot_range WITH &&
    ) WHERE (status IN ('busy', 'focus'))
);

CREATE INDEX idx_availability_slots_user_id ON availability_slots(user_id);
CREATE INDEX idx_availability_slots_range ON availability_slots USING GIST(slot_range);

-- recurrence_templates
CREATE TYPE recurrence_pattern AS ENUM ('daily', 'weekly');

CREATE TABLE recurrence_templates (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pattern      recurrence_pattern NOT NULL DEFAULT 'weekly',
    days_of_week INT[] NOT NULL DEFAULT '{}',
    start_time   TIMETZ NOT NULL,
    end_time     TIMETZ NOT NULL,
    status       slot_status NOT NULL DEFAULT 'free',
    valid_from   DATE NOT NULL,
    valid_until  DATE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recurrence_templates_user_id ON recurrence_templates(user_id);

-- blocked_days
CREATE TABLE blocked_days (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id       UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    blocked_date DATE NOT NULL,
    reason       TEXT,
    created_by   UUID NOT NULL REFERENCES users(id),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(org_id, blocked_date)
);

-- audit_log
CREATE TABLE audit_log (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    actor_id    UUID REFERENCES users(id) ON DELETE SET NULL,
    action      TEXT NOT NULL,
    target_type TEXT,
    target_id   UUID,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_org_id ON audit_log(org_id);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);

-- refresh_tokens
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    jti        TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
