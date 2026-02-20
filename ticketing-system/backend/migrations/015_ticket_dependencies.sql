CREATE TABLE IF NOT EXISTS ticket_dependencies (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  from_ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  to_ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  relation_type text NOT NULL CHECK (relation_type IN ('blocks', 'related')),
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CHECK (from_ticket_id <> to_ticket_id),
  UNIQUE (from_ticket_id, to_ticket_id, relation_type)
);

CREATE INDEX IF NOT EXISTS ticket_dependencies_project_idx
  ON ticket_dependencies(project_id);

CREATE INDEX IF NOT EXISTS ticket_dependencies_from_idx
  ON ticket_dependencies(from_ticket_id, relation_type);

CREATE INDEX IF NOT EXISTS ticket_dependencies_to_idx
  ON ticket_dependencies(to_ticket_id, relation_type);
