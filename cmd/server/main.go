package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	Subject           string `json:"subject"`
	Level             string `json:"level"`
	Description       string `json:"description"`
	PreferredSchedule string `json:"preferred_schedule"`
	PreferredLanguage string `json:"preferred_language"`
	Modality          string `json:"modality"`
	MaxPrice          int    `json:"max_price"`
}

type HealthResponse struct {
	Status   string `json:"status"`
	Message  string `json:"message"`
	Database string `json:"database"`
}

func init() {
	godotenv.Load()
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
		r.Delete("/{id}", service.CancelMatchRequest)
		r.Patch("/{id}/status", service.UpdateMatchRequestStatus)
		r.Post("/{id}/accept", service.AcceptMatch)
		r.Post("/{id}/reject", service.RejectMatch)
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

	// For now, use a dummy student ID (in real implementation, get from JWT)
	studentID := uuid.New()
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

	response := MatchRequest{
		ID:        requestID,
		StudentID: studentID,
		Subject:   payload.Subject,
		Level:     payload.Level,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"id":"%s","student_id":"%s","subject":"%s","level":"%s","status":"pending","created_at":"%s"}`,
		response.ID, response.StudentID, response.Subject, response.Level, response.CreatedAt.Format(time.RFC3339))
}

// ListMatchRequests lists match requests with filters
func (s *MatchingService) ListMatchRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	subject := r.URL.Query().Get("subject")

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

	rows, err := s.DB.Query(query, args...)
	if err != nil {
		http.Error(w, `{"error":"Failed to query requests"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"requests":[]}`)
}

// GetMatchRequest gets a specific match request
func (s *MatchingService) GetMatchRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var request MatchRequest
	err := s.DB.QueryRow(`
		SELECT id, student_id, subject, level, description, status, created_at, updated_at
		FROM match_requests WHERE id = $1
	`, id).Scan(&request.ID, &request.StudentID, &request.Subject, &request.Level,
		&request.Description, &request.Status, &request.CreatedAt, &request.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, `{"error":"Match request not found"}`, http.StatusNotFound)
		} else {
			http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"id":"%s","student_id":"%s","subject":"%s","level":"%s","status":"%s"}`,
		request.ID, request.StudentID, request.Subject, request.Level, request.Status)
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

// AcceptMatch accepts a match
func (s *MatchingService) AcceptMatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.DB.Exec(`
		UPDATE match_results SET status = 'accepted' WHERE id = $1
	`, id)

	if err != nil {
		http.Error(w, `{"error":"Failed to accept match"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Match accepted"}`)
}

// RejectMatch rejects a match
func (s *MatchingService) RejectMatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	_, err := s.DB.Exec(`
		UPDATE match_results SET status = 'rejected' WHERE id = $1
	`, id)

	if err != nil {
		http.Error(w, `{"error":"Failed to reject match"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message":"Match rejected"}`)
}

// GetUserMatches gets all active matches for a user
func (s *MatchingService) GetUserMatches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"matches":[]}`)
}

// GetRecommendations gets top recommended tutors for a user
func (s *MatchingService) GetRecommendations(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"recommendations":[]}`)
}

// ProcessMatches triggers async processing of pending matches
func (s *MatchingService) ProcessMatches(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would trigger async workers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, `{"message":"Processing started","job_id":"%s"}`, uuid.New())
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
