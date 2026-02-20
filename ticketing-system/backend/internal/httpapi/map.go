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
	var incidentCommanderID *openapi_types.UUID
	var incidentCommander *userSummary
	var incidentSeverity *TicketIncidentSeverity
	var incidentImpact *string
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
	if ticket.IncidentCommanderID != nil {
		value := toOpenapiUUID(*ticket.IncidentCommanderID)
		incidentCommanderID = &value
	}
	if ticket.IncidentCommanderID != nil && ticket.IncidentCommanderName != nil {
		incidentCommander = &userSummary{Id: toOpenapiUUID(*ticket.IncidentCommanderID), Name: *ticket.IncidentCommanderName}
	}
	if ticket.IncidentSeverity != nil {
		sev := TicketIncidentSeverity(*ticket.IncidentSeverity)
		incidentSeverity = &sev
	}
	if ticket.IncidentImpact != nil && *ticket.IncidentImpact != "" {
		incidentImpact = ticket.IncidentImpact
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
		Id:                  toOpenapiUUID(ticket.ID),
		Key:                 TicketKey(ticket.Key),
		Number:              ticket.Number,
		Type:                TicketType(ticket.Type),
		ProjectId:           projectID,
		ProjectKey:          projectKey,
		StoryId:             storyOapiID,
		Story:               story,
		Title:               ticket.Title,
		Description:         description,
		StateId:             toOpenapiUUID(ticket.StateID),
		State:               &state,
		AssigneeId:          assigneeID,
		Assignee:            assignee,
		Priority:            TicketPriority(ticket.Priority),
		IncidentEnabled:     ticket.IncidentEnabled,
		IncidentSeverity:    incidentSeverity,
		IncidentImpact:      incidentImpact,
		IncidentCommanderId: incidentCommanderID,
		IncidentCommander:   incidentCommander,
		Position:            float32(ticket.Position),
		BlockedByCount:      ticket.BlockedByCount,
		IsBlocked:           ticket.IsBlocked,
		CreatedAt:           ticket.CreatedAt,
		UpdatedAt:           ticket.UpdatedAt,
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
	case ProjectRole("admin"):
		return []ProjectPermission{
			ProjectPermissionAdmin,
			ProjectPermissionRead,
			ProjectPermissionCreate,
			ProjectPermissionUpdate,
			ProjectPermissionDelete,
		}
	case ProjectRole("contributor"):
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

func mapProjectActivity(a store.ProjectActivity) projectActivityResponse {
	resp := projectActivityResponse{
		Id:          toOpenapiUUID(a.ID),
		TicketId:    toOpenapiUUID(a.TicketID),
		TicketKey:   a.TicketKey,
		TicketTitle: a.TicketTitle,
		ActorId:     toOpenapiUUID(a.ActorID),
		ActorName:   a.ActorName,
		Action:      a.Action,
		CreatedAt:   a.CreatedAt,
	}
	if a.Field != nil {
		resp.Field = a.Field
	}
	if a.OldValue != nil {
		resp.OldValue = a.OldValue
	}
	if a.NewValue != nil {
		resp.NewValue = a.NewValue
	}
	return resp
}

func mapActivity(a store.Activity) ticketActivityResponse {
	resp := ticketActivityResponse{
		Id:        toOpenapiUUID(a.ID),
		TicketId:  toOpenapiUUID(a.TicketID),
		ActorId:   toOpenapiUUID(a.ActorID),
		ActorName: a.ActorName,
		Action:    a.Action,
		CreatedAt: a.CreatedAt,
	}
	if a.Field != nil {
		resp.Field = a.Field
	}
	if a.OldValue != nil {
		resp.OldValue = a.OldValue
	}
	if a.NewValue != nil {
		resp.NewValue = a.NewValue
	}
	return resp
}

func mapWebhookDelivery(d store.WebhookDelivery) webhookDeliveryResponse {
	resp := webhookDeliveryResponse{
		Id:         toOpenapiUUID(d.ID),
		WebhookId:  toOpenapiUUID(d.WebhookID),
		Event:      d.Event,
		Attempt:    d.Attempt,
		Delivered:  d.Delivered,
		DurationMs: d.DurationMs,
		CreatedAt:  d.CreatedAt,
	}
	if d.StatusCode != nil {
		resp.StatusCode = d.StatusCode
	}
	if d.ResponseBody != nil {
		resp.ResponseBody = d.ResponseBody
	}
	if d.Error != nil {
		resp.Error = d.Error
	}
	return resp
}

func mapProjectStats(stats store.ProjectStats) projectStatsResponse {
	return projectStatsResponse{
		TotalOpen:   stats.TotalOpen,
		TotalClosed: stats.TotalClosed,
		BlockedOpen: stats.BlockedOpen,
		ByState:     mapStatCounts(stats.ByState),
		ByPriority:  mapStatCounts(stats.ByPriority),
		ByType:      mapStatCounts(stats.ByType),
		ByAssignee:  mapStatCounts(stats.ByAssignee),
	}
}

func mapSprint(item store.Sprint) sprintResponse {
	out := sprintResponse{
		Id:               toOpenapiUUID(item.ID),
		ProjectId:        toOpenapiUUID(item.ProjectID),
		Name:             item.Name,
		StartDate:        openapi_types.Date{Time: item.StartDate},
		EndDate:          openapi_types.Date{Time: item.EndDate},
		TicketIds:        make([]openapi_types.UUID, 0, len(item.TicketIDs)),
		CommittedTickets: item.CommittedTickets,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
	for _, id := range item.TicketIDs {
		out.TicketIds = append(out.TicketIds, toOpenapiUUID(id))
	}
	if item.Goal != nil {
		out.Goal = item.Goal
	}
	return out
}

func mapCapacitySetting(item store.CapacitySetting) capacitySettingResponse {
	out := capacitySettingResponse{
		Id:        toOpenapiUUID(item.ID),
		ProjectId: toOpenapiUUID(item.ProjectID),
		Scope:     CapacitySettingScope(item.Scope),
		Label:     item.Label,
		Capacity:  item.Capacity,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
	if item.UserID != nil {
		value := toOpenapiUUID(*item.UserID)
		out.UserId = &value
	}
	return out
}

func mapSprintForecastSummary(item store.SprintForecastSummary) sprintForecastSummaryResponse {
	out := sprintForecastSummaryResponse{
		CommittedTickets:    item.CommittedTickets,
		Capacity:            item.Capacity,
		ProjectedCompletion: item.ProjectedCompletion,
		OverCapacityDelta:   item.OverCapacityDelta,
		Confidence:          float32(item.Confidence),
		Iterations:          item.Iterations,
	}
	if item.Sprint != nil {
		mapped := mapSprint(*item.Sprint)
		out.Sprint = &mapped
	}
	return out
}

func mapAiTriageSettings(item store.AiTriageSettings) aiTriageSettingsResponse {
	return aiTriageSettingsResponse{
		Enabled: item.Enabled,
	}
}

func mapAiTriageSuggestion(item store.AiTriageSuggestion) aiTriageSuggestionResponse {
	out := aiTriageSuggestionResponse{
		Id:            toOpenapiUUID(item.ID),
		ProjectId:     toOpenapiUUID(item.ProjectID),
		Summary:       item.Summary,
		Priority:      TicketPriority(item.Priority),
		StateId:       toOpenapiUUID(item.StateID),
		PromptVersion: item.PromptVersion,
		Model:         item.Model,
		Confidence: AiTriageConfidence{
			Summary:  item.ConfidenceSummary,
			Priority: item.ConfidencePriority,
			State:    item.ConfidenceState,
			Assignee: item.ConfidenceAssignee,
		},
		CreatedAt: item.CreatedAt,
	}
	if item.AssigneeID != nil {
		value := toOpenapiUUID(*item.AssigneeID)
		out.AssigneeId = &value
	}
	return out
}

func mapAiTriageSuggestionDecision(item store.AiTriageSuggestionDecision) aiTriageSuggestionDecisionResponse {
	out := aiTriageSuggestionDecisionResponse{
		Id:             toOpenapiUUID(item.ID),
		SuggestionId:   toOpenapiUUID(item.SuggestionID),
		ProjectId:      toOpenapiUUID(item.ProjectID),
		ActorId:        toOpenapiUUID(item.ActorID),
		AcceptedFields: make([]AiTriageField, 0, len(item.AcceptedFields)),
		RejectedFields: make([]AiTriageField, 0, len(item.RejectedFields)),
		CreatedAt:      item.CreatedAt,
	}
	for _, field := range item.AcceptedFields {
		out.AcceptedFields = append(out.AcceptedFields, AiTriageField(field))
	}
	for _, field := range item.RejectedFields {
		out.RejectedFields = append(out.RejectedFields, AiTriageField(field))
	}
	return out
}

func mapProjectReportingSummary(report store.ProjectReportingSummary) ProjectReportingSummary {
	throughput := make([]DateValuePoint, 0, len(report.ThroughputByDay))
	for _, point := range report.ThroughputByDay {
		throughput = append(throughput, DateValuePoint{
			Date:  openapi_types.Date{Time: point.Date},
			Value: point.Value,
		})
	}

	openByState := make([]StateOpenPoint, 0, len(report.OpenByState))
	for _, point := range report.OpenByState {
		openByState = append(openByState, StateOpenPoint{
			Date:   openapi_types.Date{Time: point.Date},
			Counts: mapStatCounts(point.Counts),
		})
	}

	return ProjectReportingSummary{
		From:                  openapi_types.Date{Time: report.From},
		To:                    openapi_types.Date{Time: report.To},
		ThroughputByDay:       throughput,
		AverageCycleTimeHours: float32(report.AverageCycleTimeHours),
		OpenByState:           openByState,
	}
}

func mapStatCounts(counts []store.StatCount) []statCountResponse {
	out := make([]statCountResponse, 0, len(counts))
	for _, sc := range counts {
		out = append(out, statCountResponse{Label: sc.Label, Value: sc.Value})
	}
	return out
}

func mapAttachment(att store.Attachment) attachmentResponse {
	return attachmentResponse{
		Id:             toOpenapiUUID(att.ID),
		TicketId:       toOpenapiUUID(att.TicketID),
		Filename:       att.Filename,
		ContentType:    att.ContentType,
		Size:           att.Size,
		UploadedBy:     toOpenapiUUID(att.UploadedBy),
		UploadedByName: att.UploadedByName,
		CreatedAt:      att.CreatedAt,
	}
}

func mapBoardFilter(filter store.BoardFilter) boardFilter {
	out := boardFilter{}
	if filter.AssigneeID != nil {
		id := toOpenapiUUID(*filter.AssigneeID)
		out.AssigneeId = &id
	}
	if filter.StateID != nil {
		id := toOpenapiUUID(*filter.StateID)
		out.StateId = &id
	}
	if filter.Priority != nil {
		priority := TicketPriority(*filter.Priority)
		out.Priority = &priority
	}
	if filter.Type != nil {
		ticketType := TicketType(*filter.Type)
		out.Type = &ticketType
	}
	if filter.Query != nil {
		out.Q = filter.Query
	}
	if filter.Blocked != nil {
		out.Blocked = filter.Blocked
	}
	return out
}

func mapStoreBoardFilter(filter boardFilter) store.BoardFilter {
	out := store.BoardFilter{}
	if filter.AssigneeId != nil {
		id := uuid.UUID(*filter.AssigneeId)
		out.AssigneeID = &id
	}
	if filter.StateId != nil {
		id := uuid.UUID(*filter.StateId)
		out.StateID = &id
	}
	if filter.Priority != nil {
		value := string(*filter.Priority)
		out.Priority = &value
	}
	if filter.Type != nil {
		value := string(*filter.Type)
		out.Type = &value
	}
	if filter.Q != nil {
		out.Query = filter.Q
	}
	if filter.Blocked != nil {
		out.Blocked = filter.Blocked
	}
	return out
}

func mapBoardFilterPreset(preset store.BoardFilterPreset) boardFilterPresetResponse {
	return boardFilterPresetResponse{
		Id:         toOpenapiUUID(preset.ID),
		ProjectId:  toOpenapiUUID(preset.ProjectID),
		OwnerId:    toOpenapiUUID(preset.OwnerID),
		Name:       preset.Name,
		Filters:    mapBoardFilter(preset.Filters),
		ShareToken: preset.ShareToken,
		CreatedAt:  preset.CreatedAt,
		UpdatedAt:  preset.UpdatedAt,
	}
}

func mapNotification(n store.Notification) notificationResponse {
	typ := NotificationType(n.Type)
	return notificationResponse{
		Id:          toOpenapiUUID(n.ID),
		ProjectId:   toOpenapiUUID(n.ProjectID),
		UserId:      toOpenapiUUID(n.UserID),
		TicketId:    toOpenapiUUID(n.TicketID),
		TicketKey:   TicketKey(n.TicketKey),
		TicketTitle: n.TicketTitle,
		Type:        typ,
		Message:     n.Message,
		ReadAt:      n.ReadAt,
		CreatedAt:   n.CreatedAt,
	}
}

func mapNotificationPreferences(p store.NotificationPreferences) notificationPreferencesResponse {
	return notificationPreferencesResponse{
		MentionEnabled:    p.MentionEnabled,
		AssignmentEnabled: p.AssignmentEnabled,
	}
}
