SELECT COUNT(*)
FROM tickets t
{{- if .Where }}
WHERE {{ .Where }}
{{- end }}
