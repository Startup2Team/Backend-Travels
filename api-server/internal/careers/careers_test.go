package careers

import (
	"testing"
)

func TestValidation(t *testing.T) {
	input := CreateApplicationInput{
		FullName:           "John Candidate",
		Email:              "john@example.com",
		Phone:              "+250789000111",
		City:               "Kigali",
		WorkRight:          "CITIZEN",
		Status:             "STUDENT",
		Position:           "FULL_STACK",
		Technologies:       "Go, React, TypeScript",
		ProjectURL:         "https://github.com/example/project",
		ProjectBody:        "Built a ride hailing app",
		GithubURL:          "https://github.com/example",
		AvailableFromStart: true,
		Consent:            true,
	}

	if err := validate.Struct(input); err != nil {
		t.Fatalf("expected valid struct, got error: %v", err)
	}
}

func TestValidation_InvalidEmail(t *testing.T) {
	input := CreateApplicationInput{
		FullName:     "John Candidate",
		Email:        "invalid-email",
		Phone:        "+250789000111",
		City:         "Kigali",
		WorkRight:    "CITIZEN",
		Status:       "STUDENT",
		Position:     "FULL_STACK",
		Technologies: "Go",
		ProjectURL:   "https://github.com/example/project",
		ProjectBody:  "Built an app",
		GithubURL:    "https://github.com/example",
		Consent:      true,
	}

	if err := validate.Struct(input); err == nil {
		t.Fatalf("expected validation error for invalid email, got nil")
	}
}
