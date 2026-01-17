INSERT INTO ticket_comments (ticket_id, author_id, author_name, message)
VALUES ($1, $2, $3, $4)
RETURNING id
