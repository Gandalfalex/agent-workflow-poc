CREATE TABLE approval_policies (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    from_state_id uuid NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    to_state_id uuid NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    required_count int NOT NULL DEFAULT 1 CHECK (required_count >= 1),
    scope text NOT NULL CHECK (scope IN ('any_member', 'role', 'group')),
    group_id uuid REFERENCES groups(id) ON DELETE SET NULL,
    role text,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (project_id, from_state_id, to_state_id)
);
CREATE INDEX ON approval_policies (project_id);

CREATE TABLE approval_requests (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    policy_id uuid NOT NULL REFERENCES approval_policies(id) ON DELETE CASCADE,
    from_state_id uuid NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    to_state_id uuid NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    requested_by uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    requested_by_name text NOT NULL DEFAULT '',
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected','cancelled')),
    required_count int NOT NULL DEFAULT 1,
    approved_count int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    resolved_at timestamptz
);
CREATE INDEX ON approval_requests (ticket_id, status);
CREATE INDEX ON approval_requests (policy_id);

CREATE TABLE approval_decisions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id uuid NOT NULL REFERENCES approval_requests(id) ON DELETE CASCADE,
    approver_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    approver_name text NOT NULL DEFAULT '',
    decision text NOT NULL CHECK (decision IN ('approved','rejected')),
    reason text,
    decided_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ON approval_decisions (request_id);
