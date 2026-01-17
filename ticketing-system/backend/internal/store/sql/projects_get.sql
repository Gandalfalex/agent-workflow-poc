SELECT id, key, name, description, created_at, updated_at
FROM projects
WHERE id = $1
