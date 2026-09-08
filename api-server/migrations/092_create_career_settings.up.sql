CREATE TABLE IF NOT EXISTS career_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    is_open BOOLEAN NOT NULL DEFAULT true,
    max_applications INT NOT NULL DEFAULT 0,
    open_at TIMESTAMPTZ,
    close_at TIMESTAMPTZ,
    closed_message TEXT NOT NULL DEFAULT 'Applications for this recruitment cycle are currently closed.',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO career_settings (id, is_open, max_applications, closed_message)
VALUES (1, true, 0, 'Applications for this recruitment cycle are currently closed.')
ON CONFLICT (id) DO NOTHING;
