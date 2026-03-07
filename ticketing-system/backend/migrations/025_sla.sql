-- TKT-029: SLA targets, due dates, and breach escalation

ALTER TABLE tickets ADD COLUMN IF NOT EXISTS due_date TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS sla_policies (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  UUID        NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    priority    TEXT        NOT NULL CHECK (priority IN ('low','medium','high','urgent')),
    ticket_type TEXT        NOT NULL CHECK (ticket_type IN ('feature','bug')),
    first_response_hours  INT NOT NULL DEFAULT 24,
    completion_hours      INT NOT NULL DEFAULT 72,
    at_risk_threshold_pct INT NOT NULL DEFAULT 80,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, priority, ticket_type)
);

CREATE INDEX IF NOT EXISTS idx_sla_policies_project ON sla_policies(project_id);
