-- Удаляем индексы
DROP INDEX IF EXISTS idx_vacancies_deleted_at;
DROP INDEX IF EXISTS idx_vacancies_created_at;
DROP INDEX IF EXISTS idx_vacancies_position_status;
DROP INDEX IF EXISTS idx_vacancies_company_id;

-- Удаляем таблицу
DROP TABLE IF EXISTS vacancies;