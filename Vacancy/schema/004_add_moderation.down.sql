DROP INDEX IF EXISTS idx_vacancies_moderation_pending;
ALTER TABLE vacancies
    DROP COLUMN IF EXISTS moderation_comment,
    DROP COLUMN IF EXISTS author_id,
    DROP COLUMN IF EXISTS moderation_status;
