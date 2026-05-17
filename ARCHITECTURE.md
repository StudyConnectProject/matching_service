# Arquitectura - Matching Service

## Visión General

El Matching Service es un microservicio que empareja estudiantes con tutores usando un algoritmo de scoring avanzado. Opera de forma asincrónica para mantener baja latencia en las APIs.

## Diagrama de Arquitectura

```
┌─────────────────────────────────────────────────────────────────┐
│                     API Gateway (Spring Cloud)                   │
│                    Enruta /api/matching/**                       │
└─────────────────────────────────────────────────────────────────┘
                                ↓
┌─────────────────────────────────────────────────────────────────┐
│                   Matching Service (Go)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌──────────────────┐    ┌──────────────────┐                  │
│  │  REST Handlers   │    │   Middleware     │                  │
│  ├──────────────────┤    ├──────────────────┤                  │
│  │ POST   /request  │    │ JWT Auth         │                  │
│  │ GET    /         │    │ CORS             │                  │
│  │ POST   /process  │    │ Request ID       │                  │
│  │ POST   /accept   │    │ Logging          │                  │
│  │ GET    /recommend│    │ Error Handling   │                  │
│  └────────┬─────────┘    └────────┬─────────┘                  │
│           │                       │                             │
│  ┌────────▼────────────────────────▼────────┐                  │
│  │        Service Layer (Business Logic)     │                  │
│  ├──────────────────────────────────────────┤                  │
│  │ - Create Match Request                   │                  │
│  │ - Search Candidates                      │                  │
│  │ - Calculate Scores                       │                  │
│  │ - Handle Acceptance/Rejection            │                  │
│  │ - Notification Publishing                │                  │
│  └────────┬─────────────────────────────────┘                  │
│           │                                                     │
│  ┌────────▼─────────────────────────┐                          │
│  │   Repository Layer (Data Access) │                          │
│  ├──────────────────────────────────┤                          │
│  │ - CRUD Operations                │                          │
│  │ - Query Optimization             │                          │
│  │ - Transaction Management         │                          │
│  └────────┬────────────────────────┬┘                          │
│           │                        │                           │
└───────────┼────────────────────────┼──────────────────────────┘
            │                        │
    ┌───────▼──────┐        ┌────────▼──────┐
    │ PostgreSQL   │        │  Event Queue  │
    │              │        │  (async jobs) │
    └──────────────┘        └───────┬───────┘
                                    │
                                    ↓
                        ┌─────────────────────┐
                        │ Notification Service│
                        └─────────────────────┘
```

## Flujo de Datos: Crear Solicitud

```
1. POST /api/matching/request
   │
   ├─ Validar JWT
   ├─ Validar payload
   ├─ Crear MatchRequest (status=pending)
   ├─ Crear MatchPreference
   └─ Return: 201 Created

   Latencia esperada: < 100ms
```

## Flujo de Datos: Procesar Matches

```
1. POST /api/matching/process (webhbok o timer)
   │
   ├─ Obtener solicitudes pendientes
   │
   ├─ Para cada solicitud:
   │  ├─ Buscar tutores candidatos
   │  │  └─ Filtrar por habilidades
   │  │  └─ Filtrar por disponibilidad
   │  │  └─ Filtrar por precio
   │  │
   │  ├─ Calcular score para cada candidato
   │  │  Score = (skill × 0.4) + (avail × 0.3) + (rating × 0.2) + (price × 0.1)
   │  │
   │  ├─ Crear MatchResults (top N candidatos)
   │  │
   │  ├─ Actualizar estado de MatchRequest (processing → completed)
   │  │
   │  └─ Publicar evento (webhook)
   │     └─ POST NOTIFICATION_WEBHOOK_URL
   │
   └─ Return: 202 Accepted

   Latencia esperada: < 5s
   Puede ejecutarse en background
```

## Componentes

### 1. Handlers (HTTP REST API)

**Responsabilidades:**
- Recibir requests HTTP
- Validar entrada
- Autenticar con JWT
- Delegar a service layer
- Retornar respuestas JSON
- Manejo de errores

**Archivos:** `main.go` (métodos handler)

**Ejemplos:**
- `CreateMatchRequest()`
- `ListMatchRequests()`
- `ProcessMatches()`
- `AcceptMatch()`

### 2. Service Layer (Lógica de Negocio)

**Responsabilidades:**
- Implementar lógica de matching
- Calcular scores
- Validar reglas de negocio
- Orquestar operaciones de data
- Publicar eventos

**Archivos:** `service.go` (crear, añadir métodos)

**Funciones principales:**
- `findTutorCandidates()` - Busca tutores que cumplan criterios
- `calculateMatchScore()` - Computa el score usando el algoritmo
- `publishMatchEvent()` - Envía webhook a notification service

### 3. Repository Layer (Acceso a Datos)

**Responsabilidades:**
- Operaciones CRUD
- Queries optimizadas
- Gestión de transacciones
- Manejo de índices

**Archivos:** `repository.go` (crear, añadir métodos)

**Métodos:**
- `GetMatchRequestByID(id UUID) (*MatchRequest, error)`
- `CreateMatchRequest(req *MatchRequest) error`
- `SearchTutorsBySkills(skills []string) ([]*TutorProfile, error)`
- `CreateMatchResult(result *MatchResult) error`

### 4. Models/Domain

**Responsabilidades:**
- Definir estructuras de datos
- Validaciones básicas
- Métodos de utilidad

**Archivos:** `models.go`

**Estructuras:**
- `MatchRequest`
- `MatchPreference`
- `TutorProfile`
- `TutorSkill`
- `MatchResult`

## Patrón de Diseño: Repository

```go
// Repository Interface
type MatchRepository interface {
    GetByID(id UUID) (*MatchRequest, error)
    Create(req *MatchRequest) error
    Update(req *MatchRequest) error
    Delete(id UUID) error
}

// Implementation
type PostgresRepository struct {
    db *sql.DB
}

func (r *PostgresRepository) GetByID(id UUID) (*MatchRequest, error) {
    // implementar
}
```

## Patrones de Concurrencia

### Goroutines para Async Processing

```go
// En ProcessMatches handler
func (s *MatchingService) ProcessMatches(w http.ResponseWriter, r *http.Request) {
    jobID := uuid.New()
    
    go func() {
        // Long-running operation en background
        s.processMatchesAsync(jobID)
    }()
    
    w.WriteHeader(http.StatusAccepted)
    // Return inmediatamente
}

func (s *MatchingService) processMatchesAsync(jobID UUID) {
    // Buscar solicitudes pendientes
    // Calcular scores
    // Crear resultados
    // Publicar eventos
}
```

## Diseño de BD: Índices

```sql
-- Foreign keys
CREATE INDEX idx_match_preferences_request_id ON match_preferences(request_id);
CREATE INDEX idx_match_results_request_id ON match_results(request_id);

-- Búsquedas comunes
CREATE INDEX idx_match_requests_status ON match_requests(status);
CREATE INDEX idx_match_requests_student_id ON match_requests(student_id);
CREATE INDEX idx_match_results_score ON match_results(score DESC);

-- Disponibilidad
CREATE INDEX idx_tutor_availability_day ON tutor_availability(day_of_week);

-- Habilidades
CREATE INDEX idx_tutor_skills_skill ON tutor_skills(skill);
CREATE INDEX idx_tutor_skills_tutor_id ON tutor_skills(tutor_id);
```

## Escalabilidad

### Horizontal Scaling

```
Múltiples instancias de Matching Service
                    ↓
            Load Balancer
                    ↓
        ┌──────────┼──────────┐
        ↓          ↓          ↓
    Service 1  Service 2  Service 3
        │          │          │
        └──────────┼──────────┘
                   ↓
         PostgreSQL (shared)
```

### Optimizaciones

1. **Connection Pooling**
   - Reutilizar conexiones a BD
   - MaxOpenConns = 10
   - MaxIdleConns = 2

2. **Caché (Future)**
   - Redis para datos de tutores
   - In-memory para configuración

3. **Batch Processing**
   - Procesar múltiples matches juntos
   - Evitar queries N+1

## Security

### Autenticación y Autorización

```
Request
   ↓
Middleware JWT
   ├─ Validar signature
   ├─ Verificar expiration
   ├─ Extraer user_id
   └─ Permitir solo acceso a recursos propios
```

### Validación de Input

```go
// Whitelist de valores permitidos
validLevels := map[string]bool{
    "beginner": true,
    "intermediate": true,
    "advanced": true,
}

if !validLevels[payload.Level] {
    return errors.New("invalid level")
}
```

## Monitoreo

### Métricas Clave

- Latencia de requests (< 200ms 95th percentile)
- Latencia de matching (< 5s)
- Tasa de error
- Matches aceptados / rechazados
- Disponibilidad de BD

### Logging

```go
log.Infof("Match request created: id=%s student_id=%s", requestID, studentID)
log.Warnf("No tutors found for request: id=%s", requestID)
log.Errorf("Database error: %v", err)
```

## Versionado de API

```
/api/v1/matching/request  (current)
/api/v2/matching/request  (future)
```

## Deployment

### Desarrollo
```bash
docker-compose up
```

### Producción
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up
```

Incluye:
- TLS/HTTPS
- Rate limiting
- Logging centralizado
- Métricas y alertas

## Futuros Mejoras

1. **ML-based Matching**
   - Ajustar pesos del algoritmo automáticamente
   - Aprender de aceptaciones/rechazos

2. **Real-time Notifications**
   - WebSockets en lugar de webhooks
   - Push notifications

3. **Advanced Filtering**
   - Búsqueda full-text
   - Filtros por experiencia
   - Filtros por certificaciones

4. **Analytics Dashboard**
   - Estadísticas de matching
   - Success rate por tutor
   - Tendencias de demanda

5. **Quality Assurance**
   - Verificación de identidad
   - Certificaciones
   - Reviews y ratings
