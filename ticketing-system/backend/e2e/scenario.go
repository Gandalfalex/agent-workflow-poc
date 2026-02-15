//go:build e2e

package e2e

import (
	"fmt"
	"testing"
	"time"
)

type Scenario struct {
	t       *testing.T
	harness *Harness
}

func NewScenario(t *testing.T, opts ...HarnessOption) *Scenario {
	t.Helper()
	return &Scenario{
		t:       t,
		harness: NewHarness(t, opts...),
	}
}

func (s *Scenario) Close() {
	s.t.Helper()
	s.harness.Close()
}

func (s *Scenario) SeedData() SeedData {
	return s.harness.SeedData()
}

func (s *Scenario) Harness() *Harness {
	return s.harness
}

func (s *Scenario) Given(description string, action func(*Scenario) error) *Scenario {
	return s.runStep("Given", description, action)
}

func (s *Scenario) When(description string, action func(*Scenario) error) *Scenario {
	return s.runStep("When", description, action)
}

func (s *Scenario) Then(description string, action func(*Scenario) error) *Scenario {
	return s.runStep("Then", description, action)
}

func (s *Scenario) And(description string, action func(*Scenario) error) *Scenario {
	return s.runStep("And", description, action)
}

func (s *Scenario) GivenAppIsRunning() *Scenario {
	return s.Given("the app is running", func(s *Scenario) error {
		return s.harness.HealthCheck()
	})
}

func (s *Scenario) WhenIGoTo(path string) *Scenario {
	return s.When(fmt.Sprintf("I go to page %q", path), func(s *Scenario) error {
		return s.harness.GoTo(path)
	})
}

func (s *Scenario) WhenIGoToRoute(routeKey string, params ...map[string]string) *Scenario {
	var routeParams map[string]string
	if len(params) > 0 {
		routeParams = params[0]
	}
	return s.When(fmt.Sprintf("I go to route %q", routeKey), func(s *Scenario) error {
		return s.harness.GoToRoute(routeKey, routeParams)
	})
}

func (s *Scenario) WhenIClick(selector string) *Scenario {
	return s.When(fmt.Sprintf("I click %q", selector), func(s *Scenario) error {
		return s.harness.Click(selector)
	})
}

func (s *Scenario) WhenIClickKey(selectorKey string) *Scenario {
	return s.When(fmt.Sprintf("I click selector key %q", selectorKey), func(s *Scenario) error {
		return s.harness.ClickKey(selectorKey)
	})
}

func (s *Scenario) WhenIFill(selector, value string) *Scenario {
	return s.When(fmt.Sprintf("I fill %q", selector), func(s *Scenario) error {
		return s.harness.Fill(selector, value)
	})
}

func (s *Scenario) WhenIFillKey(selectorKey, value string) *Scenario {
	return s.When(fmt.Sprintf("I fill selector key %q", selectorKey), func(s *Scenario) error {
		return s.harness.FillKey(selectorKey, value)
	})
}

func (s *Scenario) WhenILogInAs(identifier, password string) *Scenario {
	return s.When("I log in", func(s *Scenario) error {
		if err := s.harness.FillKey("login.identifier_input", identifier); err != nil {
			return err
		}
		if err := s.harness.FillKey("login.password_input", password); err != nil {
			return err
		}
		return s.harness.ClickKey("login.submit_button")
	})
}

func (s *Scenario) WhenIPress(selector, key string) *Scenario {
	return s.When(fmt.Sprintf("I press %q on %q", key, selector), func(s *Scenario) error {
		return s.harness.Press(selector, key)
	})
}

func (s *Scenario) WhenIPressKey(selectorKey, key string) *Scenario {
	return s.When(fmt.Sprintf("I press %q on selector key %q", key, selectorKey), func(s *Scenario) error {
		return s.harness.PressKey(selectorKey, key)
	})
}

func (s *Scenario) WhenISelectProjectByID(projectID string) *Scenario {
	return s.When(fmt.Sprintf("I select project %q", projectID), func(s *Scenario) error {
		return s.harness.SelectOptionByValueKey("nav.project_select", projectID)
	})
}

func (s *Scenario) ThenISeeSelector(selector string) *Scenario {
	return s.Then(fmt.Sprintf("I see selector %q", selector), func(s *Scenario) error {
		return s.harness.WaitVisible(selector)
	})
}

func (s *Scenario) ThenISeeSelectorKey(selectorKey string) *Scenario {
	return s.Then(fmt.Sprintf("I see selector key %q", selectorKey), func(s *Scenario) error {
		return s.harness.WaitVisibleKey(selectorKey)
	})
}

func (s *Scenario) ThenISeeText(text string) *Scenario {
	return s.Then(fmt.Sprintf("I see text %q", text), func(s *Scenario) error {
		return s.harness.ExpectTextVisible(text)
	})
}

func (s *Scenario) ThenURLContains(fragment string) *Scenario {
	return s.Then(fmt.Sprintf("the URL contains %q", fragment), func(s *Scenario) error {
		return s.harness.ExpectURLContains(fragment)
	})
}

func (s *Scenario) AndISeeSelector(selector string) *Scenario {
	return s.And(fmt.Sprintf("I see selector %q", selector), func(s *Scenario) error {
		return s.harness.WaitVisible(selector)
	})
}

func (s *Scenario) AndISeeSelectorKey(selectorKey string) *Scenario {
	return s.And(fmt.Sprintf("I see selector key %q", selectorKey), func(s *Scenario) error {
		return s.harness.WaitVisibleKey(selectorKey)
	})
}

func (s *Scenario) AndISeeText(text string) *Scenario {
	return s.And(fmt.Sprintf("I see text %q", text), func(s *Scenario) error {
		return s.harness.ExpectTextVisible(text)
	})
}

func (s *Scenario) ThenIDoNotSeeSelectorKey(selectorKey string) *Scenario {
	return s.Then(fmt.Sprintf("I do not see selector key %q", selectorKey), func(s *Scenario) error {
		return s.harness.ExpectSelectorHiddenKey(selectorKey)
	})
}

func (s *Scenario) AndIDoNotSeeSelectorKey(selectorKey string) *Scenario {
	return s.And(fmt.Sprintf("I do not see selector key %q", selectorKey), func(s *Scenario) error {
		return s.harness.ExpectSelectorHiddenKey(selectorKey)
	})
}

func (s *Scenario) WhenISelectOptionByValueKey(selectorKey, value string) *Scenario {
	return s.When(fmt.Sprintf("I select option %q on key %q", value, selectorKey), func(s *Scenario) error {
		return s.harness.SelectOptionByValueKey(selectorKey, value)
	})
}

func (s *Scenario) ThenISeeAtLeast(n int, selector string) *Scenario {
	return s.Then(fmt.Sprintf("I see at least %d elements matching %q", n, selector), func(s *Scenario) error {
		return s.harness.ExpectMinElements(selector, n)
	})
}

func (s *Scenario) WhenIWait(seconds int) *Scenario {
	return s.When(fmt.Sprintf("I wait %d seconds", seconds), func(s *Scenario) error {
		time.Sleep(time.Duration(seconds) * time.Second)
		return nil
	})
}

func (s *Scenario) WhenIClickLogout() *Scenario {
	return s.When("I click logout", func(s *Scenario) error {
		return s.harness.ClickKey("nav.logout_button")
	})
}

func (s *Scenario) WhenIClickRefresh() *Scenario {
	return s.When("I click refresh", func(s *Scenario) error {
		return s.harness.ClickKey("nav.refresh_button")
	})
}

func (s *Scenario) WhenINavigateToBoard() *Scenario {
	return s.When("I click the board tab", func(s *Scenario) error {
		return s.harness.ClickKey("nav.board_tab")
	})
}

func (s *Scenario) WhenINavigateToSettings() *Scenario {
	return s.When("I click the settings tab", func(s *Scenario) error {
		return s.harness.ClickKey("nav.settings_tab")
	})
}

func (s *Scenario) WhenICreateStory(title string) *Scenario {
	return s.When(fmt.Sprintf("I create story %q", title), func(s *Scenario) error {
		if err := s.harness.ClickKey("board.create_story_button"); err != nil {
			return err
		}
		if err := s.harness.WaitVisibleKey("story.modal"); err != nil {
			return err
		}
		if err := s.harness.FillKey("story.title_input", title); err != nil {
			return err
		}
		return s.harness.ClickKey("story.create_button")
	})
}

func (s *Scenario) WhenIOpenNewTicketModal() *Scenario {
	return s.When("I open the new ticket modal", func(s *Scenario) error {
		return s.harness.ClickKey("board.add_ticket_button")
	})
}

func (s *Scenario) WhenICreateTicket(title string) *Scenario {
	return s.When(fmt.Sprintf("I create ticket %q", title), func(s *Scenario) error {
		if err := s.harness.ClickKey("board.add_ticket_button"); err != nil {
			return err
		}
		if err := s.harness.WaitVisibleKey("new_ticket.modal"); err != nil {
			return err
		}
		if err := s.harness.FillKey("new_ticket.title_input", title); err != nil {
			return err
		}
		return s.harness.ClickKey("new_ticket.create_button")
	})
}

func (s *Scenario) WhenIClickTicketByText(text string) *Scenario {
	return s.When(fmt.Sprintf("I click on ticket with text %q", text), func(s *Scenario) error {
		if err := s.harness.ExpectTextVisible(text); err != nil {
			return err
		}
		return s.harness.page.GetByText(text).First().Click()
	})
}

func (s *Scenario) ThenIDoNotSeeText(text string) *Scenario {
	return s.Then(fmt.Sprintf("I do not see text %q", text), func(s *Scenario) error {
		return s.harness.ExpectTextHidden(text)
	})
}

func (s *Scenario) AndIDoNotSeeText(text string) *Scenario {
	return s.And(fmt.Sprintf("I do not see text %q", text), func(s *Scenario) error {
		return s.harness.ExpectTextHidden(text)
	})
}

func (s *Scenario) ThenButtonIsDisabledKey(selectorKey string) *Scenario {
	return s.Then(fmt.Sprintf("button %q is disabled", selectorKey), func(s *Scenario) error {
		disabled, err := s.harness.IsButtonDisabledKey(selectorKey)
		if err != nil {
			return err
		}
		if !disabled {
			return fmt.Errorf("expected button %q to be disabled", selectorKey)
		}
		return nil
	})
}

func (s *Scenario) ThenButtonIsEnabledKey(selectorKey string) *Scenario {
	return s.Then(fmt.Sprintf("button %q is enabled", selectorKey), func(s *Scenario) error {
		disabled, err := s.harness.IsButtonDisabledKey(selectorKey)
		if err != nil {
			return err
		}
		if disabled {
			return fmt.Errorf("expected button %q to be enabled", selectorKey)
		}
		return nil
	})
}

func (s *Scenario) WhenIAcceptNextDialog() *Scenario {
	return s.When("I register dialog auto-accept", func(s *Scenario) error {
		s.harness.HandleNextDialog(true)
		return nil
	})
}

func (s *Scenario) runStep(keyword, description string, action func(*Scenario) error) *Scenario {
	s.t.Helper()
	step := fmt.Sprintf("%s %s", keyword, description)
	s.t.Log(step)

	if action == nil {
		s.harness.failStep(step, fmt.Errorf("step action is nil"))
		return s
	}
	if err := action(s); err != nil {
		s.harness.failStep(step, err)
	}
	return s
}
