SELECT project_id, group_id, role
FROM project_groups
WHERE project_id = $1
ORDER BY group_id ASC
