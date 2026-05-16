ALTER TABLE achievements
    ADD COLUMN skill_slug VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX idx_achievements_skill_slug
    ON achievements(skill_slug)
    WHERE deleted_at IS NULL AND skill_slug <> '';
