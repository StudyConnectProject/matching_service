package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

type MatchingService struct {
	DB *sql.DB
}

type MatchRequest struct {
	ID          uuid.UUID `json:"id"`
	StudentID   uuid.UUID `json:"student_id"`
	Subject     string    `json:"subject"`
	Level       string    `json:"level"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateMatchRequestPayload struct {
	StudentID         string `json:"student_id"`
	Subject           string `json:"subject"`
	Level             string `json:"level"`
	Description       string `json:"description"`
	PreferredSchedule string `json:"preferred_schedule"`
	PreferredLanguage string `json:"preferred_language"`
	Modality          string `json:"modality"`
	MaxPrice          int    `json:"max_price"`
}

// OfferPayload is the body a tutor sends to offer themselves for a request
type OfferPayload struct {
	TutorID string  `json:"tutor_id"`
	Score   float64 `json:"score"`
}

// MatchView is a match result enriched with request info for the frontend
type MatchView struct {
	ID        uuid.UUID `json:"id"`
	RequestID uuid.UUID `json:"request_id"`
	TutorID   uuid.UUID `json:"tutor_id"`
	StudentID uuid.UUID `json:"student_id"`
	Subject   string    `json:"subject"`
	Level     string    `json:"level"`
	Score     float64   `json:"score"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Database string `json:"database"`
}

func init() {
	godotenv.Load()
}

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
CREATE INDEX IF NOT EXISTS idx_match_history_created_at ON match_history(created_at);
`

func runMigrations(db *sql.DB) error {
	stmts := strings.Split(migrationSQL, ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed on statement [%.60s...]: %w", stmt, err)
		}
	}
	return nil
}

func main() {
	cfg := LoadConfig()

	var err error
	db, err = sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)

	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	log.Println("✅ Database connected successfully")

	if err := runMigrations(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}
	log.Println("✅ Database schema ready")

	if err := runMigrations(db); err != nil {
		log.Printf("Warning: migration error: %v", err)
	}

	service := &MatchingService{DB: db}

	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health Check
	r.Get("/health", service.HealthCheck)

	// Matching API routes
	r.Route("/api/matching", func(r chi.Router) {
		r.Post("/request", service.CreateMatchRequest)
		r.Get("/", service.ListMatchRequests)
		r.Get("/{id}", service.GetMatchRequest)
		r.Put("/{id}", service.UpdateMatchRequest)
		r.Delete("/{id}", service.CancelMatchRequest)
		r.Patch("/{id}/status", service.UpdateMatchRequestStatus)
		r.Post("/{id}/accept", service.AcceptMatch)
		r.Post("/{id}/reject", service.RejectMatch)
		r.Post("/{id}/offer", service.OfferMatch)
		r.Get("/{userId}/matches", service.GetUserMatches)
		r.Get("/recommendations/{userId}", service.GetRecommendations)
		r.Post("/process", service.ProcessMatches)
	})

	addr := fmt.Sprintf(":%d", cfg.AppPort)
	log.Printf("🚀 Server starting on http://localhost:%d", cfg.AppPort)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// HealthCheck returns the health status of the service
func (s *MatchingService) HealthCheck(w http.ResponseWriter, r *http.Request) {
	dbStatus := "down"
	if err := s.DB.Ping(); err == nil {
		dbStatus = "up"
	}

	response := HealthResponse{
		Status:   "healthy",
		Message:  "Matching Service is running",
		Database: dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Simple JSON encoding without external library
	fmt.Fprintf(w, `{"status":"%s","message":"%s","database":"%s"}`, response.Status, response.Message, response.Database)
}

// CreateMatchRequest creates a new match request
func (s *MatchingService) CreateMatchRequest(w http.ResponseWriter, r *http.Request) {
	var payload CreateMatchRequestPayload

	// Parse JSON from request body
	decoder := NewJSONDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if payload.Subject == "" || payload.Level == "" {
		http.Error(w, `{"error":"Subject and Level are required"}`, http.StatusBadRequest)
		return
	}

	// Tie the request to the student. The frontend sends the authenticated
	// user's ID in the payload; fall back to a random ID if it's missing.
	studentID := uuid.New()
	if payload.StudentID != "" {
		if parsed, err := uuid.Parse(payload.StudentID); err == nil {
			studentID = parsed
		}
	}
	requestID := uuid.New()

	// Create match request
	_, err := s.DB.Exec(`
		INSERT INTO match_requests (id, student_id, subject, level, description, status)
		VALUES ($1, $2, $3, $4, $5, 'pending')
	`, requestID, studentID, payload.Subject, payload.Level, payload.Description)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create preferences if provided
	if payload.Modality != "" || payload.PreferredLanguage != "" {
		_, _ = s.DB.Exec(`
			INSERT INTO match_preferences (id, request_id, preferred_schedule, preferred_language, modality, max_price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New(), requestID, payload.PreferredSchedule, payload.PreferredLanguage, payload.Modality, payload.MaxPrice)
	}

	// Run the matching algorithm right away so the student immediately
	// sees suggested tutors instead of an empty list.
	matchCount := s.matchRequestWithTutors(requestID.String(), payload.Subject, payload.Level)

	status := "pending"
	if matchCount > 0 {
		status = "processing"
	}

	response := MatchRequest{
		ID:        requestID,
		StudentID: studentID,
		Subject:   payload.Subject,
		Level:     payload.Level,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":"%s","student_id":"%s","subject":"%s","level":"%s","status":"%s","matches_created":%d,"created_at":"%s"}`,
		response.ID, response.StudentID, response.Subject, response.Level, response.Status, matchCount, response.CreatedAt.Format(time.RFC3339))
}

// ListMatchRequests lists match requests with filters
func (s *MatchingService) ListMatchRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	subject := r.URL.Query().Get("subject")
	studentID := r.URL.Query().Get("student_id")

	query := "SELECT id, student_id, subject, level, description, status, created_at, updated_at FROM match_requests WHERE 1=1"
	args := []interface{}{}

	if status != "" {
		query += " AND status = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, status)
	}
	if subject != "" {
		query += " AND subject ILIKE $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, "%"+subject+"%")
	}
	if studentID != "" {
		query += " AND student_id = $" + fmt.Sprintf("%d", len(args)+1)
		args = append(args, studentID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"Failed to query requests"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	requests := []MatchRequest{}
	for rows.Next() {
		var req MatchRequest
		var description sql.NullString
		if err := rows.Scan(&req.ID, &req.StudentID, &req.Subject, &req.Level,
			&description, &req.Status, &req.CreatedAt, &req.UpdatedAt); err != nil {
			http.Error(w, `{"error":"Failed to read requests"}`, http.StatusInternalServerError)
			return
		}
		req.Description = description.String
		requests = append(requests, req)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"requests": requests})
}

// GetMatchRequest gets a specific match request
func (s *MatchingService) GetMatchRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var request MatchRequest
	var description sql.NullString
	err := s.DB.QueryRow(`
		SELECT id, student_id, subject, level, description, status, created_at, updated_at
		FROM match_requests WHERE id = $1
	`, id).Scan(&request.ID, &request.StudentID, &request.Subject, &request.Level,
		&description, &request.Status, &request.CreatedAt, &request.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Match request not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		}
		return
	}
	request.Description = description.String

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(request)
}

// UpdateMatchRequest updates the editable fields of a match request.
func (s *MatchingService) UpdateMatchRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var payload CreateMatchRequestPayload
	decoder := NewJSONDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}
	if payload.Subject == "" || payload.Level == "" {
		http.Error(w, `{"error":"Subject and Level are required"}`, http.StatusBadRequest)
		return
	}

	res, err := s.DB.Exec(`
		UPDATE match_requests
		SET subject = $1, level = $2, description = $3, updated_at = NOW()
		WHERE id = $4
	`, payload.Subject, payload.Level, payload.Description, id)
	if err != nil {
		http.Error(w, `{"error":"Failed to update request"}`, http.StatusInternalServerError)
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		http.Error(w, `{"error":"Match request not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Match request updated",
		"id":      id,
		"subject": payload.Subject,
		"level":   payload.Level,
	})
}

// CancelMatchRequest cancels a match request
func (s *MatchingService) CancelMatchRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.DB.Exec("UPDATE match_requests SET status = 'cancelled', updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		http.Error(w, `{"error":"Failed to cancel request"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Match request cancelled"}`)
}

// UpdateMatchRequestStatus updates the status of a match request
func (s *MatchingService) UpdateMatchRequestStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var payload map[string]interface{}
	decoder := NewJSONDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	status, ok := payload["status"].(string)
	if !ok || status == "" {
		http.Error(w, `{"error":"Status is required"}`, http.StatusBadRequest)
		return
	}

	_, err := s.DB.Exec("UPDATE match_requests SET status = $1, updated_at = NOW() WHERE id = $2", status, id)
	if err != nil {
		http.Error(w, `{"error":"Failed to update status"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Status updated"}`)
}

// AcceptMatch accepts a match: marks the result accepted, closes the
// request and records the action in the match history.
func (s *MatchingService) AcceptMatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var requestID, tutorID string
	err := s.DB.QueryRow("SELECT request_id, tutor_id FROM match_results WHERE id = $1", id).
		Scan(&requestID, &tutorID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Match not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	if _, err := s.DB.Exec("UPDATE match_results SET status = 'accepted' WHERE id = $1", id); err != nil {
		http.Error(w, `{"error":"Failed to accept match"}`, http.StatusInternalServerError)
		return
	}
	_, _ = s.DB.Exec("UPDATE match_requests SET status = 'completed', updated_at = NOW() WHERE id = $1", requestID)
	_, _ = s.DB.Exec(`INSERT INTO match_history (id, request_id, tutor_id, action)
		VALUES ($1, $2, $3, 'accepted')`, uuid.New(), requestID, tutorID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Match accepted","match_id":"%s","status":"accepted"}`, id)
}

// RejectMatch rejects a match and records the action in the match history.
func (s *MatchingService) RejectMatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var requestID, tutorID string
	err := s.DB.QueryRow("SELECT request_id, tutor_id FROM match_results WHERE id = $1", id).
		Scan(&requestID, &tutorID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Match not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	if _, err := s.DB.Exec("UPDATE match_results SET status = 'rejected' WHERE id = $1", id); err != nil {
		http.Error(w, `{"error":"Failed to reject match"}`, http.StatusInternalServerError)
		return
	}
	_, _ = s.DB.Exec(`INSERT INTO match_history (id, request_id, tutor_id, action)
		VALUES ($1, $2, $3, 'rejected')`, uuid.New(), requestID, tutorID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Match rejected","match_id":"%s","status":"rejected"}`, id)
}

// OfferMatch lets a tutor offer themselves for a student's match request.
func (s *MatchingService) OfferMatch(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "id")

	var payload OfferPayload
	decoder := NewJSONDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	tutorID, err := uuid.Parse(payload.TutorID)
	if err != nil {
		http.Error(w, `{"error":"A valid tutor_id is required"}`, http.StatusBadRequest)
		return
	}

	score := payload.Score
	if score <= 0 || score > 100 {
		score = 80
	}

	// The request must exist and still be open.
	var status string
	err = s.DB.QueryRow("SELECT status FROM match_requests WHERE id = $1", requestID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Match request not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		}
		return
	}
	if status == "cancelled" || status == "completed" {
		http.Error(w, `{"error":"This request is no longer open"}`, http.StatusConflict)
		return
	}

	// A tutor can only offer once per request.
	var existing int
	_ = s.DB.QueryRow("SELECT COUNT(*) FROM match_results WHERE request_id = $1 AND tutor_id = $2",
		requestID, tutorID).Scan(&existing)
	if existing > 0 {
		http.Error(w, `{"error":"You already offered for this request"}`, http.StatusConflict)
		return
	}

	resultID := uuid.New()
	_, err = s.DB.Exec(`
		INSERT INTO match_results (id, request_id, tutor_id, score, status)
		VALUES ($1, $2, $3, $4, 'suggested')
	`, resultID, requestID, tutorID, score)
	if err != nil {
		http.Error(w, `{"error":"Failed to create offer"}`, http.StatusInternalServerError)
		return
	}

	_, _ = s.DB.Exec("UPDATE match_requests SET status = 'processing', updated_at = NOW() WHERE id = $1 AND status = 'pending'", requestID)
	_, _ = s.DB.Exec(`INSERT INTO match_history (id, request_id, tutor_id, action)
		VALUES ($1, $2, $3, 'created')`, uuid.New(), requestID, tutorID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         resultID,
		"request_id": requestID,
		"tutor_id":   tutorID,
		"score":      score,
		"status":     "suggested",
	})
}

// GetUserMatches gets all matches for a user, either as student or as tutor.
func (s *MatchingService) GetUserMatches(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")
	status := r.URL.Query().Get("status")

	query := `
		SELECT mr.id, mr.request_id, mr.tutor_id, req.student_id, req.subject, req.level,
		       mr.score, mr.status, mr.created_at
		FROM match_results mr
		JOIN match_requests req ON req.id = mr.request_id
		WHERE (req.student_id = $1 OR mr.tutor_id = $1)`
	args := []interface{}{userID}
	if status != "" {
		query += " AND mr.status = $2"
		args = append(args, status)
	}
	query += " ORDER BY mr.created_at DESC"

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"Failed to query matches"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matches := []MatchView{}
	for rows.Next() {
		var m MatchView
		if err := rows.Scan(&m.ID, &m.RequestID, &m.TutorID, &m.StudentID,
			&m.Subject, &m.Level, &m.Score, &m.Status, &m.CreatedAt); err != nil {
			http.Error(w, `{"error":"Failed to read matches"}`, http.StatusInternalServerError)
			return
		}
		matches = append(matches, m)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"matches": matches})
}

// GetRecommendations gets top recommended tutors for a user
func (s *MatchingService) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"recommendations":[]}`)
}

// ProcessMatches runs the matching algorithm over every pending request.
func (s *MatchingService) ProcessMatches(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT id, subject, level FROM match_requests WHERE status = 'pending'")
	if err != nil {
		http.Error(w, `{"error":"Failed to load pending requests"}`, http.StatusInternalServerError)
		return
	}

	type pendingRequest struct{ id, subject, level string }
	var pending []pendingRequest
	for rows.Next() {
		var p pendingRequest
		if rows.Scan(&p.id, &p.subject, &p.level) == nil {
			pending = append(pending, p)
		}
	}
	rows.Close()

	totalMatches := 0
	for _, p := range pending {
		totalMatches += s.matchRequestWithTutors(p.id, p.subject, p.level)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"message":            "Processing completed",
		"requests_processed": len(pending),
		"matches_created":    totalMatches,
	})
}

// skillLevelWeight scores a tutor skill by mastery level.
var skillLevelWeight = map[string]float64{
	"expert":       1.0,
	"advanced":     0.8,
	"intermediate": 0.6,
	"beginner":     0.4,
}

// matchRequestWithTutors scores every available tutor against a request and
// stores the relevant ones as 'suggested' match results. It is idempotent:
// a tutor already matched to the request is skipped. Returns matches created.
func (s *MatchingService) matchRequestWithTutors(requestID, subject, level string) int {
	type tutorRow struct {
		profileID uuid.UUID
		userID    uuid.UUID
		bio       string
		rating    float64
	}

	rows, err := s.DB.Query("SELECT id, user_id, COALESCE(bio,''), COALESCE(rating,0) FROM tutor_profiles WHERE is_available = true")
	if err != nil {
		log.Printf("matching: failed to load tutors: %v", err)
		return 0
	}
	var tutors []tutorRow
	for rows.Next() {
		var t tutorRow
		if rows.Scan(&t.profileID, &t.userID, &t.bio, &t.rating) == nil {
			tutors = append(tutors, t)
		}
	}
	rows.Close()

	// Tutors already linked to this request must not be duplicated.
	matched := map[string]bool{}
	if exRows, err := s.DB.Query("SELECT tutor_id FROM match_results WHERE request_id = $1", requestID); err == nil {
		for exRows.Next() {
			var tid string
			if exRows.Scan(&tid) == nil {
				matched[tid] = true
			}
		}
		exRows.Close()
	}

	subjectLower := strings.ToLower(strings.TrimSpace(subject))

	type scoredTutor struct {
		userID   uuid.UUID
		score    float64
		relevant bool
	}
	var candidates []scoredTutor

	for _, t := range tutors {
		// Rating contributes up to 30 points.
		score := (t.rating / 5.0) * 30.0
		relevant := false

		// Each skill matching the subject adds points weighted by mastery.
		if skillRows, err := s.DB.Query("SELECT skill, level FROM tutor_skills WHERE tutor_id = $1", t.profileID); err == nil {
			for skillRows.Next() {
				var skill, lvl string
				if skillRows.Scan(&skill, &lvl) != nil {
					continue
				}
				sl := strings.ToLower(skill)
				if subjectLower != "" && (strings.Contains(sl, subjectLower) || strings.Contains(subjectLower, sl)) {
					score += skillLevelWeight[lvl] * 25.0
					relevant = true
				}
			}
			skillRows.Close()
		}

		// A subject mentioned in the tutor's bio is also a relevance signal.
		if subjectLower != "" && strings.Contains(strings.ToLower(t.bio), subjectLower) {
			score += 20.0
			relevant = true
		}

		if score < 1 {
			score = 1
		}
		if score > 100 {
			score = 100
		}
		candidates = append(candidates, scoredTutor{userID: t.userID, score: score, relevant: relevant})
	}

	// Prefer tutors relevant to the subject; otherwise fall back to the
	// 3 best-scored tutors so the student always gets suggestions.
	var chosen []scoredTutor
	for _, c := range candidates {
		if c.relevant {
			chosen = append(chosen, c)
		}
	}
	if len(chosen) == 0 {
		sort.Slice(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })
		if len(candidates) > 3 {
			candidates = candidates[:3]
		}
		chosen = candidates
	}

	created := 0
	for _, c := range chosen {
		if matched[c.userID.String()] {
			continue
		}
		if _, err := s.DB.Exec(`
			INSERT INTO match_results (id, request_id, tutor_id, score, status)
			VALUES ($1, $2, $3, $4, 'suggested')`,
			uuid.New(), requestID, c.userID, c.score); err != nil {
			log.Printf("matching: failed to insert result: %v", err)
			continue
		}
		_, _ = s.DB.Exec(`INSERT INTO match_history (id, request_id, tutor_id, action)
			VALUES ($1, $2, $3, 'created')`, uuid.New(), requestID, c.userID)
		created++
	}

	if created > 0 {
		_, _ = s.DB.Exec("UPDATE match_requests SET status = 'processing', updated_at = NOW() WHERE id = $1 AND status = 'pending'", requestID)
	}
	return created
}

// JSONDecoder is a simple JSON decoder wrapper
type JSONDecoder struct {
	decoder *json.Decoder
}

func NewJSONDecoder(r io.Reader) *JSONDecoder {
	return &JSONDecoder{
		decoder: json.NewDecoder(r),
	}
}

func (d *JSONDecoder) Decode(v interface{}) error {
	return d.decoder.Decode(v)
}
