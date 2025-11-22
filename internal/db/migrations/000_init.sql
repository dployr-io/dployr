-- DEPLOYMENTS TABLE
CREATE TABLE IF NOT EXISTS deployments (
    id TEXT PRIMARY KEY,
    config JSON NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'failed', 'completed')),
    metadata JSON NOT NULL DEFAULT '{}',
    user_id TEXT NULL REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- SERVICES TABLE
CREATE TABLE IF NOT EXISTS services (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    source TEXT NOT NULL CHECK (source IN ('remote', 'image')),
    runtime TEXT NOT NULL CHECK (runtime IN ('static', 'golang', 'php', 'python', 'nodejs', 'ruby', 'dotnet', 'java', 'docker', 'k3s', 'custom')),
    runtime_version TEXT NOT NULL,
    run_cmd TEXT,
    build_cmd TEXT,
    working_dir TEXT NOT NULL,
    static_dir TEXT,
    image TEXT,
    remote_url TEXT,
    remote_branch TEXT,
    remote_commit_hash TEXT,
    deployment_id TEXT NULL REFERENCES deployments(id) ON DELETE SET NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- INSTANCE TABLE
CREATE TABLE IF NOT EXISTS instance (
    id TEXT PRIMARY KEY,
    token TEXT NOT NULL,
    instance_id TEXT NOT NULL,
    issuer TEXT NOT NULL,
    audience TEXT NOT NULL,
    registered_at INTEGER NOT NULL DEFAULT (unixepoch()),
    last_installed_at INTEGER NOT NULL DEFAULT (unixepoch())
);

-- Prevent updates to instance table
CREATE TRIGGER trg_prevent_instance_updates
BEFORE UPDATE ON instance
FOR EACH ROW
BEGIN
    SELECT
        CASE
            WHEN NEW.token IS NOT OLD.token THEN
                RAISE(ABORT, 'token column is immutable and cannot be updated')
            WHEN NEW.instance_id IS NOT OLD.instance_id THEN
                RAISE(ABORT, 'instance_id column is immutable and cannot be updated')
            WHEN NEW.issuer IS NOT OLD.issuer THEN
                RAISE(ABORT, 'issuer column is immutable and cannot be updated')
            WHEN NEW.audience IS NOT OLD.audience THEN
                RAISE(ABORT, 'audience column is immutable and cannot be updated')
            WHEN NEW.registered_at IS NOT OLD.registered_at THEN
                RAISE(ABORT, 'registered_at column is immutable and cannot be updated')
        END;
END;
