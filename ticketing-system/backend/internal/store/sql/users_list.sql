SELECT id, name, email
FROM users
{{- if .Where }}
WHERE {{ .Where }}
{{- end }}
ORDER BY name ASC
