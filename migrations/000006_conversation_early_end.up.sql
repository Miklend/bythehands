ALTER TABLE conversation_sessions
  ADD COLUMN IF NOT EXISTS ended_early boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS ended_by_user_id uuid REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS end_reason text;

