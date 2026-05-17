# Próximos Pasos - Matching Service

## Fase de Desarrollo Siguiente

Este documento describe exactamente qué se debe implementar en las próximas 2-3 semanas.

---

## Sprint 1: Autenticación JWT (1-2 días)

### Objetivo
Implementar validación JWT en todos los endpoints (excepto `/health`)

### Tareas

1. **Crear middleware JWT** → `internal/middleware/jwt.go`
```go
func JWTMiddleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Extraer token del header Authorization: Bearer <token>
            // 2. Validar firma usando JWT_SECRET
            // 3. Extraer user_id del token
            // 4. Agregar user_id al contexto del request
            // 5. Llamar next handler
        })
    }
}
```

2. **Registrar middleware en router**
```go
r.Use(JWTMiddleware(cfg.JWT.Secret))
```

3. **Acceder user_id en handlers**
```go
userID := r.Context().Value("user_id").(string)
```

4. **Crear tests**
```go
TestJWTMiddlewareValidToken()
TestJWTMiddlewareInvalidToken()
TestJWTMiddlewareExpiredToken()
```

### Aceptación de Criterios
- [x] Todos los endpoints excepto `/health` requieren JWT
- [x] Tokens inválidos retornan 401
- [x] Tokens expirados retornan 401
- [x] User ID está disponible en handlers
- [x] Tests > 80% cobertura

---

## Sprint 2: Service Layer (3-4 días)

### Objetivo
Implementar la lógica de negocio del matching

### Tareas

1. **Crear repository interface** → `internal/repository/repository.go`
```go
type MatchRepository interface {
    // Match Requests
    CreateMatchRequest(ctx context.Context, req *domain.MatchRequest) error
    GetMatchRequest(ctx context.Context, id uuid.UUID) (*domain.MatchRequest, error)
    ListMatchRequests(ctx context.Context, filters *ListFilters) ([]*domain.MatchRequest, error)
    UpdateMatchRequest(ctx context.Context, req *domain.MatchRequest) error
    DeleteMatchRequest(ctx context.Context, id uuid.UUID) error
    
    // Tutors
    FindTutorsBySkills(ctx context.Context, skills []string) ([]*domain.TutorProfile, error)
    GetTutorProfile(ctx context.Context, userID uuid.UUID) (*domain.TutorProfile, error)
    
    // Match Results
    CreateMatchResult(ctx context.Context, result *domain.MatchResult) error
    UpdateMatchResult(ctx context.Context, result *domain.MatchResult) error
}
```

2. **Crear matching service** → `internal/service/matching.go`
```go
type MatchingService struct {
    repo MatchRepository
    logger *logrus.Logger
    notificationURL string
}

// FindCandidateTutors busca tutores que cumplen criterios
func (s *MatchingService) FindCandidateTutors(
    ctx context.Context,
    req *domain.MatchRequest,
    prefs *domain.MatchPreference,
) ([]*domain.TutorProfile, error)

// CalculateScore calcula el matching score
func (s *MatchingService) CalculateScore(
    tutor *domain.TutorProfile,
    req *domain.MatchRequest,
    prefs *domain.MatchPreference,
) (float64, error) {
    // Score = (skill × 0.4) + (avail × 0.3) + (rating × 0.2) + (price × 0.1)
    skillMatch := s.calculateSkillMatch(tutor, req.Subject, req.Level)
    availMatch := s.calculateAvailabilityMatch(tutor, prefs.PreferredSchedule)
    ratingScore := s.normalizeRating(tutor.Rating)
    priceMatch := s.calculatePriceMatch(tutor.HourlyRate, prefs.MaxPrice)
    
    score := (skillMatch * 0.4) + (availMatch * 0.3) + (ratingScore * 0.2) + (priceMatch * 0.1)
    return score, nil
}

// ProcessMatches procesa todas las solicitudes pendientes
func (s *MatchingService) ProcessMatches(ctx context.Context) error
```

3. **Implementar scoring helper functions**
```go
func (s *MatchingService) calculateSkillMatch(
    tutor *domain.TutorProfile,
    subject string,
    level string,
) float64 {
    // Retornar 0-100 basado en coincidencia de habilidades y nivel
}

func (s *MatchingService) calculateAvailabilityMatch(
    tutor *domain.TutorProfile,
    preferredSchedule string,
) float64 {
    // Retornar 0-100 basado en solapamiento de disponibilidad
}

func (s *MatchingService) calculatePriceMatch(
    hourlyRate int,
    maxPrice int,
) float64 {
    // Retornar 0-100 basado en compatibilidad de precio
}

func (s *MatchingService) normalizeRating(rating float64) float64 {
    // Convertir rating 0-5 a 0-100
    return (rating / 5.0) * 100.0
}
```

4. **Tests**
```go
TestCalculateScore_PerfectMatch()
TestCalculateScore_NoSkillMatch()
TestCalculateScore_BoundaryValues()
TestFindCandidateTutors()
TestProcessMatches()
```

### Aceptación de Criterios
- [x] Algoritmo de scoring implementado
- [x] Búsqueda de candidatos funciona
- [x] ProcessMatches obtiene resultados correctos
- [x] Tests > 80% cobertura

---

## Sprint 3: Repository Pattern (3-4 días)

### Objetivo
Implementar acceso a datos con queries optimizadas

### Tareas

1. **Crear PostgreSQL repository** → `internal/repository/postgres.go`
```go
type PostgresRepository struct {
    db *sql.DB
    logger *logrus.Logger
}

func NewPostgresRepository(db *sql.DB, logger *logrus.Logger) *PostgresRepository {
    return &PostgresRepository{db: db, logger: logger}
}

// Implementar todos los métodos del interface MatchRepository
```

2. **Implementar query methods**
```go
// Match Requests
func (r *PostgresRepository) CreateMatchRequest(
    ctx context.Context,
    req *domain.MatchRequest,
) error {
    // INSERT INTO match_requests VALUES (...)
    // INSERT INTO match_preferences VALUES (...) si existen
}

func (r *PostgresRepository) ListMatchRequests(
    ctx context.Context,
    filters *ListFilters,
) ([]*domain.MatchRequest, error) {
    // SELECT * FROM match_requests WHERE status = $1 AND subject ILIKE $2 ...
}

func (r *PostgresRepository) FindTutorsBySkills(
    ctx context.Context,
    skills []string,
) ([]*domain.TutorProfile, error) {
    // SELECT DISTINCT tp.* FROM tutor_profiles tp
    // JOIN tutor_skills ts ON tp.id = ts.tutor_id
    // WHERE ts.skill = ANY($1) AND tp.is_available = true
    // ORDER BY tp.rating DESC
}
```

3. **Usar parameterized queries** (prevenir SQL injection)
```go
// ✅ Correcto
rows, err := r.db.QueryContext(ctx, 
    "SELECT * FROM match_requests WHERE id = $1", 
    id,
)

// ❌ Incorrecto (NUNCA HACER)
query := fmt.Sprintf("SELECT * FROM match_requests WHERE id = '%s'", id)
```

4. **Tests con fixtures**
```go
// Usar transaction de test para rollback automático
func TestCreateMatchRequest(t *testing.T) {
    tx := setupTestDB(t)
    defer tx.Rollback()
    
    repo := NewPostgresRepository(tx, nil)
    
    req := &domain.MatchRequest{
        StudentID: uuid.New(),
        Subject: "Math",
        Level: "beginner",
        Status: "pending",
    }
    
    err := repo.CreateMatchRequest(context.Background(), req)
    assert.NoError(t, err)
    
    // Verificar que se creó
    retrieved, err := repo.GetMatchRequest(context.Background(), req.ID)
    assert.NoError(t, err)
    assert.Equal(t, req.Subject, retrieved.Subject)
}
```

### Aceptación de Criterios
- [x] Todos los métodos CRUD implementados
- [x] Queries optimizadas con índices
- [x] Parameterized queries para seguridad
- [x] Tests > 80% cobertura

---

## Sprint 4: Procesamiento Asincrónico (2 días)

### Objetivo
Implementar procesamiento en background sin bloquear requests

### Tareas

1. **Worker goroutine** → `internal/worker/matcher.go`
```go
type MatcherWorker struct {
    service *service.MatchingService
    logger *logrus.Logger
    done chan struct{}
}

func (w *MatcherWorker) Start(ctx context.Context) {
    go func() {
        ticker := time.NewTicker(10 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                if err := w.service.ProcessMatches(ctx); err != nil {
                    w.logger.Errorf("error processing matches: %v", err)
                }
            }
        }
    }()
}
```

2. **Event publishing** → `internal/service/events.go`
```go
func (s *MatchingService) PublishMatchEvent(
    ctx context.Context,
    event *domain.MatchEvent,
) error {
    // Convertir a JSON
    payload, err := json.Marshal(event)
    if err != nil {
        return err
    }
    
    // Enviar webhook POST al NOTIFICATION_WEBHOOK_URL
    req, _ := http.NewRequestWithContext(
        ctx,
        "POST",
        s.notificationURL,
        bytes.NewBuffer(payload),
    )
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 400 {
        return fmt.Errorf("webhook returned status %d", resp.StatusCode)
    }
    
    return nil
}
```

3. **Integración en main**
```go
// En main.go
worker := worker.NewMatcherWorker(matchingService, logger)

// Iniciar worker en background
go worker.Start(ctx)

// En graceful shutdown
cancel() // Detener worker
```

4. **Tests**
```go
TestWorkerStartsAndStops()
TestPublishMatchEvent()
TestPublishMatchEventRetry()
```

### Aceptación de Criterios
- [x] Worker procesa matches en background
- [x] No bloquea requests
- [x] Webhooks enviados correctamente
- [x] Tests pasan

---

## Sprint 5: Testing Exhaustivo (3-4 días)

### Objetivo
Alcanzar 80%+ cobertura con tests robustos

### Tareas

1. **Integration tests** → `tests/integration/matching_test.go`
```go
TestCreateMatchRequestFlow()        // end-to-end
TestProcessMatchesFlow()            // con BD real
TestRecommendationsFlow()           // recomendaciones
TestAcceptRejectFlow()              // aceptación/rechazo
```

2. **Load testing** → `tests/load/load_test.go`
```bash
# Generar 100 requests simultáneos
go-wrk -c 100 -n 10000 http://localhost:8084/health

# Verificar que P95 < 200ms, P99 < 200ms
```

3. **Benchmarks** → `main_test.go`
```go
func BenchmarkCalculateScore(b *testing.B) {
    // Setup
    for i := 0; i < b.N; i++ {
        service.CalculateScore(...)
    }
}
```

4. **Error scenarios**
```go
TestDatabaseConnectionError()
TestWebhookTimeout()
TestInvalidJWT()
TestRateLimitExceeded()
```

### Aceptación de Criterios
- [x] Coverage > 80%
- [x] Integration tests pasan
- [x] Load test verifica SLA
- [x] Error scenarios manejados

---

## Checklist de Implementación

### Sprint 1: JWT
- [ ] Middleware JWT creado
- [ ] Todos los endpoints requieren JWT (excepto /health)
- [ ] Tests JWT > 80%
- [ ] PR reviewed y merged

### Sprint 2: Service Layer
- [ ] Interface MatchRepository definido
- [ ] MatchingService implementado
- [ ] Algoritmo de scoring funciona
- [ ] Tests > 80%
- [ ] PR reviewed y merged

### Sprint 3: Repository
- [ ] PostgresRepository implementado
- [ ] Todos los métodos CRUD
- [ ] Queries optimizadas
- [ ] Tests > 80%
- [ ] PR reviewed y merged

### Sprint 4: Async
- [ ] Worker goroutine implementado
- [ ] Event publishing funciona
- [ ] Webhooks se envían correctamente
- [ ] Tests > 80%
- [ ] PR reviewed y merged

### Sprint 5: Testing
- [ ] Coverage > 80%
- [ ] Integration tests pasan
- [ ] Load tests verifican SLA
- [ ] Error scenarios manejados
- [ ] Final PR merged

---

## Comandos Útiles Durante Desarrollo

```bash
# Formatear código
make fmt

# Lint
make lint

# Tests
make test

# Coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Ejecutar servidor
make run-local

# Ver logs
make logs

# Base de datos
make db-shell
```

---

## Referencias

- [API_SPEC.md](API_SPEC.md) - Endpoints a implementar
- [ARCHITECTURE.md](ARCHITECTURE.md) - Patrones de diseño
- [examples.sh](examples.sh) - Ejemplos de uso
- [DEVELOPMENT.md](DEVELOPMENT.md) - Guía de desarrollo

---

## Timeline

```
Semana 1:
  Día 1-2: Sprint 1 (JWT)
  Día 3-4: Sprint 2 (Service Layer)
  Día 5: Review y tests

Semana 2:
  Día 1-2: Sprint 3 (Repository)
  Día 3-4: Sprint 4 (Async Processing)
  Día 5: Tests e integración

Semana 3:
  Día 1-3: Sprint 5 (Testing Exhaustivo)
  Día 4-5: Performance tuning y documentation
```

---

**Versión**: 1.0.0-MVP+Implementation
**Fecha**: 2024-01-15
**Status**: 🚀 Ready to Code
