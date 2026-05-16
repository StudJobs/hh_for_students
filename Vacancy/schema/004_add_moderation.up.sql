-- Модерация вакансий компанией:
--   1 = PENDING (HR-сотрудник создал, ждёт approve от owner-а),
--   2 = PUBLISHED (видна студентам),
--   3 = REJECTED.
-- По умолчанию 2 (PUBLISHED) — обратная совместимость для существующих вакансий
-- (созданы owner-ом, либо до появления модерации).
ALTER TABLE vacancies
    ADD COLUMN moderation_status SMALLINT NOT NULL DEFAULT 2,
    ADD COLUMN author_id UUID NULL,
    ADD COLUMN moderation_comment TEXT NOT NULL DEFAULT '';

CREATE INDEX idx_vacancies_moderation_pending
    ON vacancies(company_id)
    WHERE deleted_at IS NULL AND moderation_status = 1;
