SELECT id
FROM workflow_states
WHERE project_id = $1 AND is_default = true
ORDER BY sort_order ASC
LIMIT 1
