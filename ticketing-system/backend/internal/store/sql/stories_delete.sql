WITH target AS (
  SELECT id
  FROM stories
  WHERE id = $1
),
deleted_tickets AS (
  DELETE FROM tickets
  WHERE story_id IN (SELECT id FROM target)
)
DELETE FROM stories
WHERE id IN (SELECT id FROM target)
