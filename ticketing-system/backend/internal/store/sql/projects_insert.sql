INSERT INTO projects (key, name, description)
VALUES ($1, $2, $3)
RETURNING id
