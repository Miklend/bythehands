CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  telegram_id bigint UNIQUE,
  username text,
  display_name text,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pairs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  status text NOT NULL CHECK (status IN ('active', 'archived')),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS pair_members (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pair_id uuid NOT NULL REFERENCES pairs(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('creator', 'partner')),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE (pair_id, user_id),
  UNIQUE (pair_id, role)
);
CREATE INDEX IF NOT EXISTS idx_pair_members_pair_id ON pair_members(pair_id);

CREATE TABLE IF NOT EXISTS invites (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pair_id uuid NOT NULL REFERENCES pairs(id) ON DELETE CASCADE,
  token text NOT NULL UNIQUE,
  status text NOT NULL CHECK (status IN ('active', 'used', 'expired')),
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  used_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_invites_pair_id ON invites(pair_id);
CREATE INDEX IF NOT EXISTS idx_invites_token ON invites(token);

CREATE TABLE IF NOT EXISTS issues (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  pair_id uuid NOT NULL REFERENCES pairs(id) ON DELETE CASCADE,
  created_by_user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  title text NOT NULL,
  description text NOT NULL,
  priority text NOT NULL CHECK (priority IN ('low', 'medium', 'high')),
  visibility text NOT NULL CHECK (visibility IN ('visible', 'hidden_until_repeats', 'private')),
  repeat_threshold int NOT NULL DEFAULT 0,
  repeat_count int NOT NULL DEFAULT 0,
  status text NOT NULL CHECK (status IN ('active', 'resolved', 'postponed', 'archived')),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW(),
  resolved_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_issues_pair_id ON issues(pair_id);
CREATE INDEX IF NOT EXISTS idx_issues_pair_id_status ON issues(pair_id, status);

CREATE TABLE IF NOT EXISTS issue_repeats (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  issue_id uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  note text,
  created_at timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_issue_repeats_issue_id ON issue_repeats(issue_id);
CREATE INDEX IF NOT EXISTS idx_issue_repeats_user_id ON issue_repeats(user_id);

CREATE TABLE IF NOT EXISTS conversation_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  issue_id uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  pair_id uuid NOT NULL REFERENCES pairs(id) ON DELETE CASCADE,
  status text NOT NULL CHECK (status IN ('started', 'finished', 'cancelled')),
  goal text,
  result_status text CHECK (result_status IN ('resolved', 'partially_resolved', 'postponed', 'unresolved')),
  result_text text,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  finished_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_conversation_sessions_issue_id ON conversation_sessions(issue_id);
CREATE INDEX IF NOT EXISTS idx_conversation_sessions_pair_id ON conversation_sessions(pair_id);

