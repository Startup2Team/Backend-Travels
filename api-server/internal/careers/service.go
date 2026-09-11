package careers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/workspace/ride-platform/internal/email"
	apperrors "github.com/workspace/ride-platform/pkg/errors"
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

func parseFlexTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

func (s *Service) checkLimits(ctx context.Context) error {
	now := time.Now().UTC()

	// 1. Check Dynamic DB Settings
	settings, err := s.repo.GetSettings(ctx)
	if err == nil && settings != nil {
		if !settings.IsOpen {
			msg := settings.ClosedMessage
			if msg == "" {
				msg = "Applications for this recruitment cycle are currently closed."
			}
			return apperrors.New(http.StatusBadRequest, "APPLICATIONS_CLOSED", msg)
		}
		if settings.OpenAt != nil && now.Before(*settings.OpenAt) {
			return apperrors.New(http.StatusBadRequest, "APPLICATIONS_NOT_OPEN", fmt.Sprintf("Applications are not open yet. Application period starts on %s.", settings.OpenAt.Format("2006-01-02 15:04 UTC")))
		}
		if settings.CloseAt != nil && now.After(*settings.CloseAt) {
			return apperrors.New(http.StatusBadRequest, "APPLICATIONS_CLOSED", "Applications for this recruitment cycle have closed.")
		}
		if settings.MaxApplications > 0 && settings.TotalSubmitted >= settings.MaxApplications {
			return apperrors.New(http.StatusBadRequest, "QUOTA_REACHED", fmt.Sprintf("Application quota reached (%d/%d max submissions). Submissions are closed for this cycle.", settings.TotalSubmitted, settings.MaxApplications))
		}
	}

	// 2. Env Var Fallbacks
	openAtStr := os.Getenv("CAREERS_OPEN_AT")
	if openAtStr != "" {
		if openAt, err := parseFlexTime(openAtStr); err == nil && now.Before(openAt) {
			return apperrors.New(http.StatusBadRequest, "APPLICATIONS_NOT_OPEN", fmt.Sprintf("Applications are not open yet. Application period starts on %s.", openAt.Format("2006-01-02 15:04 UTC")))
		}
	}

	closeAtStr := os.Getenv("CAREERS_CLOSE_AT")
	if closeAtStr != "" {
		if closeAt, err := parseFlexTime(closeAtStr); err == nil && now.After(closeAt) {
			return apperrors.New(http.StatusBadRequest, "APPLICATIONS_CLOSED", "Applications for this recruitment cycle have closed.")
		}
	}

	maxAppsStr := os.Getenv("CAREERS_MAX_APPLICATIONS")
	if maxAppsStr != "" {
		var maxApps int
		if _, err := fmt.Sscanf(maxAppsStr, "%d", &maxApps); err == nil && maxApps > 0 {
			currentCount, err := s.repo.CountTotal(ctx)
			if err == nil && currentCount >= maxApps {
				return apperrors.New(http.StatusBadRequest, "QUOTA_REACHED", fmt.Sprintf("Application quota reached (%d/%d max submissions). Submissions are closed for this cycle.", currentCount, maxApps))
			}
		}
	}

	return nil
}

func (s *Service) GetSettings(ctx context.Context) (*CareerSettings, error) {
	return s.repo.GetSettings(ctx)
}

func (s *Service) UpdateSettings(ctx context.Context, input UpdateSettingsInput) (*CareerSettings, error) {
	return s.repo.UpdateSettings(ctx, input)
}

func (s *Service) Submit(ctx context.Context, input CreateApplicationInput) (*Application, error) {
	if err := s.checkLimits(ctx); err != nil {
		s.log.Warn().Err(err).Str("email", input.Email).Msg("careers: application submission blocked by limit rules")
		return nil, err
	}

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
		subject := fmt.Sprintf("Application Received — %s Position at Travelis Rwanda Ltd", position)
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

func (s *Service) UpdateStatus(ctx context.Context, id, status string, notes *string, interviewAt *string, reviewerID string) (*Application, error) {
	app, err := s.repo.UpdateStatus(ctx, id, status, notes, interviewAt, reviewerID)
	if err != nil {
		return nil, err
	}

	// Send automated status update email asynchronously to candidate's personal email
	if app != nil {
		go func(candidateName, candidateEmail, position, newStatus string, interviewTime *time.Time) {
			formattedInterview := ""
			if interviewTime != nil {
				formattedInterview = interviewTime.Format("Monday, January 2, 2006 at 3:04 PM")
			}
			html, subject := email.BuildCareerStatusChangeEmail(candidateName, position, newStatus, formattedInterview)
			if html != "" {
				if err := email.SendEmail(context.Background(), candidateEmail, subject, html); err != nil {
					s.log.Warn().Err(err).Str("email", candidateEmail).Str("status", newStatus).Msg("careers: failed to send status change email")
				} else {
					s.log.Info().Str("email", candidateEmail).Str("status", newStatus).Msg("careers: candidate status change email sent successfully")
				}
			}
		}(app.FullName, app.Email, app.Position, app.ApplicationStatus, app.InterviewAt)
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
