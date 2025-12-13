-- Users table
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY, -- Clerk ID
  email TEXT NOT NULL,
  name TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Teams table (Multi-tenancy)
CREATE TABLE IF NOT EXISTS teams (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  owner_id TEXT NOT NULL REFERENCES users(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Team Members
CREATE TABLE IF NOT EXISTS team_members (
  team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
  user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL, -- 'owner', 'member'
  PRIMARY KEY (team_id, user_id)
);

-- AWS Accounts
CREATE TABLE IF NOT EXISTS aws_accounts (
  id SERIAL PRIMARY KEY,
  team_id INTEGER REFERENCES teams(id) ON DELETE CASCADE,
  account_id TEXT NOT NULL,
  external_id TEXT NOT NULL,
  role_arn TEXT,
  verified BOOLEAN DEFAULT FALSE,
  last_verified_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(account_id, team_id)
);

-- Scans
CREATE TABLE IF NOT EXISTS scans (
  id SERIAL PRIMARY KEY,
  aws_account_id INTEGER REFERENCES aws_accounts(id) ON DELETE CASCADE,
  status TEXT NOT NULL, -- 'pending', 'in_progress', 'completed', 'failed'
  services TEXT[], -- Array of services scanned
  regions TEXT[], -- Array of regions scanned
  overall_score INTEGER,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Scan Findings
CREATE TABLE IF NOT EXISTS scan_findings (
  id SERIAL PRIMARY KEY,
  scan_id INTEGER REFERENCES scans(id) ON DELETE CASCADE,
  service TEXT NOT NULL,
  region TEXT NOT NULL,
  resource_id TEXT NOT NULL,
  resource_arn TEXT,
  check_id TEXT NOT NULL,
  status TEXT NOT NULL, -- 'PASS', 'FAIL'
  severity TEXT NOT NULL, -- 'LOW', 'MEDIUM', 'HIGH', 'CRITICAL'
  title TEXT NOT NULL,
  description TEXT,
  compliance TEXT[], -- Array of compliance frameworks
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Chat Conversations
CREATE TABLE IF NOT EXISTS chat_conversations (
  id SERIAL PRIMARY KEY,
  aws_account_id INTEGER REFERENCES aws_accounts(id) ON DELETE CASCADE,
  user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
  title TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Chat Messages
CREATE TABLE IF NOT EXISTS chat_messages (
  id SERIAL PRIMARY KEY,
  conversation_id INTEGER REFERENCES chat_conversations(id) ON DELETE CASCADE,
  role TEXT NOT NULL, -- 'user', 'assistant'
  content TEXT NOT NULL,
  tool_calls JSONB, -- For storing tool calls/results
  scan_context_id INTEGER REFERENCES scans(id),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
