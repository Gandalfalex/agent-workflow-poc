SELECT id, url, events, enabled, secret, created_at, updated_at
FROM webhooks
WHERE project_id = $1 AND id = $2
