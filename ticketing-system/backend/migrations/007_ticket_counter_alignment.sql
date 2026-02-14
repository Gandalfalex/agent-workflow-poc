-- Align ticket numbering after restores/imports and remove conflicting defaults.
-- The per-project trigger/counter is the source of truth for ticket numbers.

ALTER TABLE IF EXISTS tickets
  ALTER COLUMN number DROP DEFAULT;

INSERT INTO project_ticket_counters (project_id, next_number)
SELECT t.project_id, COALESCE(MAX(t.number), 0) + 1
FROM tickets t
GROUP BY t.project_id
ON CONFLICT (project_id) DO UPDATE
SET next_number = GREATEST(project_ticket_counters.next_number, EXCLUDED.next_number);

SELECT setval(
  'ticket_number_seq',
  (SELECT COALESCE(MAX(number), 0) + 1 FROM tickets),
  false
);
