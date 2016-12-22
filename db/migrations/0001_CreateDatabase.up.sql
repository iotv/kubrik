BEGIN;
CREATE EXTENSION "uuid-ossp";
CREATE TABLE IF NOT EXISTS users (
	id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	username           VARCHAR(31) UNIQUE,
	email              VARCHAR(255) UNIQUE,
	encrypted_password BYTEA);
COMMIT;
