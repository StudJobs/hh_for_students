CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE microtasks (
    id           UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    company_id   UUID NOT NULL,
    title        VARCHAR(255) NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    reward       INTEGER NOT NULL DEFAULT 0,
    deadline     DATE NULL,
    skill_slugs  VARCHAR(64)[] NOT NULL DEFAULT '{}',
    status       SMALLINT NOT NULL DEFAULT 1, -- 1 = OPEN
    assigned_to  UUID NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at   TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_microtasks_company        ON microtasks(company_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_microtasks_status         ON microtasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_microtasks_assigned_to    ON microtasks(assigned_to) WHERE deleted_at IS NULL AND assigned_to IS NOT NULL;
CREATE INDEX idx_microtasks_skill_slugs    ON microtasks USING GIN (skill_slugs);
CREATE INDEX idx_microtasks_title_trgm     ON microtasks USING GIN (title gin_trgm_ops);

CREATE TABLE microtask_submissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    microtask_id    UUID NOT NULL REFERENCES microtasks(id) ON DELETE CASCADE,
    student_id      UUID NOT NULL,
    solution_url    VARCHAR(512) NOT NULL,
    comment         TEXT NOT NULL DEFAULT '',
    status          SMALLINT NOT NULL DEFAULT 1, -- 1 = PENDING
    review_comment  TEXT NOT NULL DEFAULT '',
    submitted_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reviewed_at     TIMESTAMP WITH TIME ZONE NULL
);

CREATE INDEX idx_submissions_microtask  ON microtask_submissions(microtask_id);
CREATE INDEX idx_submissions_student    ON microtask_submissions(student_id);
CREATE INDEX idx_submissions_status     ON microtask_submissions(status);
