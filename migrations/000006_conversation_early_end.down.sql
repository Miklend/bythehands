ALTER TABLE conversation_sessions
  DROP COLUMN IF EXISTS end_reason,
  DROP COLUMN IF EXISTS ended_by_user_id,
  DROP COLUMN IF EXISTS ended_early;

