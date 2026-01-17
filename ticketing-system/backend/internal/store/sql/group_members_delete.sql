DELETE FROM group_memberships
WHERE group_id = $1 AND user_id = $2
