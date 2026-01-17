SELECT gm.group_id, gm.user_id, u.name
FROM group_memberships gm
LEFT JOIN users u ON u.id = gm.user_id
WHERE gm.group_id = $1
ORDER BY u.name ASC NULLS LAST
