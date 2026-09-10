-- Ensure hero_title and hero_subtitle exist on career_settings table
ALTER TABLE career_settings ADD COLUMN IF NOT EXISTS hero_title TEXT NOT NULL DEFAULT 'Join Our Engineering Team & Build the Future of Mobility';
ALTER TABLE career_settings ADD COLUMN IF NOT EXISTS hero_subtitle TEXT NOT NULL DEFAULT 'We are looking for passionate software engineers and interns to solve real-world mobility challenges across Rwanda.';
