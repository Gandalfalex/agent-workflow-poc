package httpapi

import (
	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func toOpenapiUUID(id uuid.UUID) openapi_types.UUID {
	return openapi_types.UUID(id)
}

func mapWorkflowStates(states []store.WorkflowState, projectID openapi_types.UUID) []workflowState {
	out := make([]workflowState, 0, len(states))
	for _, state := range states {
		out = append(out, workflowState{
			Id:        toOpenapiUUID(state.ID),
			ProjectId: projectID,
			Name:      state.Name,
			Order:     state.Order,
			IsDefault: state.IsDefault,
			IsClosed:  state.IsClosed,
		})
	}
	return out
}

func mapTicket(ticket store.Ticket) ticketResponse {
	projectID := toOpenapiUUID(ticket.ProjectID)
	projectKey := ProjectKey(ticket.ProjectKey)
	var assigneeID *openapi_types.UUID
	var assignee *userSummary
	var description *string

	if ticket.Description != "" {
		description = &ticket.Description
	}

	if ticket.AssigneeID != nil {
		value := toOpenapiUUID(*ticket.AssigneeID)
		assigneeID = &value
	}
	if ticket.AssigneeID != nil && ticket.AssigneeName != nil {
		assignee = &userSummary{Id: toOpenapiUUID(*ticket.AssigneeID), Name: *ticket.AssigneeName}
	}

	state := workflowState{
		Id:        toOpenapiUUID(ticket.StateID),
		ProjectId: projectID,
		Name:      ticket.StateName,
		Order:     ticket.StateOrder,
		IsDefault: ticket.StateDefault,
		IsClosed:  ticket.StateClosed,
	}

	storyOapiID := toOpenapiUUID(ticket.StoryID)
	story := &Story{
		Id:          storyOapiID,
		ProjectId:   projectID,
		Title:       ticket.StoryTitle,
		Description: ticket.StorySummary,
		CreatedAt:   ticket.StoryCreated,
		UpdatedAt:   ticket.StoryUpdated,
	}

	return ticketResponse{
		Id:          toOpenapiUUID(ticket.ID),
		Key:         TicketKey(ticket.Key),
		Number:      ticket.Number,
		Type:        TicketType(ticket.Type),
		ProjectId:   projectID,
		ProjectKey:  projectKey,
		StoryId:     storyOapiID,
		Story:       story,
		Title:       ticket.Title,
		Description: description,
		StateId:     toOpenapiUUID(ticket.StateID),
		State:       &state,
		AssigneeId:  assigneeID,
		Assignee:    assignee,
		Priority:    TicketPriority(ticket.Priority),
		Position:    float32(ticket.Position),
		CreatedAt:   ticket.CreatedAt,
		UpdatedAt:   ticket.UpdatedAt,
	}
}

func mapWebhook(hook store.Webhook, projectID openapi_types.UUID) webhookResponse {
	return webhookResponse{
		Id:        toOpenapiUUID(hook.ID),
		ProjectId: projectID,
		Url:       hook.URL,
		Events:    mapWebhookEvents(hook.Events),
		Enabled:   hook.Enabled,
		CreatedAt: hook.CreatedAt,
		UpdatedAt: hook.UpdatedAt,
	}
}

func mapWebhookEvents(events []string) []WebhookEvent {
	out := make([]WebhookEvent, 0, len(events))
	for _, event := range events {
		out = append(out, WebhookEvent(event))
	}
	return out
}

func mapProject(project store.Project) projectResponse {
	return projectResponse{
		Id:          toOpenapiUUID(project.ID),
		Key:         ProjectKey(project.Key),
		Name:        project.Name,
		Description: project.Description,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func mapGroup(group store.Group) groupResponse {
	return groupResponse{
		Id:          toOpenapiUUID(group.ID),
		Name:        group.Name,
		Description: group.Description,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}

func mapGroupMember(member store.GroupMember) groupMemberResponse {
	var user *userSummary
	if member.UserName != nil {
		user = &userSummary{
			Id:   toOpenapiUUID(member.UserID),
			Name: *member.UserName,
		}
	}
	return groupMemberResponse{
		GroupId: toOpenapiUUID(member.GroupID),
		UserId:  toOpenapiUUID(member.UserID),
		User:    user,
	}
}

func mapProjectGroup(group store.ProjectGroup) projectGroupResponse {
	role := ProjectRole(group.Role)
	return projectGroupResponse{
		ProjectId:   toOpenapiUUID(group.ProjectID),
		GroupId:     toOpenapiUUID(group.GroupID),
		Role:        role,
		Permissions: permissionsForRole(role),
	}
}

func permissionsForRole(role ProjectRole) []ProjectPermission {
	switch role {
	case ProjectRoleAdmin:
		return []ProjectPermission{
			ProjectPermissionAdmin,
			ProjectPermissionRead,
			ProjectPermissionCreate,
			ProjectPermissionUpdate,
			ProjectPermissionDelete,
		}
	case ProjectRoleContributor:
		return []ProjectPermission{
			ProjectPermissionRead,
			ProjectPermissionCreate,
			ProjectPermissionUpdate,
			ProjectPermissionDelete,
		}
	default:
		return []ProjectPermission{
			ProjectPermissionRead,
		}
	}
}

func mapStory(story store.Story) storyResponse {
	return storyResponse{
		Id:          toOpenapiUUID(story.ID),
		ProjectId:   toOpenapiUUID(story.ProjectID),
		Title:       story.Title,
		Description: story.Description,
		CreatedAt:   story.CreatedAt,
		UpdatedAt:   story.UpdatedAt,
	}
}

func mapComment(comment store.Comment) ticketCommentResponse {
	return ticketCommentResponse{
		Id:         toOpenapiUUID(comment.ID),
		TicketId:   toOpenapiUUID(comment.TicketID),
		AuthorId:   toOpenapiUUID(comment.AuthorID),
		AuthorName: comment.AuthorName,
		Message:    comment.Message,
		CreatedAt:  comment.CreatedAt,
	}
}
