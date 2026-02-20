package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	DependencyRelationBlocks    = "blocks"
	DependencyRelationBlockedBy = "blocked_by"
	DependencyRelationRelated   = "related"
)

var (
	ErrDependencyCycle        = errors.New("dependency cycle detected")
	ErrInvalidDependencyInput = errors.New("invalid dependency input")
)

type TicketDependency struct {
	ID              uuid.UUID
	ProjectID       uuid.UUID
	TicketID        uuid.UUID
	RelatedTicketID uuid.UUID
	RelationType    string
	CreatedAt       time.Time
}

type TicketDependencyCreateInput struct {
	TicketID        uuid.UUID
	RelatedTicketID uuid.UUID
	RelationType    string
	CreatedBy       *uuid.UUID
}

type TicketDependencyGraphNode struct {
	TicketID uuid.UUID
	Depth    int
}

type TicketDependencyEdge struct {
	ID           uuid.UUID
	ProjectID    uuid.UUID
	FromTicketID uuid.UUID
	ToTicketID   uuid.UUID
	RelationType string
	CreatedAt    time.Time
}

type TicketDependencyGraph struct {
	Nodes []TicketDependencyGraphNode
	Edges []TicketDependencyEdge
}

func (s *Store) CreateTicketDependency(ctx context.Context, projectID uuid.UUID, input TicketDependencyCreateInput) (TicketDependency, error) {
	source, target, storedType, err := normalizeDependencyCreate(input.TicketID, input.RelatedTicketID, input.RelationType)
	if err != nil {
		return TicketDependency{}, err
	}

	if storedType == DependencyRelationBlocks {
		cycle, err := s.blocksPathExists(ctx, projectID, target, source)
		if err != nil {
			return TicketDependency{}, err
		}
		if cycle {
			return TicketDependency{}, ErrDependencyCycle
		}
	}

	query := mustSQL("ticket_dependencies_insert", nil)
	var id uuid.UUID
	if err := s.db.QueryRow(ctx, query, projectID, source, target, storedType, input.CreatedBy).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return TicketDependency{}, errors.New("dependency already exists")
		}
		return TicketDependency{}, err
	}

	return s.GetTicketDependencyForTicket(ctx, id, projectID, input.TicketID)
}

func (s *Store) ListTicketDependencies(ctx context.Context, projectID, ticketID uuid.UUID) ([]TicketDependency, error) {
	query := mustSQL("ticket_dependencies_list_for_ticket", nil)
	return queryMany(ctx, s.db, query, scanTicketDependency, projectID, ticketID)
}

func (s *Store) GetTicketDependencyForTicket(ctx context.Context, dependencyID, projectID, ticketID uuid.UUID) (TicketDependency, error) {
	query := mustSQL("ticket_dependencies_get_for_ticket", nil)
	return queryOne(ctx, s.db, query, scanTicketDependency, dependencyID, projectID, ticketID)
}

func (s *Store) DeleteTicketDependency(ctx context.Context, dependencyID, projectID, ticketID uuid.UUID) error {
	query := mustSQL("ticket_dependencies_delete_for_ticket", nil)
	return execOne(ctx, s.db, query, pgx.ErrNoRows, dependencyID, projectID, ticketID)
}

func (s *Store) GetTicketDependencyGraph(ctx context.Context, projectID uuid.UUID, rootTicketID *uuid.UUID, depth int) (TicketDependencyGraph, error) {
	if depth < 1 {
		depth = 1
	}
	if depth > 2 {
		depth = 2
	}

	graph := TicketDependencyGraph{}
	if rootTicketID == nil {
		edges, err := queryMany(ctx, s.db, mustSQL("ticket_dependencies_edges_for_project", nil), scanTicketDependencyEdge, projectID)
		if err != nil {
			return graph, err
		}
		graph.Edges = edges
		seen := make(map[uuid.UUID]struct{})
		for _, edge := range edges {
			seen[edge.FromTicketID] = struct{}{}
			seen[edge.ToTicketID] = struct{}{}
		}
		nodes := make([]TicketDependencyGraphNode, 0, len(seen))
		for id := range seen {
			nodes = append(nodes, TicketDependencyGraphNode{TicketID: id, Depth: 0})
		}
		graph.Nodes = nodes
		return graph, nil
	}

	type rawNode struct {
		TicketID uuid.UUID
		Depth    int
	}
	scanRawNode := func(row pgx.Row) (rawNode, error) {
		var node rawNode
		err := row.Scan(&node.TicketID, &node.Depth)
		return node, err
	}

	rawNodes, err := queryMany(ctx, s.db, mustSQL("ticket_dependencies_nodes_from_root", nil), scanRawNode, projectID, *rootTicketID, depth)
	if err != nil {
		return graph, err
	}
	if len(rawNodes) == 0 {
		graph.Nodes = []TicketDependencyGraphNode{{TicketID: *rootTicketID, Depth: 0}}
		graph.Edges = []TicketDependencyEdge{}
		return graph, nil
	}

	nodeIDs := make([]uuid.UUID, 0, len(rawNodes))
	nodes := make([]TicketDependencyGraphNode, 0, len(rawNodes))
	for _, node := range rawNodes {
		nodeIDs = append(nodeIDs, node.TicketID)
		nodes = append(nodes, TicketDependencyGraphNode{TicketID: node.TicketID, Depth: node.Depth})
	}
	edges, err := queryMany(ctx, s.db, mustSQL("ticket_dependencies_edges_for_nodes", nil), scanTicketDependencyEdge, projectID, nodeIDs)
	if err != nil {
		return graph, err
	}
	graph.Nodes = nodes
	graph.Edges = edges
	return graph, nil
}

func (s *Store) blocksPathExists(ctx context.Context, projectID, start, target uuid.UUID) (bool, error) {
	query := mustSQL("ticket_dependencies_blocks_path_exists", nil)
	var exists bool
	if err := s.db.QueryRow(ctx, query, projectID, start, target).Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

func normalizeDependencyCreate(source, target uuid.UUID, relation string) (uuid.UUID, uuid.UUID, string, error) {
	if source == uuid.Nil || target == uuid.Nil || source == target {
		return uuid.Nil, uuid.Nil, "", ErrInvalidDependencyInput
	}
	normalized := strings.ToLower(strings.TrimSpace(relation))
	switch normalized {
	case DependencyRelationBlocks:
		return source, target, DependencyRelationBlocks, nil
	case DependencyRelationBlockedBy:
		return target, source, DependencyRelationBlocks, nil
	case DependencyRelationRelated:
		if strings.Compare(source.String(), target.String()) > 0 {
			return target, source, DependencyRelationRelated, nil
		}
		return source, target, DependencyRelationRelated, nil
	default:
		return uuid.Nil, uuid.Nil, "", ErrInvalidDependencyInput
	}
}

func scanTicketDependency(row pgx.Row) (TicketDependency, error) {
	var dep TicketDependency
	err := row.Scan(
		&dep.ID,
		&dep.ProjectID,
		&dep.TicketID,
		&dep.RelatedTicketID,
		&dep.RelationType,
		&dep.CreatedAt,
	)
	return dep, err
}

func scanTicketDependencyEdge(row pgx.Row) (TicketDependencyEdge, error) {
	var edge TicketDependencyEdge
	err := row.Scan(
		&edge.ID,
		&edge.ProjectID,
		&edge.FromTicketID,
		&edge.ToTicketID,
		&edge.RelationType,
		&edge.CreatedAt,
	)
	return edge, err
}
