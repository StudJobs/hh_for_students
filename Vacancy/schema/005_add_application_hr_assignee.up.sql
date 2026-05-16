-- HR-сотрудник, обрабатывающий отклик. NULL пока никто не взял в работу.
-- При первом просмотре HR'ом из его компании auto-set (через Application.AssignHR RPC).
ALTER TABLE vacancy_applications
    ADD COLUMN hr_assignee_id UUID NULL;

CREATE INDEX idx_applications_hr_assignee
    ON vacancy_applications(hr_assignee_id)
    WHERE deleted_at IS NULL AND hr_assignee_id IS NOT NULL;
