SELECT id, project_id, title, description, created_at, updated_at
FROM stories
WHERE id = $1
