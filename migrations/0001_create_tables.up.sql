CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS teams
(
        id BIGSERIAL PRIMARY KEY,
        name VARCHAR(255) UNIQUE NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users
(
        id TEXT PRIMARY KEY DEFAULT uuid_generate_v4()::TEXT,
        name VARCHAR(255) UNIQUE NOT NULL,
        is_active BOOLEAN NOT NULL DEFAULT TRUE,
        team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
        updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pull_requests
(
        id TEXT PRIMARY KEY DEFAULT uuid_generate_v4()::TEXT,
        name VARCHAR(255) NOT NULL,
        author_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
        status VARCHAR(10) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'MERGED')),
        created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
        merged_at TIMESTAMP WITH TIME ZONE NULL
);

-- Таблица 4: pr_reviewers
CREATE TABLE IF NOT EXISTS pr_reviewers
(
        pr_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
        reviewer_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
        assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

        PRIMARY KEY (pr_id, reviewer_id)
);

-- Индексы для ускорения поиска
CREATE INDEX idx_pr_reviewer_id ON pr_reviewers (reviewer_id);
CREATE INDEX idx_pr_status ON pull_requests (status);
