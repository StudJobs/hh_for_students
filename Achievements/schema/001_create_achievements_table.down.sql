-- Удаление ограничений
ALTER TABLE achievements
    DROP CONSTRAINT IF EXISTS chk_achievements_file_size_positive;

ALTER TABLE achievements
    DROP CONSTRAINT IF EXISTS chk_achievements_file_type_not_empty;

ALTER TABLE achievements
    DROP CONSTRAINT IF EXISTS chk_achievements_file_name_not_empty;

ALTER TABLE achievements
    DROP CONSTRAINT IF EXISTS chk_achievements_name_not_empty;

ALTER TABLE achievements
    DROP CONSTRAINT IF EXISTS uq_achievements_user_name;