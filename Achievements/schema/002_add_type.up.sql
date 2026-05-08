ALTER TABLE achievements
    ADD COLUMN type SMALLINT NOT NULL DEFAULT 0;

CREATE INDEX idx_achievements_type ON achievements(type) WHERE deleted_at IS NULL;
