ALTER TABLE issues
  ADD COLUMN IF NOT EXISTS repeat_limit int NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_repeated_at timestamptz;

ALTER TABLE pair_members
  ADD COLUMN IF NOT EXISTS custom_name text;

ALTER TABLE conversation_sessions
  ADD COLUMN IF NOT EXISTS ended_initiative text CHECK (ended_initiative IN ('self', 'partner', 'both')),
  ADD COLUMN IF NOT EXISTS rule_violation_limit int NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS conversation_rule_violations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id uuid NOT NULL REFERENCES conversation_sessions(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  rule_code text NOT NULL,
  note text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_conv_rule_violations_conversation_id ON conversation_rule_violations(conversation_id);
CREATE INDEX IF NOT EXISTS idx_conv_rule_violations_created_at ON conversation_rule_violations(created_at);

