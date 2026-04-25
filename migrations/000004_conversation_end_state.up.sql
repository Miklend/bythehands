ALTER TABLE conversation_sessions
  ADD COLUMN IF NOT EXISTS end_state text;

