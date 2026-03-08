CREATE TABLE simulation_runs (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id  uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    rule_id     uuid REFERENCES automation_rules(id) ON DELETE SET NULL,
    rule_snapshot jsonb NOT NULL,
    lookback_days int NOT NULL,
    total_events  int NOT NULL DEFAULT 0,
    matched_count int NOT NULL DEFAULT 0,
    failure_count int NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE simulation_results (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id           uuid NOT NULL REFERENCES simulation_runs(id) ON DELETE CASCADE,
    ticket_id        uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    ticket_key       text NOT NULL,
    event_type       text NOT NULL,
    event_timestamp  timestamptz NOT NULL,
    matched          bool NOT NULL,
    failure_reason   text,
    predicted_actions jsonb NOT NULL DEFAULT '[]'
);

CREATE INDEX simulation_results_run_id_idx ON simulation_results (run_id);
