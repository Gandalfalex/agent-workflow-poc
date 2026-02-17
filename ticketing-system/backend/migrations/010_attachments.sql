CREATE TABLE ticket_attachments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  filename text NOT NULL,
  content_type text NOT NULL,
  size bigint NOT NULL,
  storage_key text NOT NULL,
  uploaded_by uuid NOT NULL REFERENCES users(id),
  uploaded_by_name text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ticket_attachments_ticket_id_idx ON ticket_attachments(ticket_id);
