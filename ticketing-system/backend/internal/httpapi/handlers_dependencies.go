package httpapi

import (
	"context"
	"errors"
	"net/http"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h *API) ListTicketDependencies(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectAccess(w, r, ticket.ProjectID) {
		return
	}

	items, err := h.store.ListTicketDependencies(r.Context(), ticket.ProjectID, ticketID)
	if handleListError(w, r, err, "ticket dependencies", "ticket_dependency_list") {
		return
	}

	writeJSON(w, http.StatusOK, ticketDependencyListResponse{
		Items: mapSlice(items, func(dep store.TicketDependency) TicketDependency {
			return h.mapTicketDependencyWithRelatedTicket(r.Context(), dep)
		}),
	})
}

func (h *API) CreateTicketDependency(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
		return
	}

	req, ok := decodeJSON[ticketDependencyCreateRequest](w, r, "ticket_dependency_create")
	if !ok {
		return
	}
	relatedTicketID := uuid.UUID(req.RelatedTicketId)
	if relatedTicketID == ticketID {
		writeError(w, http.StatusBadRequest, "invalid_dependency", "ticket cannot depend on itself")
		return
	}
	related, err := h.store.GetTicket(r.Context(), relatedTicketID)
	if handleDBError(w, r, err, "related ticket", "ticket_dependency_related_load") {
		return
	}
	if related.ProjectID != ticket.ProjectID {
		writeError(w, http.StatusBadRequest, "invalid_dependency", "related ticket must be in the same project")
		return
	}

	var createdBy *uuid.UUID
	if actorID, ok := currentUserID(w, r); ok {
		createdBy = &actorID
	}
	created, err := h.store.CreateTicketDependency(r.Context(), ticket.ProjectID, store.TicketDependencyCreateInput{
		TicketID:        ticketID,
		RelatedTicketID: relatedTicketID,
		RelationType:    string(req.RelationType),
		CreatedBy:       createdBy,
	})
	if errors.Is(err, store.ErrDependencyCycle) {
		writeError(w, http.StatusConflict, "dependency_cycle", "cyclic dependencies are not allowed")
		return
	}
	if errors.Is(err, store.ErrInvalidDependencyInput) {
		writeError(w, http.StatusBadRequest, "invalid_dependency", "invalid dependency input")
		return
	}
	if err != nil && err.Error() == "dependency already exists" {
		writeError(w, http.StatusConflict, "dependency_exists", "dependency already exists")
		return
	}
	if handleDBErrorWithCode(w, r, err, "ticket dependency", "ticket_dependency_create", "ticket_dependency_create_failed") {
		return
	}

	writeJSON(w, http.StatusCreated, h.mapTicketDependencyWithRelatedTicket(r.Context(), created))
}

func (h *API) DeleteTicketDependency(w http.ResponseWriter, r *http.Request, id openapi_types.UUID, dependencyId openapi_types.UUID) {
	ticketID := uuid.UUID(id)
	ticket, err := h.store.GetTicket(r.Context(), ticketID)
	if handleDBError(w, r, err, "ticket", "ticket_load") {
		return
	}
	if !h.requireProjectRole(w, r, ticket.ProjectID, roleContributor) {
		return
	}
	if err := h.store.DeleteTicketDependency(r.Context(), uuid.UUID(dependencyId), ticket.ProjectID, ticketID); handleDeleteError(w, r, err, "ticket dependency", "ticket_dependency_delete") {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *API) GetProjectDependencyGraph(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params GetProjectDependencyGraphParams) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}

	var root *uuid.UUID
	if params.RootTicketId != nil {
		rootID := uuid.UUID(*params.RootTicketId)
		ticket, err := h.store.GetTicket(r.Context(), rootID)
		if handleDBError(w, r, err, "ticket", "ticket_dependency_graph_root") {
			return
		}
		if ticket.ProjectID != projectUUID {
			writeError(w, http.StatusBadRequest, "invalid_ticket", "rootTicketId must belong to the project")
			return
		}
		root = &rootID
	}

	depth := derefInt(params.Depth, 2)
	graph, err := h.store.GetTicketDependencyGraph(r.Context(), projectUUID, root, depth)
	if handleDBError(w, r, err, "ticket dependency graph", "ticket_dependency_graph") {
		return
	}

	ticketByID := make(map[uuid.UUID]ticketResponse, len(graph.Nodes))
	for _, node := range graph.Nodes {
		if _, ok := ticketByID[node.TicketID]; ok {
			continue
		}
		ticket, err := h.store.GetTicket(r.Context(), node.TicketID)
		if err != nil {
			continue
		}
		ticketByID[node.TicketID] = mapTicket(ticket)
	}

	nodes := make([]TicketDependencyGraphNode, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		mapped, ok := ticketByID[node.TicketID]
		if !ok {
			continue
		}
		nodes = append(nodes, TicketDependencyGraphNode{
			Ticket: mapped,
			Depth:  node.Depth,
		})
	}

	edges := make([]TicketDependencyGraphEdge, 0, len(graph.Edges))
	for _, edge := range graph.Edges {
		edges = append(edges, TicketDependencyGraphEdge{
			Id:             toOpenapiUUID(edge.ID),
			SourceTicketId: toOpenapiUUID(edge.FromTicketID),
			TargetTicketId: toOpenapiUUID(edge.ToTicketID),
			RelationType:   DependencyRelationType(edge.RelationType),
		})
	}

	writeJSON(w, http.StatusOK, TicketDependencyGraphResponse{
		Nodes: nodes,
		Edges: edges,
	})
}

func (h *API) mapTicketDependencyWithRelatedTicket(ctx context.Context, dep store.TicketDependency) TicketDependency {
	out := TicketDependency{
		Id:              toOpenapiUUID(dep.ID),
		ProjectId:       toOpenapiUUID(dep.ProjectID),
		TicketId:        toOpenapiUUID(dep.TicketID),
		RelatedTicketId: toOpenapiUUID(dep.RelatedTicketID),
		RelationType:    DependencyRelationType(dep.RelationType),
		CreatedAt:       dep.CreatedAt,
	}
	related, err := h.store.GetTicket(ctx, dep.RelatedTicketID)
	if err == nil {
		mapped := mapTicket(related)
		out.RelatedTicket = &mapped
	}
	return out
}
