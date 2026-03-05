CREATE TABLE automation_rules (
  id                  uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id          uuid        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name                text        NOT NULL,
  enabled             boolean     NOT NULL DEFAULT true,
  trigger_event       text        NOT NULL,
  trigger_conditions  jsonb       NOT NULL DEFAULT '{}',
  actions             jsonb       NOT NULL DEFAULT '[]',
  execution_count     int         NOT NULL DEFAULT 0,
  last_executed_at    timestamptz,
  created_at          timestamptz NOT NULL DEFAULT now(),
  updated_at          timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX automation_rules_project_enabled_idx ON automation_rules(project_id) WHERE enabled = true;

CREATE TABLE automation_executions (
  id            uuid        PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_id       uuid        NOT NULL REFERENCES automation_rules(id) ON DELETE CASCADE,
  ticket_id     uuid        NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  trigger_event text        NOT NULL,
  actions_run   jsonb       NOT NULL DEFAULT '[]',
  created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX automation_executions_rule_id_idx    ON automation_executions(rule_id);
CREATE INDEX automation_executions_ticket_id_idx  ON automation_executions(ticket_id);
