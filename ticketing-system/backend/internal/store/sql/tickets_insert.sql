INSERT INTO tickets (project_id, title, description, type, story_id, state_id, assignee_id, priority, position)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id
