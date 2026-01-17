UPDATE webhooks
SET {{ .Updates }}
WHERE project_id = ${{ .ProjectArg }} AND id = ${{ .IDArg }}
