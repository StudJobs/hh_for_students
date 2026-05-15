ALTER TABLE companies
    ADD COLUMN cleanup_vacancies_after_days INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN cleanup_tasks_after_days INTEGER NOT NULL DEFAULT 0;
