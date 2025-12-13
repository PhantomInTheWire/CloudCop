-- Users table with external_id for STS AssumeRole
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  full_name TEXT NOT NULL,
  external_id UUID NOT NULL DEFAULT gen_random_uuid(),
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Account connections linking user accounts to AWS accounts
CREATE TABLE IF NOT EXISTS account_connections (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  external_id UUID NOT NULL,
  account_id TEXT NOT NULL,
  account_arn TEXT,
  verified BOOLEAN NOT NULL DEFAULT false,
  last_verified_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, account_id)
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_account_connections_user_id ON account_connections(user_id);
CREATE INDEX IF NOT EXISTS idx_account_connections_account_id ON account_connections(account_id);
