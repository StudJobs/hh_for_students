-- Отклики студентов на вакансии. Живут в БД Vacancy (одна транзакция при необходимости join).
CREATE TABLE IF NOT EXISTS vacancy_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vacancy_id UUID NOT NULL,
    student_id UUID NOT NULL,
    cover_letter TEXT NOT NULL DEFAULT '',
    -- 1 = PENDING, 2 = ACCEPTED, 3 = REJECTED. Совпадает с ApplicationStatus в proto.
    status SMALLINT NOT NULL DEFAULT 1,
    hr_comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Идемпотентность Apply: студент не может откликнуться дважды на одну вакансию.
-- Частичный индекс — позволяет повторно откликнуться после withdraw (soft-delete).
CREATE UNIQUE INDEX IF NOT EXISTS idx_applications_unique_active
    ON vacancy_applications (vacancy_id, student_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_applications_vacancy
    ON vacancy_applications (vacancy_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_applications_student
    ON vacancy_applications (student_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_applications_status
    ON vacancy_applications (status)
    WHERE deleted_at IS NULL;
