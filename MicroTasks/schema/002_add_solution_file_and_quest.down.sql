DROP INDEX IF EXISTS idx_microtasks_skill_quest;
ALTER TABLE microtasks
    DROP COLUMN IF EXISTS target_skill_slug,
    DROP COLUMN IF EXISTS target_student_id,
    DROP COLUMN IF EXISTS is_skill_quest;

ALTER TABLE microtask_submissions
    ALTER COLUMN solution_url SET DEFAULT '',
    ALTER COLUMN solution_url SET NOT NULL;
ALTER TABLE microtask_submissions
    DROP COLUMN IF EXISTS solution_file_name;
