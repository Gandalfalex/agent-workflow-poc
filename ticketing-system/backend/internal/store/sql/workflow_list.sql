SELECT id, name, sort_order, is_default, is_closed, created_at, updated_at
FROM workflow_states
WHERE project_id = $1
ORDER BY sort_order ASC
