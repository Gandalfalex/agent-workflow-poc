-- Story points on stories (capacity budget) and tickets (effort estimate)
ALTER TABLE stories ADD COLUMN IF NOT EXISTS story_points integer;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS story_points integer;
ALTER TABLE tickets ADD COLUMN IF NOT EXISTS time_estimate integer;

-- Default sprint duration on projects
ALTER TABLE projects ADD COLUMN IF NOT EXISTS default_sprint_duration_days integer;

-- Time entries for time tracking on tickets
CREATE TABLE IF NOT EXISTS time_entries (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_name text NOT NULL,
  minutes integer NOT NULL,
  description text,
  logged_at date NOT NULL DEFAULT CURRENT_DATE,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT time_entries_minutes_positive CHECK (minutes > 0)
);
CREATE INDEX IF NOT EXISTS time_entries_ticket_id_idx ON time_entries(ticket_id);
