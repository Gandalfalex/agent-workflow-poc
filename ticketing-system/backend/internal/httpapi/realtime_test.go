package httpapi

import (
	"testing"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func TestProjectLiveHubPublishesToProjectSubscribers(t *testing.T) {
	hub := newProjectLiveHub()
	projectID := uuid.New()
	userID := uuid.New()

	sub, unsubscribe := hub.subscribe(projectID, userID)
	defer unsubscribe()

	evt := projectLiveEvent{
		Type:      projectEventBoardRefresh,
		ProjectID: openapi_types.UUID(projectID),
		Timestamp: time.Now().UTC(),
	}
	hub.publish(projectID, evt, nil)

	select {
	case got := <-sub.ch:
		if got.Type != evt.Type {
			t.Fatalf("expected type %q, got %q", evt.Type, got.Type)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected event to be published")
	}
}

func TestProjectLiveHubPublishesOnlyToTargetUser(t *testing.T) {
	hub := newProjectLiveHub()
	projectID := uuid.New()
	targetUser := uuid.New()
	otherUser := uuid.New()

	targetSub, unsubscribeTarget := hub.subscribe(projectID, targetUser)
	defer unsubscribeTarget()
	otherSub, unsubscribeOther := hub.subscribe(projectID, otherUser)
	defer unsubscribeOther()

	evt := projectLiveEvent{
		Type:      projectEventNotificationsUnread,
		ProjectID: openapi_types.UUID(projectID),
		Timestamp: time.Now().UTC(),
	}
	hub.publish(projectID, evt, &targetUser)

	select {
	case <-targetSub.ch:
		// expected
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("expected targeted subscriber to receive event")
	}

	select {
	case <-otherSub.ch:
		t.Fatalf("did not expect non-target subscriber to receive event")
	case <-time.After(200 * time.Millisecond):
		// expected
	}
}
