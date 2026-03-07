CREATE TABLE releases (
    id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id    uuid        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name          text        NOT NULL,
    version       text,
    status        text        NOT NULL DEFAULT 'planned'
                              CHECK (status IN ('planned','in_progress','released','archived')),
    target_date   date,
    notes         text,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX ON releases (project_id, status);

ALTER TABLE tickets ADD COLUMN release_id uuid REFERENCES releases(id) ON DELETE SET NULL;
CREATE INDEX ON tickets (release_id) WHERE release_id IS NOT NULL;
