CREATE TABLE IF NOT EXISTS notifications (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  type text NOT NULL CHECK (type IN ('mention', 'assignment')),
  message text NOT NULL,
  read_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS notifications_user_project_created_idx
  ON notifications(user_id, project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS notifications_user_project_unread_idx
  ON notifications(user_id, project_id, read_at)
  WHERE read_at IS NULL;

CREATE TABLE IF NOT EXISTS notification_preferences (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  mention_enabled boolean NOT NULL DEFAULT true,
  assignment_enabled boolean NOT NULL DEFAULT true,
  updated_at timestamptz NOT NULL DEFAULT now()
);
