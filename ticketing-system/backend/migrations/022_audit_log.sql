CREATE TABLE audit_log (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id      uuid        NOT NULL,
  actor_name    text        NOT NULL,
  action        text        NOT NULL,
  resource_type text        NOT NULL,
  resource_id   text        NOT NULL,
  resource_name text        NOT NULL DEFAULT '',
  project_id    uuid        REFERENCES projects(id) ON DELETE SET NULL,
  details       jsonb       NOT NULL DEFAULT '{}',
  created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX audit_log_actor_id_idx      ON audit_log(actor_id);
CREATE INDEX audit_log_project_id_idx    ON audit_log(project_id) WHERE project_id IS NOT NULL;
CREATE INDEX audit_log_resource_type_idx ON audit_log(resource_type);
CREATE INDEX audit_log_created_at_idx    ON audit_log(created_at DESC);
