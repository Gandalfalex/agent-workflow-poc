CREATE TABLE ticket_locks (
    ticket_id    UUID PRIMARY KEY REFERENCES tickets(id) ON DELETE CASCADE,
    locked_by    UUID NOT NULL,
    locked_by_name TEXT NOT NULL,
    acquired_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL
);
CREATE INDEX ON ticket_locks (expires_at);
