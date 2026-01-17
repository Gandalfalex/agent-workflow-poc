INSERT INTO project_groups (project_id, group_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (project_id, group_id) DO UPDATE SET role = EXCLUDED.role
RETURNING project_id, group_id, role
