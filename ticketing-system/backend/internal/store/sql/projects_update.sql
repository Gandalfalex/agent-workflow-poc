UPDATE projects
SET {{ .Updates }}
WHERE id = ${{ .IDArg }}
