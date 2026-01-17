SELECT COALESCE(MAX(position), 0) + 1
FROM tickets
WHERE state_id = $1
