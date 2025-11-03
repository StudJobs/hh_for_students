DROP TRIGGER IF EXISTS update_profiles_updated_at ON profiles;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_profiles_age;
DROP INDEX IF EXISTS idx_profiles_name;
DROP INDEX IF EXISTS idx_profiles_created_at;
DROP INDEX IF EXISTS idx_profiles_resume_id;
DROP INDEX IF EXISTS idx_profiles_deleted_at;
DROP INDEX IF EXISTS idx_profiles_profession_category;
DROP INDEX IF EXISTS idx_profiles_email;

DROP TABLE IF EXISTS profiles;