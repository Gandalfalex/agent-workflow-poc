CREATE TABLE IF NOT EXISTS projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  key text NOT NULL UNIQUE,
  name text NOT NULL,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS groups (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL UNIQUE,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS group_memberships (
  group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (group_id, user_id)
);

CREATE TABLE IF NOT EXISTS project_groups (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  group_id uuid NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
  role text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, group_id),
  CONSTRAINT project_groups_role_check
    CHECK (role IN ('admin', 'contributor', 'viewer'))
);

CREATE TABLE IF NOT EXISTS project_ticket_counters (
  project_id uuid PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
  next_number integer NOT NULL DEFAULT 1
);

INSERT INTO projects (key, name, description)
VALUES ('DEMO', 'Default Project', 'Default project')
ON CONFLICT (key) DO NOTHING;

DO $$
DECLARE
  default_project_id uuid;
BEGIN
  SELECT id INTO default_project_id FROM projects WHERE key = 'DEMO' LIMIT 1;

  ALTER TABLE workflow_states ADD COLUMN IF NOT EXISTS project_id uuid;
  UPDATE workflow_states SET project_id = default_project_id WHERE project_id IS NULL;
  ALTER TABLE workflow_states ALTER COLUMN project_id SET NOT NULL;
  ALTER TABLE workflow_states
    ADD CONSTRAINT workflow_states_project_fk
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;
  ALTER TABLE workflow_states
    ADD CONSTRAINT workflow_states_project_name_unique
    UNIQUE (project_id, name);

  ALTER TABLE tickets ADD COLUMN IF NOT EXISTS project_id uuid;
  UPDATE tickets SET project_id = default_project_id WHERE project_id IS NULL;
  ALTER TABLE tickets ALTER COLUMN project_id SET NOT NULL;
  ALTER TABLE tickets
    ADD CONSTRAINT tickets_project_fk
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

  ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS project_id uuid;
  UPDATE webhooks SET project_id = default_project_id WHERE project_id IS NULL;
  ALTER TABLE webhooks ALTER COLUMN project_id SET NOT NULL;
  ALTER TABLE webhooks
    ADD CONSTRAINT webhooks_project_fk
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

  ALTER TABLE tickets DROP CONSTRAINT IF EXISTS tickets_number_unique;
  ALTER TABLE tickets
    ADD CONSTRAINT tickets_project_number_unique
    UNIQUE (project_id, number);

  WITH ordered AS (
    SELECT id,
           project_id,
           row_number() OVER (PARTITION BY project_id ORDER BY created_at, id) AS seq
    FROM tickets
  )
  UPDATE tickets t
  SET number = ordered.seq
  FROM ordered
  WHERE t.id = ordered.id;

  UPDATE tickets
    SET key = (SELECT key FROM projects WHERE id = tickets.project_id) || '-' || lpad(tickets.number::text, 3, '0');

  INSERT INTO project_ticket_counters (project_id, next_number)
  SELECT project_id, COALESCE(MAX(number), 0) + 1
  FROM tickets
  GROUP BY project_id
  ON CONFLICT (project_id) DO UPDATE
    SET next_number = EXCLUDED.next_number;
END $$;

CREATE OR REPLACE FUNCTION assign_ticket_key() RETURNS trigger AS $$
DECLARE
  project_key text;
  next_number integer;
BEGIN
  IF NEW.project_id IS NULL THEN
    RAISE EXCEPTION 'project_id required';
  END IF;

  SELECT key INTO project_key FROM projects WHERE id = NEW.project_id;
  IF project_key IS NULL THEN
    RAISE EXCEPTION 'invalid project_id';
  END IF;

  IF NEW.number IS NULL OR NEW.number = 0 THEN
    INSERT INTO project_ticket_counters (project_id, next_number)
    VALUES (NEW.project_id, 2)
    ON CONFLICT (project_id) DO UPDATE
      SET next_number = project_ticket_counters.next_number + 1
    RETURNING next_number - 1 INTO next_number;
  ELSE
    next_number := NEW.number;
    INSERT INTO project_ticket_counters (project_id, next_number)
    VALUES (NEW.project_id, NEW.number + 1)
    ON CONFLICT (project_id) DO UPDATE
      SET next_number = GREATEST(project_ticket_counters.next_number, NEW.number + 1);
  END IF;

  NEW.number := next_number;
  NEW.key := project_key || '-' || lpad(next_number::text, 3, '0');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS set_ticket_key ON tickets;
CREATE TRIGGER set_ticket_key
  BEFORE INSERT ON tickets
  FOR EACH ROW
  EXECUTE FUNCTION assign_ticket_key();
