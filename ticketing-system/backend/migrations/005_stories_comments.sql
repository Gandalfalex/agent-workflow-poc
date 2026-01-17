CREATE TABLE IF NOT EXISTS stories (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title text NOT NULL,
  description text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE tickets
  ADD COLUMN IF NOT EXISTS type text NOT NULL DEFAULT 'feature';

ALTER TABLE tickets
  ADD CONSTRAINT tickets_type_check
  CHECK (type IN ('feature', 'bug'));

ALTER TABLE tickets
  ADD COLUMN IF NOT EXISTS story_id uuid REFERENCES stories(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS tickets_story_id_idx ON tickets(story_id);

CREATE TABLE IF NOT EXISTS ticket_comments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  author_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  author_name text NOT NULL,
  message text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ticket_comments_ticket_id_idx ON ticket_comments(ticket_id);
