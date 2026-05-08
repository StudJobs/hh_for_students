DROP INDEX IF EXISTS idx_achievements_verification_pending;
ALTER TABLE achievements
    DROP COLUMN IF EXISTS review_comment,
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS verification_status;
