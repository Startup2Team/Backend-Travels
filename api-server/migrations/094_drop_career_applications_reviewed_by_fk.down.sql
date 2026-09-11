ALTER TABLE career_applications ADD CONSTRAINT career_applications_reviewed_by_fkey FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL;
