package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"ticketing-system/backend/internal/store"
)

func (h *API) ListReleases(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleViewer) {
		return
	}
	releases, err := h.store.ListReleases(r.Context(), projectID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "releases_list_failed", "failed to list releases")
		return
	}
	writeJSON(w, http.StatusOK, releaseListResponse{Items: mapReleases(releases)})
}

func (h *API) CreateRelease(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleContributor) {
		return
	}
	req, ok := decodeJSON[releaseCreateRequest](w, r, "release_create")
	if !ok {
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusBadRequest, "invalid_release", "name is required")
		return
	}
	input := store.ReleaseCreateInput{
		Name:  req.Name,
		Notes: req.Notes,
	}
	if req.Version != nil {
		input.Version = req.Version
	}
	if req.Status != nil {
		input.Status = string(*req.Status)
	}
	if req.TargetDate != nil {
		d := openapi_types.Date{Time: req.TargetDate.Time}
		input.TargetDate = &d
	}
	release, err := h.store.CreateRelease(r.Context(), projectID, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "release_create_failed", "failed to create release")
		return
	}
	writeJSON(w, http.StatusCreated, mapRelease(release))
}

func (h *API) GetRelease(w http.ResponseWriter, r *http.Request, projectId, releaseId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleViewer) {
		return
	}
	release, err := h.store.GetRelease(r.Context(), uuid.UUID(releaseId), projectID)
	if handleDBError(w, r, err, "release", "release_get") {
		return
	}
	writeJSON(w, http.StatusOK, mapRelease(release))
}

func (h *API) UpdateRelease(w http.ResponseWriter, r *http.Request, projectId, releaseId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleContributor) {
		return
	}
	req, ok := decodeJSON[releaseUpdateRequest](w, r, "release_update")
	if !ok {
		return
	}
	input := store.ReleaseUpdateInput{}
	if req.Name != nil {
		input.Name = req.Name
	}
	if req.Version != nil {
		input.Version = req.Version
	}
	if req.Status != nil {
		s := string(*req.Status)
		input.Status = &s
	}
	if req.TargetDate != nil {
		d := openapi_types.Date{Time: req.TargetDate.Time}
		input.TargetDate = &d
	}
	if req.Notes != nil {
		input.Notes = req.Notes
	}
	release, err := h.store.UpdateRelease(r.Context(), uuid.UUID(releaseId), projectID, input)
	if handleDBError(w, r, err, "release", "release_update") {
		return
	}
	writeJSON(w, http.StatusOK, mapRelease(release))
}

func (h *API) DeleteRelease(w http.ResponseWriter, r *http.Request, projectId, releaseId openapi_types.UUID) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleContributor) {
		return
	}
	if err := h.store.DeleteRelease(r.Context(), uuid.UUID(releaseId), projectID); err != nil {
		writeError(w, http.StatusInternalServerError, "release_delete_failed", "failed to delete release")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) ExportRelease(w http.ResponseWriter, r *http.Request, projectId, releaseId openapi_types.UUID, params ExportReleaseParams) {
	projectID := uuid.UUID(projectId)
	if !h.requireProjectRole(w, r, projectID, roleViewer) {
		return
	}
	release, err := h.store.GetRelease(r.Context(), uuid.UUID(releaseId), projectID)
	if handleDBError(w, r, err, "release", "release_export") {
		return
	}
	tickets, err := h.store.GetReleaseTickets(r.Context(), uuid.UUID(releaseId))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "release_export_failed", "failed to get release tickets")
		return
	}

	format := "markdown"
	if params.Format != nil {
		format = string(*params.Format)
	}

	var content string
	if format == "json" {
		content = buildJSONNotes(release, tickets)
	} else {
		content = buildMarkdownNotes(release, tickets)
	}

	writeJSON(w, http.StatusOK, releaseExportResponse{
		ReleaseId: openapi_types.UUID(release.ID),
		Name:      release.Name,
		Version:   release.Version,
		Status:    ReleaseStatus(release.Status),
		Content:   content,
	})
}

func buildMarkdownNotes(r store.Release, tickets []store.ReleaseTicket) string {
	var sb strings.Builder
	title := r.Name
	if r.Version != nil && *r.Version != "" {
		title = fmt.Sprintf("%s (%s)", r.Name, *r.Version)
	}
	sb.WriteString(fmt.Sprintf("# Release %s\n\n", title))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", r.Status))
	if r.TargetDate != nil {
		sb.WriteString(fmt.Sprintf("**Target Date:** %s\n", r.TargetDate.Time.Format("2006-01-02")))
	}
	if r.Notes != nil && *r.Notes != "" {
		sb.WriteString(fmt.Sprintf("\n## Notes\n\n%s\n", *r.Notes))
	}

	byType := groupTicketsByType(tickets)
	order := []string{"feature", "bug"}
	for _, extra := range tickets {
		found := false
		for _, t := range order {
			if t == extra.Type {
				found = true
				break
			}
		}
		if !found {
			order = append(order, extra.Type)
		}
	}

	for _, ticketType := range order {
		list, ok := byType[ticketType]
		if !ok {
			continue
		}
		heading := strings.Title(ticketType) + "s" //nolint:staticcheck
		if ticketType == "bug" {
			heading = "Bug Fixes"
		} else if ticketType == "feature" {
			heading = "Features"
		}
		sb.WriteString(fmt.Sprintf("\n## %s\n\n", heading))
		for _, t := range list {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", t.Key, t.Title))
		}
	}
	return sb.String()
}

func buildJSONNotes(r store.Release, tickets []store.ReleaseTicket) string {
	type ticketEntry struct {
		Key       string `json:"key"`
		Title     string `json:"title"`
		Type      string `json:"type"`
		IsClosed  bool   `json:"isClosed"`
		StateName string `json:"stateName"`
	}
	type payload struct {
		Name      string        `json:"name"`
		Version   *string       `json:"version,omitempty"`
		Status    string        `json:"status"`
		CreatedAt time.Time     `json:"createdAt"`
		Tickets   []ticketEntry `json:"tickets"`
	}
	entries := make([]ticketEntry, 0, len(tickets))
	for _, t := range tickets {
		entries = append(entries, ticketEntry{
			Key:       t.Key,
			Title:     t.Title,
			Type:      t.Type,
			IsClosed:  t.IsClosed,
			StateName: t.StateName,
		})
	}
	p := payload{
		Name:      r.Name,
		Version:   r.Version,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		Tickets:   entries,
	}
	b, _ := json.MarshalIndent(p, "", "  ")
	return string(b)
}

func groupTicketsByType(tickets []store.ReleaseTicket) map[string][]store.ReleaseTicket {
	m := map[string][]store.ReleaseTicket{}
	for _, t := range tickets {
		m[t.Type] = append(m[t.Type], t)
	}
	return m
}

func mapRelease(r store.Release) releaseResponse {
	resp := releaseResponse{
		Id:               openapi_types.UUID(r.ID),
		ProjectId:        openapi_types.UUID(r.ProjectID),
		Name:             r.Name,
		Status:           ReleaseStatus(r.Status),
		TicketCount:      r.TicketCount,
		ClosedTicketCount: r.ClosedTicketCount,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
	if r.Version != nil {
		resp.Version = r.Version
	}
	if r.TargetDate != nil {
		d := openapi_types.Date{Time: r.TargetDate.Time}
		resp.TargetDate = &d
	}
	if r.Notes != nil {
		resp.Notes = r.Notes
	}
	return resp
}

func mapReleases(releases []store.Release) []releaseResponse {
	out := make([]releaseResponse, 0, len(releases))
	for _, r := range releases {
		out = append(out, mapRelease(r))
	}
	return out
}
