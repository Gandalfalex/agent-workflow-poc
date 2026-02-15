//go:build e2e

package e2e

import "testing"

func TestLoginScreenLoadsFromBackendServedFrontend(t *testing.T) {
	t.Parallel()

	scenario := NewScenario(t)
	defer scenario.Close()

	scenario.
		GivenAppIsRunning().
		WhenIGoToRoute("home").
		ThenISeeSelectorKey("login.view").
		AndISeeText("Sign in to the workspace").
		AndISeeSelectorKey("login.identifier_input").
		AndISeeSelectorKey("login.password_input").
		AndISeeSelectorKey("login.submit_button")
}
