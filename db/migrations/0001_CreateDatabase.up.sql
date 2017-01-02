CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE IF NOT EXISTS users (
	id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	username           VARCHAR(31) UNIQUE,
	email              VARCHAR(255) UNIQUE NOT NULL,
	encrypted_password BYTEA);
