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
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
