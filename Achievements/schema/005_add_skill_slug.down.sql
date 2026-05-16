DROP INDEX IF EXISTS idx_achievements_skill_slug;
ALTER TABLE achievements DROP COLUMN IF EXISTS skill_slug;
