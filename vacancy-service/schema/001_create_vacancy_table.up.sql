CREATE TABLE vacancies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(128) NOT NULL,
    experience INTEGER NOT NULL DEFAULT 0,
    salary INTEGER NOT NULL DEFAULT 0,
    position_status VARCHAR(50) NOT NULL,
    schedule VARCHAR(64) NOT NULL DEFAULT 0,
    work_format VARCHAR(64),
    company_id UUID NOT NULL,
    attachment_id VARCHAR(255),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_vacancies_company_id ON vacancies(company_id);
CREATE INDEX idx_vacancies_position_status ON vacancies(position_status);
CREATE INDEX idx_vacancies_created_at ON vacancies(created_at);
CREATE INDEX idx_vacancies_deleted_at ON vacancies(deleted_at) WHERE deleted_at IS NULL;