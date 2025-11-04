-- Удаляем индексы
DROP INDEX IF EXISTS idx_companies_city;
DROP INDEX IF EXISTS idx_companies_company_type;
DROP INDEX IF EXISTS idx_companies_created_at;
DROP INDEX IF EXISTS idx_companies_deleted_at;

-- Удаляем таблицу
DROP TABLE IF EXISTS companies;