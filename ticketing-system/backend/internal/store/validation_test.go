package store

import (
	"testing"
)

func TestNormalizeProjectKey(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "valid uppercase key",
			input:    "PROJ",
			expected: "PROJ",
		},
		{
			name:     "lowercase converted to uppercase",
			input:    "proj",
			expected: "PROJ",
		},
		{
			name:     "mixed case converted to uppercase",
			input:    "PrOj",
			expected: "PROJ",
		},
		{
			name:     "with numbers",
			input:    "PR01",
			expected: "PR01",
		},
		{
			name:     "all numbers",
			input:    "1234",
			expected: "1234",
		},
		{
			name:     "trimmed whitespace",
			input:    "  PROJ  ",
			expected: "PROJ",
		},
		{
			name:        "too short",
			input:       "PRO",
			expectError: true,
		},
		{
			name:        "too long",
			input:       "PROJX",
			expectError: true,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "    ",
			expectError: true,
		},
		{
			name:        "contains special characters",
			input:       "PR-1",
			expectError: true,
		},
		{
			name:        "contains space",
			input:       "PR 1",
			expectError: true,
		},
		{
			name:        "contains underscore",
			input:       "PR_1",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeProjectKey(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %q, got result %q", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestNormalizePriority(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "low",
			input:    "low",
			expected: "low",
		},
		{
			name:     "medium",
			input:    "medium",
			expected: "medium",
		},
		{
			name:     "high",
			input:    "high",
			expected: "high",
		},
		{
			name:     "urgent",
			input:    "urgent",
			expected: "urgent",
		},
		{
			name:     "uppercase LOW",
			input:    "LOW",
			expected: "low",
		},
		{
			name:     "mixed case MeDiUm",
			input:    "MeDiUm",
			expected: "medium",
		},
		{
			name:     "with whitespace",
			input:    "  high  ",
			expected: "high",
		},
		{
			name:     "empty defaults to medium",
			input:    "",
			expected: "medium",
		},
		{
			name:     "whitespace only defaults to medium",
			input:    "   ",
			expected: "medium",
		},
		{
			name:     "invalid value defaults to medium",
			input:    "critical",
			expected: "medium",
		},
		{
			name:     "another invalid value defaults to medium",
			input:    "normal",
			expected: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePriority(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePriority(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeTicketType(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "feature",
			input:    "feature",
			expected: "feature",
		},
		{
			name:     "bug",
			input:    "bug",
			expected: "bug",
		},
		{
			name:     "uppercase FEATURE",
			input:    "FEATURE",
			expected: "feature",
		},
		{
			name:     "uppercase BUG",
			input:    "BUG",
			expected: "bug",
		},
		{
			name:     "mixed case FeAtUrE",
			input:    "FeAtUrE",
			expected: "feature",
		},
		{
			name:     "with whitespace",
			input:    "  bug  ",
			expected: "bug",
		},
		{
			name:     "empty defaults to feature",
			input:    "",
			expected: "feature",
		},
		{
			name:     "whitespace only defaults to feature",
			input:    "   ",
			expected: "feature",
		},
		{
			name:        "invalid type task",
			input:       "task",
			expectError: true,
		},
		{
			name:        "invalid type epic",
			input:       "epic",
			expectError: true,
		},
		{
			name:        "invalid type story",
			input:       "story",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeTicketType(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %q, got result %q", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("normalizeTicketType(%q) = %q, expected %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestNormalizeProjectRole(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "admin",
			input:    "admin",
			expected: "admin",
		},
		{
			name:     "contributor",
			input:    "contributor",
			expected: "contributor",
		},
		{
			name:     "viewer",
			input:    "viewer",
			expected: "viewer",
		},
		{
			name:     "uppercase ADMIN",
			input:    "ADMIN",
			expected: "admin",
		},
		{
			name:     "mixed case CoNtRiBuToR",
			input:    "CoNtRiBuToR",
			expected: "contributor",
		},
		{
			name:     "with whitespace",
			input:    "  viewer  ",
			expected: "viewer",
		},
		{
			name:        "empty",
			input:       "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectError: true,
		},
		{
			name:        "invalid role owner",
			input:       "owner",
			expectError: true,
		},
		{
			name:        "invalid role member",
			input:       "member",
			expectError: true,
		},
		{
			name:        "invalid role editor",
			input:       "editor",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeProjectRole(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %q, got result %q", tt.input, result)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("normalizeProjectRole(%q) = %q, expected %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

func TestValidateWebhookURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:  "valid http URL",
			input: "http://example.com/webhook",
		},
		{
			name:  "valid https URL",
			input: "https://example.com/webhook",
		},
		{
			name:  "valid URL with port",
			input: "http://localhost:8080/webhook",
		},
		{
			name:  "valid URL with path and query",
			input: "https://api.example.com/v1/webhook?token=abc",
		},
		{
			name:  "valid URL trimmed",
			input: "  https://example.com/webhook  ",
		},
		{
			name:        "empty URL",
			input:       "",
			expectError: true,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectError: true,
		},
		{
			name:        "missing scheme",
			input:       "example.com/webhook",
			expectError: true,
		},
		{
			name:        "missing host",
			input:       "http:///webhook",
			expectError: true,
		},
		{
			name:        "invalid URL",
			input:       "not a url",
			expectError: true,
		},
		{
			name:        "relative path only",
			input:       "/webhook",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookURL(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for input %q", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for input %q: %v", tt.input, err)
				}
			}
		})
	}
}

func TestValidateWebhookEvents(t *testing.T) {
	tests := []struct {
		name        string
		events      []string
		expectError bool
	}{
		{
			name:   "single valid event ticket.created",
			events: []string{"ticket.created"},
		},
		{
			name:   "single valid event ticket.updated",
			events: []string{"ticket.updated"},
		},
		{
			name:   "single valid event ticket.deleted",
			events: []string{"ticket.deleted"},
		},
		{
			name:   "single valid event ticket.state_changed",
			events: []string{"ticket.state_changed"},
		},
		{
			name:   "multiple valid events",
			events: []string{"ticket.created", "ticket.updated", "ticket.deleted"},
		},
		{
			name:   "all valid events",
			events: []string{"ticket.created", "ticket.updated", "ticket.deleted", "ticket.state_changed"},
		},
		{
			name:        "invalid event",
			events:      []string{"ticket.invalid"},
			expectError: true,
		},
		{
			name:        "mixed valid and invalid",
			events:      []string{"ticket.created", "invalid.event"},
			expectError: true,
		},
		{
			name:        "empty event string",
			events:      []string{""},
			expectError: true,
		},
		{
			name:        "project.created not allowed",
			events:      []string{"project.created"},
			expectError: true,
		},
		{
			name:        "comment.created not allowed",
			events:      []string{"comment.created"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateWebhookEvents(tt.events)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for events %v", tt.events)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for events %v: %v", tt.events, err)
				}
			}
		})
	}
}

func TestProjectKeyPattern(t *testing.T) {
	validKeys := []string{"PROJ", "ABCD", "TEST", "PR01", "1234", "A1B2"}
	for _, key := range validKeys {
		if !projectKeyPattern.MatchString(key) {
			t.Errorf("expected %q to match project key pattern", key)
		}
	}

	invalidKeys := []string{"", "PRO", "PROJX", "proj", "PR-1", "PR_1", "PR 1", "PR.1"}
	for _, key := range invalidKeys {
		if projectKeyPattern.MatchString(key) {
			t.Errorf("expected %q to NOT match project key pattern", key)
		}
	}
}
