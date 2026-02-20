CREATE TABLE IF NOT EXISTS board_filter_presets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  owner_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name text NOT NULL,
  filters jsonb NOT NULL DEFAULT '{}'::jsonb,
  share_token text UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT board_filter_presets_name_nonempty CHECK (length(trim(name)) > 0)
);

CREATE INDEX IF NOT EXISTS idx_board_filter_presets_owner_project
  ON board_filter_presets (owner_id, project_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS idx_board_filter_presets_share_token
  ON board_filter_presets (share_token)
  WHERE share_token IS NOT NULL;
