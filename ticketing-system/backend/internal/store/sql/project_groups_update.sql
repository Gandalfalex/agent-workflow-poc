UPDATE project_groups
SET role = $3
WHERE project_id = $1 AND group_id = $2
RETURNING project_id, group_id, role
