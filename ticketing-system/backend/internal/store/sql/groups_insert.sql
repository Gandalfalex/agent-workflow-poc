INSERT INTO groups (name, description)
VALUES ($1, $2)
RETURNING id
