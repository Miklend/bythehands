-- UX/MVP extensions: test mode, preferences, repeats disagreements, conversations notes/side-issues, pause.

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS is_virtual boolean NOT NULL DEFAULT false;

ALTER TABLE pairs
  ADD COLUMN IF NOT EXISTS is_test boolean NOT NULL DEFAULT false;

-- conversation status: add paused
DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'conversation_sessions_status_check'
  ) THEN
    ALTER TABLE conversation_sessions DROP CONSTRAINT conversation_sessions_status_check;
  END IF;
END $$;

ALTER TABLE conversation_sessions
  ADD CONSTRAINT conversation_sessions_status_check CHECK (status IN ('started', 'paused', 'finished', 'cancelled'));

CREATE TABLE IF NOT EXISTS user_preferences (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  current_pair_id uuid REFERENCES pairs(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_user_preferences_current_pair_id ON user_preferences(current_pair_id);

CREATE TABLE IF NOT EXISTS issue_repeat_disagreements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  repeat_id uuid NOT NULL REFERENCES issue_repeats(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  note text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE (repeat_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_issue_repeat_disagreements_repeat_id ON issue_repeat_disagreements(repeat_id);

CREATE TABLE IF NOT EXISTS conversation_notes (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id uuid NOT NULL REFERENCES conversation_sessions(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  text text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_conversation_notes_conversation_id ON conversation_notes(conversation_id);

CREATE TABLE IF NOT EXISTS conversation_side_issues (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id uuid NOT NULL REFERENCES conversation_sessions(id) ON DELETE CASCADE,
  issue_id uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  UNIQUE (conversation_id, issue_id)
);
CREATE INDEX IF NOT EXISTS idx_conversation_side_issues_conversation_id ON conversation_side_issues(conversation_id);

