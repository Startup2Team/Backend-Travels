package careers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	apperrors "github.com/workspace/ride-platform/pkg/errors"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, input CreateApplicationInput) (*Application, error) {
	if input.Source == "" {
		input.Source = "landing_careers"
	}

	query := `
		INSERT INTO career_applications (
			full_name, email, phone, city, work_right, status,
			institution, graduation_year, position, technologies,
			project_url, project_body, github_url, linkedin_url, portfolio_url,
			cv_url, available_from_start, heard_from, consent, source
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20
		)
		RETURNING
			id, full_name, email, phone, city, work_right, status,
			institution, graduation_year, position, technologies,
			project_url, project_body, github_url, linkedin_url, portfolio_url,
			cv_url, available_from_start, heard_from, consent, source,
			application_status, reviewer_notes, reviewed_by, reviewed_at,
			created_at, updated_at
	`

	app := &Application{}
	err := r.db.QueryRow(ctx, query,
		input.FullName, input.Email, input.Phone, input.City, input.WorkRight, input.Status,
		input.Institution, input.GraduationYear, input.Position, input.Technologies,
		input.ProjectURL, input.ProjectBody, input.GithubURL, input.LinkedinURL, input.PortfolioURL,
		input.CVURL, input.AvailableFromStart, input.HeardFrom, input.Consent, input.Source,
	).Scan(
		&app.ID, &app.FullName, &app.Email, &app.Phone, &app.City, &app.WorkRight, &app.Status,
		&app.Institution, &app.GraduationYear, &app.Position, &app.Technologies,
		&app.ProjectURL, &app.ProjectBody, &app.GithubURL, &app.LinkedinURL, &app.PortfolioURL,
		&app.CVURL, &app.AvailableFromStart, &app.HeardFrom, &app.Consent, &app.Source,
		&app.ApplicationStatus, &app.ReviewerNotes, &app.ReviewedBy, &app.ReviewedAt,
		&app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("career repository create: %w", err)
	}
	return app, nil
}

func (r *Repository) List(ctx context.Context, filter ListFilter) ([]*Application, int, error) {
	where := "WHERE 1=1"
	args := []interface{}{}
	argIdx := 1

	if filter.Position != "" {
		where += fmt.Sprintf(" AND position = $%d", argIdx)
		args = append(args, filter.Position)
		argIdx++
	}
	if filter.ApplicationStatus != "" {
		where += fmt.Sprintf(" AND application_status = $%d", argIdx)
		args = append(args, filter.ApplicationStatus)
		argIdx++
	}
	if filter.Search != "" {
		where += fmt.Sprintf(" AND (full_name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+filter.Search+"%")
		argIdx++
	}

	countQuery := "SELECT COUNT(*) FROM career_applications " + where
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("career repository count: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT
			id, full_name, email, phone, city, work_right, status,
			institution, graduation_year, position, technologies,
			project_url, project_body, github_url, linkedin_url, portfolio_url,
			cv_url, available_from_start, heard_from, consent, source,
			application_status, reviewer_notes, reviewed_by, reviewed_at,
			created_at, updated_at
		FROM career_applications
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("career repository list: %w", err)
	}
	defer rows.Close()

	var apps []*Application
	for rows.Next() {
		app := &Application{}
		if err := rows.Scan(
			&app.ID, &app.FullName, &app.Email, &app.Phone, &app.City, &app.WorkRight, &app.Status,
			&app.Institution, &app.GraduationYear, &app.Position, &app.Technologies,
			&app.ProjectURL, &app.ProjectBody, &app.GithubURL, &app.LinkedinURL, &app.PortfolioURL,
			&app.CVURL, &app.AvailableFromStart, &app.HeardFrom, &app.Consent, &app.Source,
			&app.ApplicationStatus, &app.ReviewerNotes, &app.ReviewedBy, &app.ReviewedAt,
			&app.CreatedAt, &app.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("career repository scan: %w", err)
		}
		apps = append(apps, app)
	}

	return apps, total, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*Application, error) {
	query := `
		SELECT
			id, full_name, email, phone, city, work_right, status,
			institution, graduation_year, position, technologies,
			project_url, project_body, github_url, linkedin_url, portfolio_url,
			cv_url, available_from_start, heard_from, consent, source,
			application_status, reviewer_notes, reviewed_by, reviewed_at,
			created_at, updated_at
		FROM career_applications
		WHERE id = $1
	`
	app := &Application{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.FullName, &app.Email, &app.Phone, &app.City, &app.WorkRight, &app.Status,
		&app.Institution, &app.GraduationYear, &app.Position, &app.Technologies,
		&app.ProjectURL, &app.ProjectBody, &app.GithubURL, &app.LinkedinURL, &app.PortfolioURL,
		&app.CVURL, &app.AvailableFromStart, &app.HeardFrom, &app.Consent, &app.Source,
		&app.ApplicationStatus, &app.ReviewerNotes, &app.ReviewedBy, &app.ReviewedAt,
		&app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("career repository get by id: %w", err)
	}
	return app, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string, notes *string, reviewerID string) (*Application, error) {
	query := `
		UPDATE career_applications
		SET application_status = $1,
		    reviewer_notes = COALESCE($2, reviewer_notes),
		    reviewed_by = $3,
		    reviewed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $4
		RETURNING
			id, full_name, email, phone, city, work_right, status,
			institution, graduation_year, position, technologies,
			project_url, project_body, github_url, linkedin_url, portfolio_url,
			cv_url, available_from_start, heard_from, consent, source,
			application_status, reviewer_notes, reviewed_by, reviewed_at,
			created_at, updated_at
	`
	app := &Application{}
	err := r.db.QueryRow(ctx, query, status, notes, reviewerID, id).Scan(
		&app.ID, &app.FullName, &app.Email, &app.Phone, &app.City, &app.WorkRight, &app.Status,
		&app.Institution, &app.GraduationYear, &app.Position, &app.Technologies,
		&app.ProjectURL, &app.ProjectBody, &app.GithubURL, &app.LinkedinURL, &app.PortfolioURL,
		&app.CVURL, &app.AvailableFromStart, &app.HeardFrom, &app.Consent, &app.Source,
		&app.ApplicationStatus, &app.ReviewerNotes, &app.ReviewedBy, &app.ReviewedAt,
		&app.CreatedAt, &app.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("career repository update status: %w", err)
	}
	return app, nil
}

func (r *Repository) CountTotal(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM career_applications`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("career repository count total: %w", err)
	}
	return count, nil
}

func (r *Repository) GetSettings(ctx context.Context) (*CareerSettings, error) {
	query := `
		SELECT is_open, max_applications, open_at, close_at, closed_message, updated_at
		FROM career_settings
		WHERE id = 1
	`
	s := &CareerSettings{}
	err := r.db.QueryRow(ctx, query).Scan(
		&s.IsOpen, &s.MaxApplications, &s.OpenAt, &s.CloseAt, &s.ClosedMessage, &s.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			s.IsOpen = true
			s.MaxApplications = 0
			s.ClosedMessage = "Applications for this recruitment cycle are currently closed."
			s.UpdatedAt = time.Now()
		} else {
			return nil, fmt.Errorf("career repository get settings: %w", err)
		}
	}

	count, _ := r.CountTotal(ctx)
	s.TotalSubmitted = count

	return s, nil
}

func (r *Repository) UpdateSettings(ctx context.Context, input UpdateSettingsInput) (*CareerSettings, error) {
	current, err := r.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	if input.IsOpen != nil {
		current.IsOpen = *input.IsOpen
	}
	if input.MaxApplications != nil {
		current.MaxApplications = *input.MaxApplications
	}
	if input.ClosedMessage != nil && strings.TrimSpace(*input.ClosedMessage) != "" {
		current.ClosedMessage = strings.TrimSpace(*input.ClosedMessage)
	}

	if input.OpenAt != nil {
		if *input.OpenAt == "" {
			current.OpenAt = nil
		} else if t, err := parseFlexTime(*input.OpenAt); err == nil {
			current.OpenAt = &t
		}
	}

	if input.CloseAt != nil {
		if *input.CloseAt == "" {
			current.CloseAt = nil
		} else if t, err := parseFlexTime(*input.CloseAt); err == nil {
			current.CloseAt = &t
		}
	}

	query := `
		INSERT INTO career_settings (id, is_open, max_applications, open_at, close_at, closed_message, updated_at)
		VALUES (1, $1, $2, $3, $4, $5, NOW())
		ON CONFLICT (id) DO UPDATE SET
			is_open = EXCLUDED.is_open,
			max_applications = EXCLUDED.max_applications,
			open_at = EXCLUDED.open_at,
			close_at = EXCLUDED.close_at,
			closed_message = EXCLUDED.closed_message,
			updated_at = NOW()
		RETURNING is_open, max_applications, open_at, close_at, closed_message, updated_at
	`

	s := &CareerSettings{}
	err = r.db.QueryRow(ctx, query,
		current.IsOpen, current.MaxApplications, current.OpenAt, current.CloseAt, current.ClosedMessage,
	).Scan(
		&s.IsOpen, &s.MaxApplications, &s.OpenAt, &s.CloseAt, &s.ClosedMessage, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("career repository update settings: %w", err)
	}

	count, _ := r.CountTotal(ctx)
	s.TotalSubmitted = count

	return s, nil
}
