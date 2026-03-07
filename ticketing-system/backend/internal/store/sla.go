package store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type SlaPolicyEntry struct {
	Priority           string
	TicketType         string
	FirstResponseHours int
	CompletionHours    int
	AtRiskThresholdPct int
}

// ApplySlaStatus computes SlaStatus and SlaBreachAt for each ticket using the
// provided policy map (keyed by "priority:type"). Open tickets only.
func ApplySlaStatus(tickets []Ticket, policies []SlaPolicyEntry) []Ticket {
	policyMap := make(map[string]SlaPolicyEntry, len(policies))
	for _, p := range policies {
		policyMap[p.Priority+":"+p.TicketType] = p
	}

	now := time.Now()
	for i := range tickets {
		t := &tickets[i]
		if t.StateClosed {
			continue
		}
		pol, ok := policyMap[t.Priority+":"+t.Type]
		if !ok {
			continue
		}
		breachAt := t.CreatedAt.Add(time.Duration(pol.CompletionHours) * time.Hour)
		t.SlaBreachAt = &breachAt

		atRiskAt := t.CreatedAt.Add(time.Duration(float64(pol.CompletionHours)*float64(pol.AtRiskThresholdPct)/100.0) * time.Hour)
		var status string
		switch {
		case now.After(breachAt):
			status = "breached"
		case now.After(atRiskAt):
			status = "at_risk"
		default:
			status = "on_track"
		}
		t.SlaStatus = &status
	}
	return tickets
}

func (s *Store) GetSlaPolicies(ctx context.Context, projectID uuid.UUID) ([]SlaPolicyEntry, error) {
	return queryMany(ctx, s.db, mustSQL("sla_policies_by_project", nil), scanSlaPolicy, projectID)
}

func (s *Store) UpdateSlaPolicies(ctx context.Context, projectID uuid.UUID, entries []SlaPolicyEntry) ([]SlaPolicyEntry, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if _, err := tx.Exec(ctx, mustSQL("sla_policies_delete_by_project", nil), projectID); err != nil {
		return nil, err
	}

	upsertSQL := mustSQL("sla_policies_upsert", nil)
	for _, e := range entries {
		if _, err := tx.Exec(ctx, upsertSQL,
			projectID,
			e.Priority,
			e.TicketType,
			e.FirstResponseHours,
			e.CompletionHours,
			e.AtRiskThresholdPct,
		); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return s.GetSlaPolicies(ctx, projectID)
}

func scanSlaPolicy(row pgx.Row) (SlaPolicyEntry, error) {
	var e SlaPolicyEntry
	err := row.Scan(&e.Priority, &e.TicketType, &e.FirstResponseHours, &e.CompletionHours, &e.AtRiskThresholdPct)
	return e, err
}
