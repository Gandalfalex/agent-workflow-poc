package store

import (
	"strings"
	"testing"
)

func TestMustSQLAllowsNameWithoutExtension(t *testing.T) {
	withoutExt := mustSQL("workflow_list", nil)
	withExt := mustSQL("workflow_list.sql", nil)

	if withoutExt != withExt {
		t.Fatalf("expected template output to match for names with and without .sql suffix")
	}
	if !strings.Contains(withoutExt, "FROM workflow_states") {
		t.Fatalf("expected rendered SQL to include workflow query body")
	}
}

func TestMustSQLRendersHelperBasedTemplate(t *testing.T) {
	query := mustSQL("tickets_get", nil)

	checks := []string{
		"SELECT",
		"FROM tickets t",
		"JOIN projects p ON p.id = t.project_id",
		"LEFT JOIN users u ON u.id = t.assignee_id",
		"WHERE t.id = $1",
	}

	for _, want := range checks {
		if !strings.Contains(query, want) {
			t.Fatalf("expected rendered SQL to contain %q", want)
		}
	}
}
