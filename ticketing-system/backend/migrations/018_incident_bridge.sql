ALTER TABLE tickets
  ADD COLUMN IF NOT EXISTS incident_enabled boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS incident_severity text,
  ADD COLUMN IF NOT EXISTS incident_impact text,
  ADD COLUMN IF NOT EXISTS incident_commander_id uuid REFERENCES users(id);

CREATE TABLE IF NOT EXISTS ticket_webhook_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  webhook_id uuid REFERENCES webhooks(id) ON DELETE SET NULL,
  event text NOT NULL,
  payload jsonb NOT NULL DEFAULT '{}'::jsonb,
  delivered boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS ticket_webhook_events_ticket_id_idx
  ON ticket_webhook_events(ticket_id);
CREATE INDEX IF NOT EXISTS ticket_webhook_events_created_at_idx
  ON ticket_webhook_events(created_at DESC);
