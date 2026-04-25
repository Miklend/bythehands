ALTER TABLE conversation_sessions
  DROP COLUMN IF EXISTS questions,
  DROP COLUMN IF EXISTS start_state;

