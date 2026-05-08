ALTER TABLE vacancy
    ADD COLUMN skill_slugs VARCHAR(64)[] NOT NULL DEFAULT '{}'::VARCHAR[];

CREATE INDEX idx_vacancy_skill_slugs ON vacancy USING GIN (skill_slugs)
    WHERE deleted_at IS NULL;
