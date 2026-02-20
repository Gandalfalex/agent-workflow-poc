CREATE TABLE IF NOT EXISTS sprints (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name text NOT NULL,
  goal text,
  start_date date NOT NULL,
  end_date date NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT sprints_date_range_check CHECK (end_date >= start_date)
);

CREATE INDEX IF NOT EXISTS sprints_project_id_idx ON sprints(project_id, start_date DESC, created_at DESC);

CREATE TABLE IF NOT EXISTS sprint_tickets (
  sprint_id uuid NOT NULL REFERENCES sprints(id) ON DELETE CASCADE,
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (sprint_id, ticket_id)
);

CREATE INDEX IF NOT EXISTS sprint_tickets_ticket_id_idx ON sprint_tickets(ticket_id);

CREATE TABLE IF NOT EXISTS capacity_settings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  scope text NOT NULL,
  user_id uuid REFERENCES users(id) ON DELETE CASCADE,
  label text NOT NULL,
  capacity integer NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT capacity_settings_scope_check CHECK (scope IN ('team', 'user')),
  CONSTRAINT capacity_settings_capacity_nonnegative CHECK (capacity >= 0),
  CONSTRAINT capacity_settings_user_scope_check CHECK (
    (scope = 'user' AND user_id IS NOT NULL) OR
    (scope = 'team' AND user_id IS NULL)
  )
);

CREATE INDEX IF NOT EXISTS capacity_settings_project_id_idx ON capacity_settings(project_id, scope, label);
