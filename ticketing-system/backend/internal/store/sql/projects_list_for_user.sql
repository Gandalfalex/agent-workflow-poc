SELECT DISTINCT p.id, p.key, p.name, p.description, p.created_at, p.updated_at
FROM projects p
JOIN project_groups pg ON pg.project_id = p.id
JOIN group_memberships gm ON gm.group_id = pg.group_id
WHERE gm.user_id = $1
ORDER BY p.name ASC
