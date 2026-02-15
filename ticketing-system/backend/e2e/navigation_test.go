//go:build e2e

package e2e

import "testing"

func TestLogoutRedirectsToLoginScreen(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenIClickLogout().
		ThenISeeSelectorKey("login.view").
		AndISeeSelectorKey("login.submit_button")
}

func TestNavigateBetweenBoardAndSettings(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		WhenINavigateToSettings().
		ThenURLContains("/projects/" + seed.ProjectID + "/settings").
		WhenINavigateToBoard().
		ThenURLContains("/projects/" + seed.ProjectID + "/board").
		AndISeeSelectorKey("board.add_ticket_button")
}

func TestRefreshBoardPreservesView(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	seed := scenario.SeedData()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		WhenILogInAs("AdminUser", "admin123").
		WhenISelectProjectByID(seed.ProjectID).
		ThenISeeSelectorKey("board.add_ticket_button").
		WhenIClickRefresh().
		ThenISeeSelectorKey("board.add_ticket_button").
		AndISeeText("E2E Story")
}

func TestUnauthenticatedAccessShowsLogin(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	// Navigate to home without authentication - should show login screen
	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		ThenISeeSelectorKey("login.view").
		AndISeeSelectorKey("login.identifier_input").
		AndISeeSelectorKey("login.password_input")
}
