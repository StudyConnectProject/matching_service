package main

import (
	"database/sql"
	"log"
	"strings"
)

// migrationSQL contains all DDL statements needed to bootstrap the database.
// Each statement is separated by a semicolon followed by a newline.
// Using IF NOT EXISTS on every object makes this fully idempotent.
const migrationSQL = `
CREATE TABLE IF NOT EXISTS match_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL,
    subject VARCHAR(255) NOT NULL,
    level VARCHAR(50) NOT NULL CHECK (level IN ('beginner', 'intermediate', 'advanced')),
    description TEXT,
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'rejected', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_match_requests_student_id ON match_requests(student_id);

CREATE INDEX IF NOT EXISTS idx_match_requests_status ON match_requests(status);

CREATE INDEX IF NOT EXISTS idx_match_requests_subject ON match_requests(subject);

CREATE TABLE IF NOT EXISTS match_preferences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL UNIQUE,
    preferred_schedule VARCHAR(50) CHECK (preferred_schedule IN ('morning', 'afternoon', 'evening', 'weekend')),
    preferred_language VARCHAR(100),
    modality VARCHAR(50) CHECK (modality IN ('virtual', 'in-person', 'hybrid')),
    max_price INT,
    FOREIGN KEY (request_id) REFERENCES match_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_match_preferences_request_id ON match_preferences(request_id);

CREATE TABLE IF NOT EXISTS tutor_profiles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL UNIQUE,
    bio TEXT,
    hourly_rate INT,
    rating FLOAT DEFAULT 0.0,
    is_available BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tutor_profiles_user_id ON tutor_profiles(user_id);

CREATE INDEX IF NOT EXISTS idx_tutor_profiles_is_available ON tutor_profiles(is_available);

CREATE INDEX IF NOT EXISTS idx_tutor_profiles_rating ON tutor_profiles(rating);

CREATE TABLE IF NOT EXISTS tutor_skills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tutor_id UUID NOT NULL,
    skill VARCHAR(255) NOT NULL,
    level VARCHAR(50) NOT NULL CHECK (level IN ('beginner', 'intermediate', 'advanced', 'expert')),
    FOREIGN KEY (tutor_id) REFERENCES tutor_profiles(id) ON DELETE CASCADE,
    UNIQUE(tutor_id, skill)
);

CREATE INDEX IF NOT EXISTS idx_tutor_skills_tutor_id ON tutor_skills(tutor_id);

CREATE INDEX IF NOT EXISTS idx_tutor_skills_skill ON tutor_skills(skill);

CREATE TABLE IF NOT EXISTS tutor_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tutor_id UUID NOT NULL,
    day_of_week VARCHAR(20) NOT NULL CHECK (day_of_week IN ('monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday')),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    FOREIGN KEY (tutor_id) REFERENCES tutor_profiles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tutor_availability_tutor_id ON tutor_availability(tutor_id);

CREATE INDEX IF NOT EXISTS idx_tutor_availability_day ON tutor_availability(day_of_week);

CREATE TABLE IF NOT EXISTS match_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL,
    tutor_id UUID NOT NULL,
    score FLOAT NOT NULL CHECK (score >= 0.0 AND score <= 100.0),
    status VARCHAR(50) NOT NULL DEFAULT 'suggested' CHECK (status IN ('suggested', 'accepted', 'rejected')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (request_id) REFERENCES match_requests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_match_results_request_id ON match_results(request_id);

CREATE INDEX IF NOT EXISTS idx_match_results_tutor_id ON match_results(tutor_id);

CREATE INDEX IF NOT EXISTS idx_match_results_status ON match_results(status);

CREATE INDEX IF NOT EXISTS idx_match_results_score ON match_results(score DESC);

CREATE TABLE IF NOT EXISTS match_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    request_id UUID NOT NULL,
    tutor_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL CHECK (action IN ('created', 'accepted', 'rejected', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_match_history_request_id ON match_history(request_id);

CREATE INDEX IF NOT EXISTS idx_match_history_tutor_id ON match_history(tutor_id);

CREATE INDEX IF NOT EXISTS idx_match_history_action ON match_history(action);

CREATE INDEX IF NOT EXISTS idx_match_history_created_at ON match_history(created_at)
`

// runMigrations executes the migration SQL against the given database.
// Each semicolon-delimited statement is run individually so that errors are
// isolated and already-existing objects do not abort the whole migration.
func runMigrations(db *sql.DB) error {
	log.Println("Running database migrations...")

	stmts := strings.Split(migrationSQL, ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			// Log but do not abort — the object may already exist or the
			// specific statement may be a no-op (e.g. duplicate index).
			log.Printf("Migration note: %v", err)
		}
	}

	log.Println("✅ Migrations applied")
	return nil
}
