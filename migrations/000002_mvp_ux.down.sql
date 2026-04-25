ALTER TABLE conversation_sessions
  DROP CONSTRAINT IF EXISTS conversation_sessions_status_check;
ALTER TABLE conversation_sessions
  ADD CONSTRAINT conversation_sessions_status_check CHECK (status IN ('started', 'finished', 'cancelled'));

DROP TABLE IF EXISTS conversation_side_issues;
DROP TABLE IF EXISTS conversation_notes;
DROP TABLE IF EXISTS issue_repeat_disagreements;
DROP TABLE IF EXISTS user_preferences;

ALTER TABLE pairs
  DROP COLUMN IF EXISTS is_test;

ALTER TABLE users
  DROP COLUMN IF EXISTS is_virtual;

