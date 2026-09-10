CREATE TABLE IF NOT EXISTS career_settings (
    id INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    is_open BOOLEAN NOT NULL DEFAULT true,
    max_applications INT NOT NULL DEFAULT 0,
    hero_title TEXT NOT NULL DEFAULT 'Join Our Engineering Team & Build the Future of Mobility',
    hero_subtitle TEXT NOT NULL DEFAULT 'We are looking for passionate software engineers and interns to solve real-world mobility challenges across Rwanda.',
    open_at TIMESTAMPTZ,
    close_at TIMESTAMPTZ,
    closed_message TEXT NOT NULL DEFAULT 'Applications for our software engineering and internship programs are currently closed for this hiring cycle. Please check back for future openings!',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO career_settings (id, is_open, max_applications, hero_title, hero_subtitle, closed_message)
VALUES (
    1,
    true,
    0,
    'Join Our Engineering Team & Build the Future of Mobility',
    'We are looking for passionate software engineers and interns to solve real-world mobility challenges across Rwanda.',
    'Applications for our software engineering and internship programs are currently closed for this hiring cycle. Please check back for future openings!'
)
ON CONFLICT (id) DO NOTHING;
