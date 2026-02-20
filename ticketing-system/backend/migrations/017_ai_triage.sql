CREATE TABLE IF NOT EXISTS ai_triage_settings (
  project_id uuid PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
  enabled boolean NOT NULL DEFAULT false,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ai_triage_suggestions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  actor_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  input_title text NOT NULL,
  input_description text,
  input_type text,
  suggested_summary text NOT NULL,
  suggested_priority text NOT NULL,
  suggested_state_id uuid NOT NULL REFERENCES workflow_states(id) ON DELETE RESTRICT,
  suggested_assignee_id uuid REFERENCES users(id) ON DELETE SET NULL,
  confidence_summary real NOT NULL,
  confidence_priority real NOT NULL,
  confidence_state real NOT NULL,
  confidence_assignee real NOT NULL,
  prompt_version text NOT NULL,
  model text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ai_triage_suggestions_project_idx ON ai_triage_suggestions(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS ai_triage_suggestions_actor_idx ON ai_triage_suggestions(actor_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ai_triage_suggestion_decisions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  suggestion_id uuid NOT NULL REFERENCES ai_triage_suggestions(id) ON DELETE CASCADE,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  actor_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  accepted_fields text[] NOT NULL,
  rejected_fields text[] NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ai_triage_decisions_suggestion_idx ON ai_triage_suggestion_decisions(suggestion_id, created_at DESC);
