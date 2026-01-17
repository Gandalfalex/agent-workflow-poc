SELECT t.id, t.project_id, p.key, t.key, t.number, t.type, t.story_id, s2.title, s2.description, s2.created_at, s2.updated_at,
       t.title, t.description, t.state_id, t.assignee_id,
       t.priority, t.position, t.created_at, t.updated_at,
       s.name, s.sort_order, s.is_default, s.is_closed,
       u.name
FROM tickets t
JOIN projects p ON p.id = t.project_id
JOIN workflow_states s ON s.id = t.state_id
LEFT JOIN stories s2 ON s2.id = t.story_id
LEFT JOIN users u ON u.id = t.assignee_id
{{- if .Where }}
WHERE {{ .Where }}
{{- end }}
ORDER BY s.sort_order ASC, t.position ASC
LIMIT ${{ .LimitArg }} OFFSET ${{ .OffsetArg }}
