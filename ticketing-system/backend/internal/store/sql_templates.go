package store

import (
	"bytes"
	"embed"
	"strings"
	"sync"
	"text/template"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

var (
	sqlTemplates *template.Template
	sqlOnce      sync.Once
)

func mustSQL(name string, data any) string {
	sqlOnce.Do(func() {
		sqlTemplates = template.Must(template.New("sql").Funcs(template.FuncMap{
			"trim": strings.TrimSpace,
		}).ParseFS(sqlFiles, "sql/*.sql"))
	})

	var buf bytes.Buffer
	if err := sqlTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}
