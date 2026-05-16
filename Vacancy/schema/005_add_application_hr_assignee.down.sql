DROP INDEX IF EXISTS idx_applications_hr_assignee;
ALTER TABLE vacancy_applications DROP COLUMN IF EXISTS hr_assignee_id;
