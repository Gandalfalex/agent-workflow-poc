ALTER TABLE sprints
  ADD COLUMN status text NOT NULL DEFAULT 'planned'
    CHECK (status IN ('planned', 'active', 'completed')),
  ADD COLUMN completed_at timestamptz;
