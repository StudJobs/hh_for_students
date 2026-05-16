-- Файл-решение (опционально вместо/вместе с URL).
ALTER TABLE microtask_submissions
    ADD COLUMN solution_file_name VARCHAR(255) NOT NULL DEFAULT '';

-- solution_url становится опциональным (студент может прислать только файл).
ALTER TABLE microtask_submissions
    ALTER COLUMN solution_url DROP NOT NULL,
    ALTER COLUMN solution_url DROP DEFAULT;

-- Поля типа «квест от эксперта».
ALTER TABLE microtasks
    ADD COLUMN is_skill_quest    BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN target_student_id UUID NULL,
    ADD COLUMN target_skill_slug VARCHAR(64) NOT NULL DEFAULT '';

CREATE INDEX idx_microtasks_skill_quest
    ON microtasks(is_skill_quest, target_student_id)
    WHERE deleted_at IS NULL AND is_skill_quest = TRUE;
