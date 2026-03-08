//go:build e2e

package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

// seedAttachmentTicket creates a ticket directly via the store for attachment tests.
func seedAttachmentTicket(s *Scenario, projectID, storyID, backlogID uuid.UUID, title string) error {
	_, err := s.Harness().Store().CreateTicket(s.Harness().Context(), projectID, store.TicketCreateInput{
		Title:   title,
		Type:    "feature",
		StoryID: storyID,
		StateID: &backlogID,
	})
	if err != nil {
		return fmt.Errorf("seed ticket: %w", err)
	}
	return nil
}

// writeTempFile writes content to a named file inside dir and returns the full path.
func writeTempFile(dir, name string, content []byte) (string, error) {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		return "", fmt.Errorf("create temp file %q: %w", name, err)
	}
	return path, nil
}

func TestUploadAndListAttachment(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Attachment Upload Ticket %d", ts)

	tmpDir := t.TempDir()
	var testFile string

	scenario.
		Given("a ticket exists for attachment upload testing", func(s *Scenario) error {
			return seedAttachmentTicket(s, projectID, storyID, backlogID, title)
		}).
		And("a local test file is prepared", func(s *Scenario) error {
			var err error
			testFile, err = writeTempFile(tmpDir, "test-upload.txt", []byte("hello e2e attachment test"))
			return err
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		ThenISeeSelectorKey("ticket.attachments_section").
		WhenIUploadFileViaInput(testFile).
		ThenISeeText("test-upload.txt").
		AndISeeSelectorKey("ticket.attachment_item").
		AndISeeSelectorKey("ticket.attachment_download_link")
}

func TestDeleteAttachment(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Attachment Delete Ticket %d", ts)

	tmpDir := t.TempDir()
	var testFile string

	scenario.
		Given("a ticket exists for attachment deletion testing", func(s *Scenario) error {
			return seedAttachmentTicket(s, projectID, storyID, backlogID, title)
		}).
		And("a local test file is prepared", func(s *Scenario) error {
			var err error
			testFile, err = writeTempFile(tmpDir, "delete-me.txt", []byte("file to be deleted"))
			return err
		}).
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal").
		WhenIUploadFileViaInput(testFile).
		ThenISeeText("delete-me.txt").
		WhenIClickKey("ticket.attachment_delete_button").
		ThenIDoNotSeeText("delete-me.txt")
}
