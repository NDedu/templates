DROP INDEX IF EXISTS idx_users_email;
CREATE INDEX idx_users_email ON users (email);
