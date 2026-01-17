UPDATE groups
SET name = COALESCE($2, name),
    description = COALESCE($3, description),
    updated_at = now()
WHERE id = $1
RETURNING id
