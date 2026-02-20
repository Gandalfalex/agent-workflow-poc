package httpapi

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"ticketing-system/backend/internal/store"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var mentionPattern = regexp.MustCompile(`@([A-Za-z0-9._-]{2,64})`)

func (h *API) ListNotifications(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, params ListNotificationsParams) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	userID, _, ok := currentActor(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	unreadOnly := false
	if params.UnreadOnly != nil {
		unreadOnly = *params.UnreadOnly
	}
	items, err := h.store.ListNotifications(r.Context(), store.NotificationFilter{
		ProjectID:  projectUUID,
		UserID:     userID,
		UnreadOnly: unreadOnly,
		Limit:      limit,
	})
	if handleListError(w, r, err, "notifications", "notification_list") {
		return
	}
	writeJSON(w, http.StatusOK, notificationListResponse{Items: mapSlice(items, mapNotification)})
}

func (h *API) GetNotificationUnreadCount(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	userID, _, ok := currentActor(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	count, err := h.store.CountUnreadNotifications(r.Context(), projectUUID, userID)
	if handleListError(w, r, err, "notifications", "notification_count") {
		return
	}
	writeJSON(w, http.StatusOK, notificationUnreadCountResponse{Count: count})
}

func (h *API) MarkNotificationRead(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID, notificationId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	userID, _, ok := currentActor(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	item, err := h.store.MarkNotificationRead(r.Context(), uuid.UUID(notificationId), projectUUID, userID)
	if handleDBError(w, r, err, "notification", "notification_mark_read") {
		return
	}
	h.publishUserNotificationEvents(r.Context(), projectUUID, userID)
	writeJSON(w, http.StatusOK, mapNotification(item))
}

func (h *API) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	userID, _, ok := currentActor(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	updated, err := h.store.MarkAllNotificationsRead(r.Context(), projectUUID, userID)
	if handleDBError(w, r, err, "notifications", "notification_mark_all_read") {
		return
	}
	h.publishUserNotificationEvents(r.Context(), projectUUID, userID)
	writeJSON(w, http.StatusOK, notificationMarkAllResponse{Updated: updated})
}

func (h *API) GetNotificationPreferences(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	userID, _, ok := currentActor(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	prefs, err := h.store.GetNotificationPreferences(r.Context(), userID)
	if handleDBError(w, r, err, "notification preferences", "notification_prefs_get") {
		return
	}
	writeJSON(w, http.StatusOK, mapNotificationPreferences(prefs))
}

func (h *API) UpdateNotificationPreferences(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	userID, _, ok := currentActor(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	req, ok := decodeJSON[notificationPreferencesUpdateRequest](w, r, "notification_prefs_update")
	if !ok {
		return
	}
	prefs, err := h.store.UpdateNotificationPreferences(r.Context(), userID, store.NotificationPreferencesUpdateInput{
		MentionEnabled:    req.MentionEnabled,
		AssignmentEnabled: req.AssignmentEnabled,
	})
	if handleDBErrorWithCode(w, r, err, "notification preferences", "notification_prefs_update", "notification_prefs_update_failed") {
		return
	}
	writeJSON(w, http.StatusOK, mapNotificationPreferences(prefs))
}

func currentActor(r *http.Request) (uuid.UUID, string, bool) {
	user, ok := authUser(r.Context())
	if !ok {
		return uuid.Nil, "", false
	}
	actorID, err := uuid.Parse(user.ID)
	if err != nil {
		return uuid.Nil, "", false
	}
	return actorID, user.Name, true
}

func extractMentions(message string) []string {
	matches := mentionPattern.FindAllStringSubmatch(message, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(m[1]))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func (h *API) notifyMentions(r *http.Request, projectID uuid.UUID, ticket store.Ticket, actorID uuid.UUID, actorName, message string) {
	mentions := extractMentions(message)
	if len(mentions) == 0 {
		return
	}
	users, err := h.store.ListUsers(r.Context(), "")
	if err != nil {
		logRequestError(r, "notification_mentions_users_list_failed", err)
		return
	}
	userByName := make(map[string]store.UserSummary, len(users))
	for _, user := range users {
		userByName[strings.ToLower(user.Name)] = user
		if parts := strings.Split(strings.ToLower(user.Email), "@"); len(parts) > 0 && parts[0] != "" {
			userByName[parts[0]] = user
		}
	}
	for _, mention := range mentions {
		user, ok := userByName[mention]
		if !ok || user.ID == actorID {
			continue
		}
		role, err := h.store.GetProjectRoleForUser(r.Context(), projectID, user.ID)
		if err != nil || role == "" {
			continue
		}
		prefs, err := h.store.GetNotificationPreferences(r.Context(), user.ID)
		if err != nil {
			continue
		}
		if !prefs.MentionEnabled {
			continue
		}
		_, err = h.store.CreateNotification(r.Context(), store.NotificationCreateInput{
			ProjectID: projectID,
			UserID:    user.ID,
			TicketID:  ticket.ID,
			Type:      "mention",
			Message:   fmt.Sprintf("%s mentioned you on %s", actorName, ticket.Key),
		})
		if err != nil {
			logRequestError(r, "notification_mention_create_failed", err)
			continue
		}
		h.publishUserNotificationEvents(r.Context(), projectID, user.ID)
	}
}

func (h *API) notifyAssignment(r *http.Request, before, after store.Ticket, actorID uuid.UUID, actorName string) {
	if after.AssigneeID == nil {
		return
	}
	changed := before.AssigneeID == nil || *before.AssigneeID != *after.AssigneeID
	if !changed || *after.AssigneeID == actorID {
		return
	}
	prefs, err := h.store.GetNotificationPreferences(r.Context(), *after.AssigneeID)
	if err != nil || !prefs.AssignmentEnabled {
		return
	}
	_, err = h.store.CreateNotification(r.Context(), store.NotificationCreateInput{
		ProjectID: after.ProjectID,
		UserID:    *after.AssigneeID,
		TicketID:  after.ID,
		Type:      "assignment",
		Message:   fmt.Sprintf("%s assigned you to %s", actorName, after.Key),
	})
	if err != nil {
		logRequestError(r, "notification_assignment_create_failed", err)
		return
	}
	h.publishUserNotificationEvents(r.Context(), after.ProjectID, *after.AssigneeID)
}
