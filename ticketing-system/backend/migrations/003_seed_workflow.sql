INSERT INTO workflow_states (name, sort_order, is_default, is_closed)
SELECT name, sort_order, is_default, is_closed
FROM (
  VALUES
    ('Backlog', 1, true, false),
    ('In Progress', 2, false, false),
    ('Review', 3, false, false),
    ('Done', 4, false, true)
) AS seed(name, sort_order, is_default, is_closed)
WHERE NOT EXISTS (SELECT 1 FROM workflow_states);
