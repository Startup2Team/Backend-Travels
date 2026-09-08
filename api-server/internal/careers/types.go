package careers

import "time"

type Application struct {
	ID                 string     `json:"id"`
	FullName           string     `json:"full_name"`
	Email              string     `json:"email"`
	Phone              string     `json:"phone"`
	City               string     `json:"city"`
	WorkRight          string     `json:"work_right"`
	Status             string     `json:"status"`
	Institution        *string    `json:"institution,omitempty"`
	GraduationYear     *string    `json:"graduation_year,omitempty"`
	Position           string     `json:"position"`
	Technologies       string     `json:"technologies"`
	ProjectURL         string     `json:"project_url"`
	ProjectBody        string     `json:"project_body"`
	GithubURL          string     `json:"github_url"`
	LinkedinURL        *string    `json:"linkedin_url,omitempty"`
	PortfolioURL       *string    `json:"portfolio_url,omitempty"`
	CVURL              *string    `json:"cv_url,omitempty"`
	AvailableFromStart bool       `json:"available_from_start"`
	HeardFrom          *string    `json:"heard_from,omitempty"`
	Consent            bool       `json:"consent"`
	Source             string     `json:"source"`
	ApplicationStatus  string     `json:"application_status"` // NEW, UNDER_REVIEW, INTERVIEW_SCHEDULED, ACCEPTED, REJECTED
	ReviewerNotes      *string    `json:"reviewer_notes,omitempty"`
	ReviewedBy         *string    `json:"reviewed_by,omitempty"`
	ReviewedAt         *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type CreateApplicationInput struct {
	FullName           string  `json:"full_name"            validate:"required,min=2,max=255"`
	Email              string  `json:"email"                validate:"required,email"`
	Phone              string  `json:"phone"                validate:"required,min=5,max=50"`
	City               string  `json:"city"                 validate:"required"`
	WorkRight          string  `json:"work_right"           validate:"required,oneof=CITIZEN PERMIT NEITHER"`
	Status             string  `json:"status"               validate:"required,oneof=STUDENT GRADUATE EMPLOYED"`
	Institution        *string `json:"institution"`
	GraduationYear     *string `json:"graduation_year"`
	Position           string  `json:"position"             validate:"required"`
	Technologies       string  `json:"technologies"         validate:"required"`
	ProjectURL         string  `json:"project_url"          validate:"required,url"`
	ProjectBody        string  `json:"project_body"         validate:"required"`
	GithubURL          string  `json:"github_url"           validate:"required,url"`
	LinkedinURL        *string `json:"linkedin_url"`
	PortfolioURL       *string `json:"portfolio_url"`
	CVURL              *string `json:"cv_url"`
	AvailableFromStart bool    `json:"available_from_start"`
	HeardFrom          *string `json:"heard_from"`
	Consent            bool    `json:"consent"              validate:"required"`
	Source             string  `json:"source"`
}

type UpdateStatusInput struct {
	ApplicationStatus string  `json:"application_status" validate:"required,oneof=NEW UNDER_REVIEW INTERVIEW_SCHEDULED ACCEPTED REJECTED"`
	ReviewerNotes     *string `json:"reviewer_notes"`
}

type ListFilter struct {
	Position          string
	ApplicationStatus string
	Search            string
	Limit             int
	Offset            int
}
