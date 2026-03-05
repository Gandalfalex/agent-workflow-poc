-- Performance indexes for portfolio query at 50+ projects
CREATE INDEX IF NOT EXISTS idx_sprints_project_status
  ON sprints (project_id, status);

CREATE INDEX IF NOT EXISTS idx_ticket_deps_blocked_ticket
  ON ticket_dependencies (to_ticket_id);

CREATE INDEX IF NOT EXISTS idx_tickets_project_state
  ON tickets (project_id, state_id);
