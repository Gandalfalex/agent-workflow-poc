SELECT DISTINCT pg.project_id
FROM project_groups pg
JOIN group_memberships gm ON gm.group_id = pg.group_id
WHERE gm.user_id = $1
