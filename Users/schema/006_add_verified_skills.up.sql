ALTER TABLE profiles
    ADD COLUMN verified_skill_slugs VARCHAR(64)[] NOT NULL DEFAULT '{}';
