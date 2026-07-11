-- Recension schema (idempotent; versions tracked in schema_migrations)

CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    api_key TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    slug TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Multi-tenant membership: users access teams only via this table.
CREATE TABLE IF NOT EXISTS team_members (
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('owner', 'admin', 'member', 'viewer')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (team_id, user_id)
);

CREATE INDEX IF NOT EXISTS team_members_user_idx ON team_members (user_id);

CREATE TABLE IF NOT EXISTS suites (
    id UUID PRIMARY KEY,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    baseline_batch_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (team_id, slug)
);

CREATE TABLE IF NOT EXISTS batches (
    id UUID PRIMARY KEY,
    suite_id UUID NOT NULL REFERENCES suites(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    sealed_at TIMESTAMPTZ,
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    meta JSONB NOT NULL DEFAULT '{}'::jsonb,
    UNIQUE (suite_id, slug)
);

CREATE TABLE IF NOT EXISTS elements (
    id UUID PRIMARY KEY,
    suite_id UUID NOT NULL REFERENCES suites(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    UNIQUE (suite_id, name)
);

CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    batch_id UUID NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    element_id UUID NOT NULL REFERENCES elements(id) ON DELETE CASCADE,
    built_at TIMESTAMPTZ NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (batch_id, element_id)
);

CREATE TABLE IF NOT EXISTS comparisons (
    id UUID PRIMARY KEY,
    src_message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    dst_message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    src_batch_id UUID NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    dst_batch_id UUID NOT NULL REFERENCES batches(id) ON DELETE CASCADE,
    result JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS comparisons_src_batch_idx ON comparisons (src_batch_id);
CREATE INDEX IF NOT EXISTS messages_batch_idx ON messages (batch_id);
