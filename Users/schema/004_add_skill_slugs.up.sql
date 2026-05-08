ALTER TABLE profiles
    ADD COLUMN skill_slugs VARCHAR(64)[] NOT NULL DEFAULT '{}'::VARCHAR[];

CREATE INDEX idx_profiles_skill_slugs ON profiles USING GIN (skill_slugs)
    WHERE deleted_at IS NULL;
