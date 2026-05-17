// Domain models for the matching service
package domain

import (
	"time"

	"github.com/google/uuid"
)

// MatchRequest represents a student's request to find a tutor
type MatchRequest struct {
	ID          uuid.UUID
	StudentID   uuid.UUID
	Subject     string
	Level       string // beginner | intermediate | advanced
	Description string
	Status      string // pending | processing | completed | rejected | cancelled
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// MatchPreference represents additional preferences for a match request
type MatchPreference struct {
	ID                uuid.UUID
	RequestID         uuid.UUID
	PreferredSchedule string // morning | afternoon | evening | weekend
	PreferredLanguage string
	Modality          string // virtual | in-person | hybrid
	MaxPrice          int
}

// TutorProfile represents a tutor's profile
type TutorProfile struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Bio         string
	HourlyRate  int
	Rating      float64
	IsAvailable bool
	UpdatedAt   time.Time
}

// TutorSkill represents a skill that a tutor has
type TutorSkill struct {
	ID      uuid.UUID
	TutorID uuid.UUID
	Skill   string
	Level   string // beginner | intermediate | advanced | expert
}

// TutorAvailability represents when a tutor is available
type TutorAvailability struct {
	ID        uuid.UUID
	TutorID   uuid.UUID
	DayOfWeek string // monday | tuesday | ... | sunday
	StartTime string // HH:MM format
	EndTime   string // HH:MM format
}

// MatchResult represents the result of the matching algorithm
type MatchResult struct {
	ID        uuid.UUID
	RequestID uuid.UUID
	TutorID   uuid.UUID
	Score     float64 // 0.0 to 100.0
	Status    string  // suggested | accepted | rejected
	CreatedAt time.Time
}

// MatchHistory represents the history of actions on a match
type MatchHistory struct {
	ID        uuid.UUID
	RequestID uuid.UUID
	TutorID   uuid.UUID
	Action    string    // created | accepted | rejected | cancelled
	CreatedAt time.Time
}

// MatchingRequest is the request payload for creating a match request
type MatchingRequestPayload struct {
	Subject           string `json:"subject" binding:"required"`
	Level             string `json:"level" binding:"required,oneof=beginner intermediate advanced"`
	Description       string `json:"description"`
	PreferredSchedule string `json:"preferred_schedule" binding:"oneof=morning afternoon evening weekend''"`
	PreferredLanguage string `json:"preferred_language"`
	Modality          string `json:"modality" binding:"oneof=virtual in-person hybrid''"`
	MaxPrice          int    `json:"max_price"`
}

// MatchingResponse is the response payload for match operations
type MatchingResponse struct {
	ID        uuid.UUID `json:"id"`
	StudentID uuid.UUID `json:"student_id"`
	Subject   string    `json:"subject"`
	Level     string    `json:"level"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RecommendationResponse represents a tutor recommendation
type RecommendationResponse struct {
	TutorID    uuid.UUID `json:"tutor_id"`
	Name       string    `json:"name"`
	Bio        string    `json:"bio"`
	HourlyRate int       `json:"hourly_rate"`
	Rating     float64   `json:"rating"`
	Score      float64   `json:"score"`
}

// HealthResponse is the response for health checks
type HealthResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Version  string `json:"version"`
	Database string `json:"database"`
}
