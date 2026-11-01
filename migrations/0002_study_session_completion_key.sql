ALTER TABLE study_sessions ADD COLUMN completion_key TEXT;
CREATE UNIQUE INDEX idx_study_sessions_completion_key
  ON study_sessions(completion_key);
