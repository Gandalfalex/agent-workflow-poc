-- Fix ambiguous column reference in assign_ticket_key function
-- The issue: variable 'next_number' conflicts with column name 'next_number'

CREATE OR REPLACE FUNCTION assign_ticket_key() RETURNS trigger AS $$
DECLARE
  project_key text;
  assigned_number integer;
BEGIN
  IF NEW.project_id IS NULL THEN
    RAISE EXCEPTION 'project_id required';
  END IF;

  SELECT key INTO project_key FROM projects WHERE id = NEW.project_id;
  IF project_key IS NULL THEN
    RAISE EXCEPTION 'invalid project_id';
  END IF;

  IF NEW.number IS NULL OR NEW.number = 0 THEN
    INSERT INTO project_ticket_counters (project_id, next_number)
    VALUES (NEW.project_id, 2)
    ON CONFLICT (project_id) DO UPDATE
      SET next_number = project_ticket_counters.next_number + 1
    RETURNING project_ticket_counters.next_number - 1 INTO assigned_number;
  ELSE
    assigned_number := NEW.number;
    INSERT INTO project_ticket_counters (project_id, next_number)
    VALUES (NEW.project_id, NEW.number + 1)
    ON CONFLICT (project_id) DO UPDATE
      SET next_number = GREATEST(project_ticket_counters.next_number, NEW.number + 1);
  END IF;

  NEW.number := assigned_number;
  NEW.key := project_key || '-' || lpad(assigned_number::text, 3, '0');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
