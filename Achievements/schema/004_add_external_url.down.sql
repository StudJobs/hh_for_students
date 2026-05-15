ALTER TABLE achievements
    DROP COLUMN IF EXISTS external_url,
    DROP COLUMN IF EXISTS description;
