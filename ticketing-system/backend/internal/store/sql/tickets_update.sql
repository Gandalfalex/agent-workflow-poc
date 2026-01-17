UPDATE tickets
SET {{ .Updates }}
WHERE id = ${{ .IDArg }}
