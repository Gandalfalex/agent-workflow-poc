UPDATE stories
SET title = COALESCE($2, title),
    description = COALESCE($3, description),
    updated_at = now()
WHERE id = $1
RETURNING id
