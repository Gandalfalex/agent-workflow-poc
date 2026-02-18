CREATE TABLE ticket_activities (
  id         uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id  uuid        NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  actor_id   uuid        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  actor_name text        NOT NULL,
  action     text        NOT NULL,
  field      text,
  old_value  text,
  new_value  text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ticket_activities_ticket_id_idx ON ticket_activities(ticket_id);
