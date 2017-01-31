CREATE TABLE IF NOT EXISTS facebook_users (
  facebook_user_id VARCHAR(32) PRIMARY KEY,
  user_id          UUID REFERENCES users (id) ON DELETE CASCADE UNIQUE NOT NULL
);