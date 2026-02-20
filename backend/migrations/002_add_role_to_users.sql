-- Agrega columna role a users y soporta roles: USER, AGENT, ADMIN
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role VARCHAR(10) NOT NULL DEFAULT 'USER';

ALTER TABLE users
    DROP CONSTRAINT IF EXISTS chk_role;

ALTER TABLE users
    ADD CONSTRAINT chk_role CHECK (role IN ('USER', 'AGENT', 'ADMIN'));
