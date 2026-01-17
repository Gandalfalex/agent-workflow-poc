SELECT id, url, events, enabled, secret, created_at, updated_at
FROM webhooks
WHERE project_id = $1 AND enabled = true AND events ? $2
ORDER BY created_at DESC
