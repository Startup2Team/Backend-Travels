CREATE TABLE IF NOT EXISTS career_applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    city VARCHAR(100) NOT NULL,
    work_right VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    institution VARCHAR(255),
    graduation_year VARCHAR(20),
    position VARCHAR(50) NOT NULL,
    technologies TEXT NOT NULL,
    project_url TEXT NOT NULL,
    project_body TEXT NOT NULL,
    github_url TEXT NOT NULL,
    linkedin_url TEXT,
    portfolio_url TEXT,
    cv_url TEXT,
    available_from_start BOOLEAN NOT NULL DEFAULT TRUE,
    heard_from VARCHAR(255),
    consent BOOLEAN NOT NULL DEFAULT TRUE,
    source VARCHAR(50) DEFAULT 'landing_careers',
    application_status VARCHAR(50) NOT NULL DEFAULT 'NEW',
    reviewer_notes TEXT,
    reviewed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_career_apps_email ON career_applications(email);
CREATE INDEX IF NOT EXISTS idx_career_apps_position ON career_applications(position);
CREATE INDEX IF NOT EXISTS idx_career_apps_status ON career_applications(application_status);
CREATE INDEX IF NOT EXISTS idx_career_apps_created_at ON career_applications(created_at DESC);
