ALTER TABLE vacancies
    ADD COLUMN skill_slugs VARCHAR(64)[] NOT NULL DEFAULT '{}'::VARCHAR[];

CREATE INDEX idx_vacancies_skill_slugs ON vacancies USING GIN (skill_slugs)
    WHERE deleted_at IS NULL;
