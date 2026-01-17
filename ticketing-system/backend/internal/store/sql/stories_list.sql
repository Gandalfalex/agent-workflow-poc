SELECT id, project_id, title, description, created_at, updated_at
FROM stories
WHERE project_id = $1
ORDER BY created_at DESC
