ALTER TABLE companies
    DROP COLUMN IF EXISTS cleanup_vacancies_after_days,
    DROP COLUMN IF EXISTS cleanup_tasks_after_days;
