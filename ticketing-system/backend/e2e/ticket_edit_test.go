//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"ticketing-system/backend/internal/store"
)

func seedTicketForEdit(s *Scenario, projectID, storyID, stateID uuid.UUID, input store.TicketCreateInput) (store.Ticket, error) {
	input.StoryID = storyID
	input.StateID = &stateID
	ticket, err := s.Harness().Store().CreateTicket(s.Harness().Context(), projectID, input)
	if err != nil {
		return store.Ticket{}, fmt.Errorf("seed ticket %q: %w", input.Title, err)
	}
	return ticket, nil
}

func navigateAndOpenTicket(s *Scenario, projectID, title string) *Scenario {
	return s.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(projectID).
		ThenURLContains("/projects/" + projectID + "/board").
		WhenIClickRefresh().
		ThenISeeText(title).
		WhenIClickTicketByText(title).
		ThenISeeSelectorKey("ticket.modal")
}

func TestOpenTicketAndEditTitle(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	originalTitle := fmt.Sprintf("Original Title %d", ts)
	updatedTitle := fmt.Sprintf("Updated Title %d", ts)

	scenario.
		Given("a ticket with the original title is pre-seeded into the project", func(s *Scenario) error {
			_, err := seedTicketForEdit(s, projectID, storyID, backlogID, store.TicketCreateInput{
				Title: originalTitle,
				Type:  "feature",
			})
			return err
		})

	navigateAndOpenTicket(scenario, seed.ProjectID, originalTitle).
		When("I clear and type the new title", func(s *Scenario) error {
			sel, err := s.Harness().Selector("ticket.title_input")
			if err != nil {
				return err
			}
			if err := s.Harness().WaitVisible(sel); err != nil {
				return err
			}
			if err := s.Harness().page.Locator(sel).Fill(updatedTitle); err != nil {
				return fmt.Errorf("fill new title: %w", err)
			}
			return nil
		}).
		WhenIClickKey("ticket.save_button").
		ThenISeeText(updatedTitle)
}

func TestOpenTicketAndChangeState(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	title := fmt.Sprintf("State Change Ticket %d", time.Now().UnixNano())

	scenario.
		Given("a backlog ticket is pre-seeded", func(s *Scenario) error {
			_, err := seedTicketForEdit(s, projectID, storyID, backlogID, store.TicketCreateInput{
				Title: title,
				Type:  "feature",
			})
			return err
		})

	navigateAndOpenTicket(scenario, seed.ProjectID, title).
		WhenISelectOptionByValueKey("ticket.state_select", seed.InProgressID).
		WhenIClickKey("ticket.save_button").
		ThenISeeText(title)
}

func TestOpenTicketAndChangePriority(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	title := fmt.Sprintf("Priority Change Ticket %d", time.Now().UnixNano())

	scenario.
		Given("a low-priority ticket is pre-seeded", func(s *Scenario) error {
			_, err := seedTicketForEdit(s, projectID, storyID, backlogID, store.TicketCreateInput{
				Title:    title,
				Type:     "feature",
				Priority: "low",
			})
			return err
		})

	navigateAndOpenTicket(scenario, seed.ProjectID, title).
		WhenISelectOptionByValueKey("ticket.priority_select", "urgent").
		WhenIClickKey("ticket.save_button").
		ThenISeeText(title)
}

func TestOpenTicketAndChangeType(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	title := fmt.Sprintf("Type Change Ticket %d", time.Now().UnixNano())

	scenario.
		Given("a feature ticket is pre-seeded", func(s *Scenario) error {
			_, err := seedTicketForEdit(s, projectID, storyID, backlogID, store.TicketCreateInput{
				Title: title,
				Type:  "feature",
			})
			return err
		})

	navigateAndOpenTicket(scenario, seed.ProjectID, title).
		WhenISelectOptionByValueKey("ticket.type_select", "bug").
		WhenIClickKey("ticket.save_button").
		ThenISeeText(title)
}

func TestOpenTicketAndAddComment(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	title := fmt.Sprintf("Commentable Ticket %d", ts)
	commentText := fmt.Sprintf("This is a test comment %d", ts)

	scenario.
		Given("a ticket is pre-seeded for commenting", func(s *Scenario) error {
			_, err := seedTicketForEdit(s, projectID, storyID, backlogID, store.TicketCreateInput{
				Title: title,
				Type:  "feature",
			})
			return err
		})

	navigateAndOpenTicket(scenario, seed.ProjectID, title).
		WhenIFillKey("ticket.comment_input", commentText).
		WhenIClickKey("ticket.post_comment_button").
		ThenISeeText(commentText)
}

func TestOpenTicketEditMultipleFieldsAndSave(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	projectID := uuid.MustParse(seed.ProjectID)
	storyID := uuid.MustParse(seed.StoryID)
	backlogID := uuid.MustParse(seed.BacklogID)

	ts := time.Now().UnixNano()
	originalTitle := fmt.Sprintf("Multi Edit Original %d", ts)
	updatedTitle := fmt.Sprintf("Multi Edit Updated %d", ts)

	scenario.
		Given("a low-priority feature ticket is pre-seeded for multi-field editing", func(s *Scenario) error {
			_, err := seedTicketForEdit(s, projectID, storyID, backlogID, store.TicketCreateInput{
				Title:    originalTitle,
				Type:     "feature",
				Priority: "low",
			})
			return err
		})

	navigateAndOpenTicket(scenario, seed.ProjectID, originalTitle).
		When("I update the title", func(s *Scenario) error {
			sel, err := s.Harness().Selector("ticket.title_input")
			if err != nil {
				return err
			}
			return s.Harness().page.Locator(sel).Fill(updatedTitle)
		}).
		WhenISelectOptionByValueKey("ticket.type_select", "bug").
		WhenISelectOptionByValueKey("ticket.priority_select", "urgent").
		WhenISelectOptionByValueKey("ticket.state_select", seed.InProgressID).
		WhenIClickKey("ticket.save_button").
		ThenISeeText(updatedTitle)
}
