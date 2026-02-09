package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Group struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type GroupCreateInput struct {
	Name        string
	Description *string
}

type GroupUpdateInput struct {
	Name        *string
	Description *string
}

type GroupMember struct {
	GroupID  uuid.UUID
	UserID   uuid.UUID
	UserName *string
}

type ProjectGroup struct {
	ProjectID uuid.UUID
	GroupID   uuid.UUID
	Role      string
}

func (s *Store) ListGroups(ctx context.Context) ([]Group, error) {
	query := mustSQL("groups_list.sql", nil)
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	groups := []Group{}
	for rows.Next() {
		group, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		groups = append(groups, group)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (s *Store) GetGroup(ctx context.Context, id uuid.UUID) (Group, error) {
	query := mustSQL("groups_get.sql", nil)
	row := s.db.QueryRow(ctx, query, id)
	return scanGroup(row)
}

func (s *Store) CreateGroup(ctx context.Context, input GroupCreateInput) (Group, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return Group{}, errors.New("name required")
	}

	query := mustSQL("groups_insert.sql", nil)
	var id uuid.UUID
	if err := s.db.QueryRow(ctx, query, name, input.Description).Scan(&id); err != nil {
		return Group{}, err
	}
	return s.GetGroup(ctx, id)
}

func (s *Store) UpdateGroup(ctx context.Context, id uuid.UUID, input GroupUpdateInput) (Group, error) {
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return Group{}, errors.New("name required")
		}
		input.Name = &name
	}

	query := mustSQL("groups_update.sql", nil)
	var updatedID uuid.UUID
	if err := s.db.QueryRow(ctx, query, id, input.Name, input.Description).Scan(&updatedID); err != nil {
		return Group{}, err
	}
	return s.GetGroup(ctx, updatedID)
}

func (s *Store) DeleteGroup(ctx context.Context, id uuid.UUID) error {
	query := mustSQL("groups_delete.sql", nil)
	tag, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]GroupMember, error) {
	query := mustSQL("group_members_list.sql", nil)
	rows, err := s.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []GroupMember{}
	for rows.Next() {
		var member GroupMember
		if err := rows.Scan(&member.GroupID, &member.UserID, &member.UserName); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (s *Store) AddGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) (GroupMember, error) {
	// Validate that the user exists before attempting to add them
	checkQuery := mustSQL("users_exists.sql", nil)
	var exists bool
	if err := s.db.QueryRow(ctx, checkQuery, userID).Scan(&exists); err != nil {
		return GroupMember{}, err
	}
	if !exists {
		return GroupMember{}, errors.New("user not found")
	}

	query := mustSQL("group_members_insert.sql", nil)
	var member GroupMember
	if err := s.db.QueryRow(ctx, query, groupID, userID).Scan(&member.GroupID, &member.UserID, &member.UserName); err != nil {
		return GroupMember{}, err
	}
	return member, nil
}

func (s *Store) DeleteGroupMember(ctx context.Context, groupID uuid.UUID, userID uuid.UUID) error {
	query := mustSQL("group_members_delete.sql", nil)
	tag, err := s.db.Exec(ctx, query, groupID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (s *Store) ListProjectGroups(ctx context.Context, projectID uuid.UUID) ([]ProjectGroup, error) {
	query := mustSQL("project_groups_list.sql", nil)
	rows, err := s.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []ProjectGroup{}
	for rows.Next() {
		var item ProjectGroup
		if err := rows.Scan(&item.ProjectID, &item.GroupID, &item.Role); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (s *Store) AddProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID, role string) (ProjectGroup, error) {
	role, err := normalizeProjectRole(role)
	if err != nil {
		return ProjectGroup{}, err
	}
	query := mustSQL("project_groups_insert.sql", nil)
	var item ProjectGroup
	if err := s.db.QueryRow(ctx, query, projectID, groupID, role).Scan(&item.ProjectID, &item.GroupID, &item.Role); err != nil {
		return ProjectGroup{}, err
	}
	return item, nil
}

func (s *Store) UpdateProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID, role string) (ProjectGroup, error) {
	role, err := normalizeProjectRole(role)
	if err != nil {
		return ProjectGroup{}, err
	}
	query := mustSQL("project_groups_update.sql", nil)
	var item ProjectGroup
	if err := s.db.QueryRow(ctx, query, projectID, groupID, role).Scan(&item.ProjectID, &item.GroupID, &item.Role); err != nil {
		return ProjectGroup{}, err
	}
	return item, nil
}

func (s *Store) DeleteProjectGroup(ctx context.Context, projectID uuid.UUID, groupID uuid.UUID) error {
	query := mustSQL("project_groups_delete.sql", nil)
	tag, err := s.db.Exec(ctx, query, projectID, groupID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func normalizeProjectRole(value string) (string, error) {
	role := strings.ToLower(strings.TrimSpace(value))
	switch role {
	case "admin", "contributor", "viewer":
		return role, nil
	default:
		return "", errors.New("invalid project role")
	}
}

func scanGroup(row pgx.Row) (Group, error) {
	var group Group
	err := row.Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	return group, err
}
