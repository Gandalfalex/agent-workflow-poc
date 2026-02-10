package store

import (
	"bytes"
	"embed"
	"strings"
	"sync"
	"text/template"
)

//go:embed sql/*.go.templ
var sqlFiles embed.FS

var (
	sqlTemplates *template.Template
	sqlOnce      sync.Once
)

func mustSQL(name string, data any) string {
	sqlOnce.Do(func() {
		sqlTemplates = template.Must(template.New("sql").Funcs(template.FuncMap{
			"trim": strings.TrimSpace,
		}).ParseFS(sqlFiles, "sql/*.go.templ"))
	})

	var buf bytes.Buffer
	templateName := strings.TrimSpace(name)
	if !strings.HasSuffix(templateName, ".sql") {
		templateName += ".sql"
	}
	if err := sqlTemplates.ExecuteTemplate(&buf, templateName, data); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}
