DROP TABLE IF EXISTS conversation_rule_violations;

ALTER TABLE conversation_sessions
  DROP COLUMN IF EXISTS rule_violation_limit,
  DROP COLUMN IF EXISTS ended_initiative;

ALTER TABLE pair_members
  DROP COLUMN IF EXISTS custom_name;

ALTER TABLE issues
  DROP COLUMN IF EXISTS last_repeated_at,
  DROP COLUMN IF EXISTS repeat_limit;

