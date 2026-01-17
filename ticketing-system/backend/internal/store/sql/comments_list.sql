SELECT id, ticket_id, author_id, author_name, message, created_at
FROM ticket_comments
WHERE ticket_id = $1
ORDER BY created_at ASC
