package httpapi

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/net/websocket"
)

const (
	projectEventHeartbeat           = "heartbeat"
	projectEventNotificationsUnread = "notifications.unread_count"
	projectEventNotificationsChange = "notifications.changed"
	projectEventBoardRefresh        = "board.refresh"
	projectEventActivityChanged     = "activity.changed"
)

type projectLiveEvent struct {
	Type      string             `json:"type"`
	ProjectID openapi_types.UUID `json:"projectId"`
	Timestamp time.Time          `json:"timestamp"`
	Payload   map[string]any     `json:"payload,omitempty"`
}

type projectLiveSubscriber struct {
	userID uuid.UUID
	ch     chan projectLiveEvent
}

type projectLiveHub struct {
	mu          sync.RWMutex
	subscribers map[uuid.UUID]map[*projectLiveSubscriber]struct{}
}

func newProjectLiveHub() *projectLiveHub {
	return &projectLiveHub{
		subscribers: map[uuid.UUID]map[*projectLiveSubscriber]struct{}{},
	}
}

func (h *projectLiveHub) subscribe(projectID, userID uuid.UUID) (*projectLiveSubscriber, func()) {
	sub := &projectLiveSubscriber{
		userID: userID,
		ch:     make(chan projectLiveEvent, 32),
	}

	h.mu.Lock()
	if _, ok := h.subscribers[projectID]; !ok {
		h.subscribers[projectID] = map[*projectLiveSubscriber]struct{}{}
	}
	h.subscribers[projectID][sub] = struct{}{}
	h.mu.Unlock()

	return sub, func() {
		h.mu.Lock()
		if set, ok := h.subscribers[projectID]; ok {
			delete(set, sub)
			if len(set) == 0 {
				delete(h.subscribers, projectID)
			}
		}
		h.mu.Unlock()
		close(sub.ch)
	}
}

func (h *projectLiveHub) publish(projectID uuid.UUID, evt projectLiveEvent, userID *uuid.UUID) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	set, ok := h.subscribers[projectID]
	if !ok {
		return
	}
	for sub := range set {
		if userID != nil && sub.userID != *userID {
			continue
		}
		select {
		case sub.ch <- evt:
		default:
			// Drop when client is slow; frontend falls back to polling as needed.
		}
	}
}

func isWebSocketUpgradeRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

func (h *API) StreamProjectEvents(w http.ResponseWriter, r *http.Request, projectId openapi_types.UUID) {
	projectUUID := uuid.UUID(projectId)
	user, ok := authUser(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "missing session")
		return
	}
	userID, err := uuid.Parse(user.ID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized", "invalid session")
		return
	}
	if !h.requireProjectAccess(w, r, projectUUID) {
		return
	}
	if !isWebSocketUpgradeRequest(r) {
		writeError(w, http.StatusUpgradeRequired, "upgrade_required", "websocket upgrade required")
		return
	}

	server := websocket.Server{
		Handler: websocket.Handler(func(conn *websocket.Conn) {
			defer conn.Close()
			sub, unsubscribe := h.live.subscribe(projectUUID, userID)
			defer unsubscribe()
			heartbeat := time.NewTicker(20 * time.Second)
			defer heartbeat.Stop()

			done := make(chan struct{})
			go func() {
				defer close(done)
				for {
					var msg any
					if recvErr := websocket.JSON.Receive(conn, &msg); recvErr != nil {
						return
					}
				}
			}()

			ctx := r.Context()
			for {
				select {
				case <-ctx.Done():
					return
				case <-done:
					return
				case <-heartbeat.C:
					if sendErr := websocket.JSON.Send(conn, projectLiveEvent{
						Type:      projectEventHeartbeat,
						ProjectID: openapi_types.UUID(projectUUID),
						Timestamp: time.Now().UTC(),
					}); sendErr != nil {
						return
					}
				case evt, ok := <-sub.ch:
					if !ok {
						return
					}
					if sendErr := websocket.JSON.Send(conn, evt); sendErr != nil {
						return
					}
				}
			}
		}),
		Handshake: func(_ *websocket.Config, _ *http.Request) error { return nil },
	}

	server.ServeHTTP(w, r)
}

func (h *API) publishProjectLiveEvent(projectID uuid.UUID, eventType string, payload map[string]any) {
	if h.live == nil {
		return
	}
	h.live.publish(projectID, projectLiveEvent{
		Type:      eventType,
		ProjectID: openapi_types.UUID(projectID),
		Timestamp: time.Now().UTC(),
		Payload:   payload,
	}, nil)
}

func (h *API) publishUserNotificationEvents(ctx context.Context, projectID, userID uuid.UUID) {
	if h.live == nil {
		return
	}
	uid := userID
	count, err := h.store.CountUnreadNotifications(ctx, projectID, userID)
	if err == nil {
		h.live.publish(projectID, projectLiveEvent{
			Type:      projectEventNotificationsUnread,
			ProjectID: openapi_types.UUID(projectID),
			Timestamp: time.Now().UTC(),
			Payload: map[string]any{
				"userId": userID.String(),
				"count":  count,
			},
		}, &uid)
	}
	h.live.publish(projectID, projectLiveEvent{
		Type:      projectEventNotificationsChange,
		ProjectID: openapi_types.UUID(projectID),
		Timestamp: time.Now().UTC(),
		Payload: map[string]any{
			"userId": userID.String(),
		},
	}, &uid)
}
