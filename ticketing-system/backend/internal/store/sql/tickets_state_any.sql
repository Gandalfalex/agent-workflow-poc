SELECT id
FROM workflow_states
WHERE project_id = $1
ORDER BY sort_order ASC
LIMIT 1
