-- Удаляем индексы
DROP INDEX IF EXISTS idx_pr_status;
DROP INDEX IF EXISTS idx_pr_reviewer_id;

-- Удаляем таблицы
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;

-- Удаляем расширение
DROP EXTENSION IF EXISTS "uuid-ossp";
