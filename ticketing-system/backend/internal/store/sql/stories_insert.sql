INSERT INTO stories (project_id, title, description)
VALUES ($1, $2, $3)
RETURNING id
