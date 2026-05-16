CREATE TABLE company_members (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL,
    user_id     UUID NOT NULL,
    status      SMALLINT NOT NULL DEFAULT 1, -- 1=PENDING, 2=APPROVED, 3=REJECTED
    note        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    reviewed_at TIMESTAMP WITH TIME ZONE NULL,
    UNIQUE (company_id, user_id)
);

CREATE INDEX idx_company_members_user ON company_members(user_id, status);
CREATE INDEX idx_company_members_pending ON company_members(company_id) WHERE status = 1;
