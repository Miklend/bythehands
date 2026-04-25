ALTER TABLE conversation_sessions
  ADD COLUMN IF NOT EXISTS questions text,
  ADD COLUMN IF NOT EXISTS start_state text;

