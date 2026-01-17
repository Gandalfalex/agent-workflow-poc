SELECT id, name, description, created_at, updated_at
FROM groups
WHERE id = $1
