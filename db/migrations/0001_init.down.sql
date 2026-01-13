DROP TRIGGER IF EXISTS users_updated_at ON users;
DROP TRIGGER IF EXISTS subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS plans_updated_at ON plans;
DROP TRIGGER IF EXISTS students_updated_at ON students;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS plans;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS users;
