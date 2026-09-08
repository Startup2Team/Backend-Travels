package careers

import (
	"context"
	"fmt"
	"strings"

	"github.com/rs/zerolog"
	"github.com/workspace/ride-platform/internal/email"
)

type Service struct {
	repo *Repository
	log  zerolog.Logger
}

func NewService(repo *Repository, log zerolog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) Submit(ctx context.Context, input CreateApplicationInput) (*Application, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.FullName = strings.TrimSpace(input.FullName)
	input.Phone = strings.TrimSpace(input.Phone)

	app, err := s.repo.Create(ctx, input)
	if err != nil {
		s.log.Error().Err(err).Str("email", input.Email).Msg("careers: failed to save application")
		return nil, err
	}

	s.log.Info().
		Str("application_id", app.ID).
		Str("position", app.Position).
		Str("email", app.Email).
		Msg("careers: new candidate application received")

	// Send automated confirmation email asynchronously to applicant's personal email
	go func(candidateName, candidateEmail, position string) {
		html := email.BuildCareerApplicationReceivedEmail(candidateName, position)
		subject := fmt.Sprintf("Application Received — %s Position at Rides", position)
		if err := email.SendEmail(context.Background(), candidateEmail, subject, html); err != nil {
			s.log.Warn().Err(err).Str("email", candidateEmail).Msg("careers: failed to send application confirmation email")
		} else {
			s.log.Info().Str("email", candidateEmail).Msg("careers: application confirmation email sent successfully")
		}
	}(app.FullName, app.Email, app.Position)

	return app, nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]*Application, int, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id string) (*Application, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) UpdateStatus(ctx context.Context, id, status string, notes *string, reviewerID string) (*Application, error) {
	app, err := s.repo.UpdateStatus(ctx, id, status, notes, reviewerID)
	if err != nil {
		return nil, err
	}

	// Send automated status update email asynchronously to candidate's personal email
	if app != nil {
		go func(candidateName, candidateEmail, position, newStatus string) {
			html, subject := email.BuildCareerStatusChangeEmail(candidateName, position, newStatus)
			if html != "" {
				if err := email.SendEmail(context.Background(), candidateEmail, subject, html); err != nil {
					s.log.Warn().Err(err).Str("email", candidateEmail).Str("status", newStatus).Msg("careers: failed to send status change email")
				} else {
					s.log.Info().Str("email", candidateEmail).Str("status", newStatus).Msg("careers: candidate status change email sent successfully")
				}
			}
		}(app.FullName, app.Email, app.Position, app.ApplicationStatus)
	}

	return app, nil
}

func (s *Service) ExportCSV(ctx context.Context, filter ListFilter) ([]byte, error) {
	filter.Limit = 10000 // Fetch all matching records for CSV export
	apps, _, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	var sb strings.Builder
	// CSV Header
	sb.WriteString("ID,Submitted At,Full Name,Email,Phone,City,Position,Status,Work Right,Institution,Graduation Year,Technologies,GitHub,LinkedIn,Portfolio,CV URL,App Status,Reviewer Notes\n")

	for _, app := range apps {
		institution := ""
		if app.Institution != nil {
			institution = *app.Institution
		}
		gradYear := ""
		if app.GraduationYear != nil {
			gradYear = *app.GraduationYear
		}
		linkedin := ""
		if app.LinkedinURL != nil {
			linkedin = *app.LinkedinURL
		}
		portfolio := ""
		if app.PortfolioURL != nil {
			portfolio = *app.PortfolioURL
		}
		cv := ""
		if app.CVURL != nil {
			cv = *app.CVURL
		}
		notes := ""
		if app.ReviewerNotes != nil {
			notes = *app.ReviewerNotes
		}

		sb.WriteString(fmt.Sprintf("%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q,%q\n",
			app.ID,
			app.CreatedAt.Format("2006-01-02 15:04:05"),
			app.FullName,
			app.Email,
			app.Phone,
			app.City,
			app.Position,
			app.Status,
			app.WorkRight,
			institution,
			gradYear,
			app.Technologies,
			app.GithubURL,
			linkedin,
			portfolio,
			cv,
			app.ApplicationStatus,
			notes,
		))
	}

	return []byte(sb.String()), nil
}
